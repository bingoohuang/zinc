package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	v1 "github.com/prabhatsharma/zinc/pkg/meta/v1"
	"golang.org/x/crypto/argon2"

	"github.com/blugelabs/bluge"
	"github.com/prabhatsharma/zinc/pkg/core"
)

func GetUser(userId string) (bool, ZincUser, error) {
	userExists := false
	var user ZincUser

	query := bluge.NewTermQuery(userId)
	searchRequest := bluge.NewTopNSearch(1, query)
	usersIndex := core.ZincSystemIndexList[core.SystemIndexUsers]
	reader, _ := usersIndex.Writer.Reader()
	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Printf("error executing search: %v", err)
	}

	for next, err := dmi.Next(); err == nil && next != nil; {
		userExists = true

		err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case "_id":
				user.ID = string(value)
			case "name":
				user.Name = string(value)
			case "salt":
				user.Salt = string(value)
			case "password":
				user.Password = string(value)
			case "role":
				user.Role = string(value)
			case "created_at":
				user.CreatedAt, _ = bluge.DecodeDateTime(value)
			case "@timestamp":
				user.Timestamp, _ = bluge.DecodeDateTime(value)
			}
			return true
		})
		if err != nil {
			log.Printf("error accessing stored fields: %v", err)
			return userExists, user, err
		} else {
			return userExists, user, nil
		}
	}

	return false, user, nil
}

func GetAllUsersWorker() (v1.SearchResponse, error) {
	usersIndex := core.ZincSystemIndexList[core.SystemIndexUsers]
	var Hits []v1.Hit

	query := bluge.NewMatchAllQuery()
	searchRequest := bluge.NewTopNSearch(1000, query).WithStandardAggregations()
	reader, _ := usersIndex.Writer.Reader()
	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Printf("error executing search: %v", err)
	}

	for next, err := dmi.Next(); err == nil && next != nil; {
		var user SimpleUser
		if err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case "_id":
				user.ID = string(value)
			case "name":
				user.Name = string(value)
			case "role":
				user.Role = string(value)
			case "created_at":
				user.CreatedAt, _ = bluge.DecodeDateTime(value)
			case "@timestamp":
				user.Timestamp, _ = bluge.DecodeDateTime(value)
			}
			return true
		}); err != nil {
			log.Printf("error accessing stored fields: %v", err)
		}

		hit := v1.Hit{
			Index:     usersIndex.Name,
			Type:      usersIndex.Name,
			ID:        user.ID,
			Score:     next.Score,
			Timestamp: user.Timestamp,
			Source:    user,
		}

		next, err = dmi.Next()
		Hits = append(Hits, hit)
	}

	resp := v1.SearchResponse{
		Took:     int(dmi.Aggregations().Duration().Milliseconds()),
		MaxScore: dmi.Aggregations().Metric("max_score"),
		Hits: v1.Hits{
			Total: v1.Total{
				Value: int(dmi.Aggregations().Count()),
			},
			Hits: Hits,
		},
	}

	return resp, nil
}

func CreateUser(userId, name, plainPassword, role string) (*ZincUser, error) {
	userExists, existingUser, err := GetUser(userId)
	if err != nil {
		return nil, err
	}

	var user *ZincUser
	if userExists {
		user = &existingUser
		if plainPassword != "" {
			user.Salt = GenerateSalt()
			user.Password = GeneratePassword(plainPassword, user.Salt)
		}
		user.Name = name
		user.Role = role
		user.Timestamp = time.Now()
	} else {
		user = &ZincUser{
			SimpleUser: SimpleUser{
				ID:        userId,
				Name:      name,
				Role:      role,
				CreatedAt: time.Now(),
				Timestamp: time.Now(),
			},
		}

		user.Salt = GenerateSalt()
		user.Password = GeneratePassword(plainPassword, user.Salt)
	}

	doc := bluge.NewDocument(user.ID)

	doc.AddField(bluge.NewTextField("name", user.Name).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("password", []byte(user.Password)).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("role", []byte(user.Role)).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("salt", []byte(user.Salt)).StoreValue())
	doc.AddField(bluge.NewDateTimeField("created_at", user.CreatedAt).StoreValue())
	doc.AddField(bluge.NewDateTimeField("updated_at", user.Timestamp).StoreValue())

	doc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))

	usersIndexWriter := core.ZincSystemIndexList[core.SystemIndexUsers].Writer

	if err = usersIndexWriter.Update(doc.ID(), doc); err != nil {
		fmt.Println("error updating document:", err)
		return nil, err
	}

	return user, nil
}

func GeneratePassword(password, salt string) string {
	params := &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  128,
		KeyLength:   32,
		Time:        2,
		Threads:     4,
	}

	hash := argon2.IDKey([]byte(password), []byte(salt), params.Time, params.Memory, params.Threads, params.KeyLength)

	return string(hash)
}

func GenerateSalt() string {
	return uuid.New().String()
}

type SimpleUser struct {
	ID        string    `json:"_id"` // this will be email
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Timestamp time.Time `json:"@timestamp"`
}

type ZincUser struct {
	SimpleUser
	Salt     string `json:"salt"`
	Password string `json:"password"`
}

type Argon2Params struct {
	Time        uint32
	Memory      uint32
	Threads     uint8
	KeyLength   uint32
	SaltLength  uint32
	Parallelism uint8
	Iterations  uint32
}

func DeleteUser(userId string) bool {
	bdoc := bluge.NewDocument(userId)
	bdoc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))
	usersIndexWriter := core.ZincSystemIndexList[core.SystemIndexUsers].Writer

	err := usersIndexWriter.Delete(bdoc.ID())
	if err != nil {
		fmt.Println("error deleting user:", err)
		return false
	}

	return true
}

func ZincAuth(c *gin.Context) {
	// Get the Basic Authentication credentials
	user, password, hasAuth := c.Request.BasicAuth()

	if !hasAuth {
		c.AbortWithStatusJSON(401, gin.H{
			"auth": "Missing credentials",
		})
		return
	}

	if _, ok := VerifyUser(user, password); ok {
		c.Next()
		return
	}

	c.AbortWithStatusJSON(401, gin.H{"auth": "Invalid credentials"})
	return
}

func VerifyUser(user, password string) (SimpleUser, bool) {
	reader, _ := core.ZincSystemIndexList[core.SystemIndexUsers].Writer.Reader()
	defer reader.Close()

	termQuery := bluge.NewTermQuery(user).SetField("_id")
	searchRequest := bluge.NewTopNSearch(1000, termQuery)

	dmi, err := reader.Search(context.Background(), searchRequest)
	if err != nil {
		log.Printf("error executing search: %v", err)
	}

	storedSalt := ""
	storedPassword := ""
	var sUser SimpleUser

	for next, err := dmi.Next(); err == nil && next != nil; {
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case "salt":
				storedSalt = string(value)
			case "_id":
				sUser.ID = string(value)
			case "password":
				storedPassword = string(value)
			case "name":
				sUser.Name = string(value)
			case "role":
				sUser.Role = string(value)
			}
			return true
		})
		if err != nil {
			log.Printf("error accessing stored fields: %v", err)
		}

		if GeneratePassword(password, storedSalt) == storedPassword {
			return sUser, true
		}

		next, err = dmi.Next()
	}

	return sUser, false
}
