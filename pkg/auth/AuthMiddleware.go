package auth

import (
	"context"

	"github.com/blugelabs/bluge"
	"github.com/prabhatsharma/zinc/pkg/core"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

func ZincAuth(c *gin.Context) {
	// Get the Basic Authentication credentials
	user, password, hasAuth := c.Request.BasicAuth()

	if !hasAuth {
		c.AbortWithStatusJSON(401, gin.H{
			"auth": "Missing credentials",
		})
		return
	}

	if ok, _ := VerifyCredentials(user, password); ok {
		c.Next()
		return
	}

	c.AbortWithStatusJSON(401, gin.H{"auth": "Invalid credentials"})
	return
}

func VerifyCredentials(user, password string) (bool, SimpleUser) {
	reader, _ := core.ZincSystemIndexList["_users"].Writer.Reader()
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

		incomingEncryptedPassword := GeneratePassword(password, storedSalt)

		if incomingEncryptedPassword == storedPassword {
			return true, sUser
		}

		next, err = dmi.Next()
	}

	return false, sUser
}
