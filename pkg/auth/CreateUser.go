package auth

import (
	"fmt"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/google/uuid"
	"github.com/prabhatsharma/zinc/pkg/core"
	"golang.org/x/crypto/argon2"
)

func CreateUser(userId, name, plaintextPassword, role string) (*ZincUser, error) {
	var newUser *ZincUser

	userExists, existingUser, err := GetUser(userId)
	if err != nil {
		return nil, err
	}

	if userExists {
		newUser = &existingUser
		if plaintextPassword != "" {
			newUser.Salt = GenerateSalt()
			newUser.Password = GeneratePassword(plaintextPassword, newUser.Salt)
		}
		newUser.Name = name
		newUser.Role = role
		newUser.Timestamp = time.Now()
	} else {
		newUser = &ZincUser{
			ID:        userId,
			Name:      name,
			Role:      role,
			CreatedAt: time.Now(),
			Timestamp: time.Now(),
		}

		newUser.Salt = GenerateSalt()
		newUser.Password = GeneratePassword(plaintextPassword, newUser.Salt)
	}

	doc := bluge.NewDocument(newUser.ID)

	doc.AddField(bluge.NewTextField("name", newUser.Name).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("password", []byte(newUser.Password)).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("role", []byte(newUser.Role)).StoreValue())
	doc.AddField(bluge.NewStoredOnlyField("salt", []byte(newUser.Salt)).StoreValue())
	doc.AddField(bluge.NewDateTimeField("created_at", newUser.CreatedAt).StoreValue())
	doc.AddField(bluge.NewDateTimeField("updated_at", newUser.Timestamp).StoreValue())

	doc.AddField(bluge.NewCompositeFieldExcluding("_all", nil))

	usersIndexWriter := core.ZincSystemIndexList["_users"].Writer

	err = usersIndexWriter.Update(doc.ID(), doc)
	if err != nil {
		fmt.Println("error updating document:", err)
		return nil, err
	}

	return newUser, nil
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

type ZincUser struct {
	ID        string    `json:"_id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Salt      string    `json:"salt"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	Timestamp time.Time `json:"@timestamp"`
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
