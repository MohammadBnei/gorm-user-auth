package handler

import (
	"log"
	"strconv"

	"github.com/MohammadBnei/gorm-user-auth/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

/*
GetUser gets a user by their ID from the userService and returns it in the response body.

Parameters:
  - c (*gin.Context): the context of the current HTTP request
  - h (*UserHandler): the handler that handles user-related requests

Errors:
  - 400 Bad Request: if the parameter id cannot be converted to an integer, or if there is an error retrieving the user
*/
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, user)
}
