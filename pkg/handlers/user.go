package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prabhatsharma/zinc/pkg/auth"
)

func GetUsers(c *gin.Context) {
	res, err := auth.GetAllUsersWorker()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, res)
}

func ValidateUser(c *gin.Context) {
	var user auth.ZincUser

	c.BindJSON(&user)

	validationResult, loggedInUser := auth.VerifyCredentials(user.ID, user.Password)
	c.JSON(200, gin.H{
		"validated": validationResult,
		"user":      loggedInUser,
	})
}

func CreateUpdateUser(c *gin.Context) {
	var user auth.ZincUser
	c.BindJSON(&user)

	newUser, err := auth.CreateUser(user.ID, user.Name, user.Password, user.Role)
	if err != nil {
		c.JSON(200, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": newUser,
	})
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("userID")

	c.JSON(200, gin.H{
		"deleted": auth.DeleteUser(userID),
	})
}
