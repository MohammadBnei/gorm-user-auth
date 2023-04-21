package handler

import (
	"fmt"
	"time"

	"github.com/MohammadBnei/gorm-user-auth/config"
	"github.com/MohammadBnei/gorm-user-auth/model"
	"github.com/MohammadBnei/gorm-user-auth/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	RTService   *service.RTService
	UserService *service.UserService
	*config.Config
}

func NewAuthHandler(rTService *service.RTService, userService *service.UserService, config *config.Config) *AuthHandler {
	return &AuthHandler{
		RTService:   rTService,
		UserService: userService,
		Config:      config,
	}
}

/*
GenerateToken generates a JWT token for a given user.

Args:

	AuthHandler (*AuthHandler): A pointer to the AuthHandler object.
	user (*model.User): A pointer to the User object.

Returns:

	string: The generated JWT token.
	error: An error if one occurred during the generation process.
*/
func (authHandler *AuthHandler) GenerateToken(user *model.User) (string, error) {

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(authHandler.JWT_SECRET))

}

/*
Login handles the login request. It parses the request body into a LoginDTO struct
and attempts to retrieve a user from the UserService instance with the email provided
in the LoginDTO. If a user is found, the password is checked against the user's hashed
password. If the password matches, a JWT is generated and set as a cookie in the response.
A refresh token is also generated and set as a cookie in the response. Finally, a JSON
response is returned with the JWT, the refresh token, and the user object.

@param authHandler *AuthHandler: an instance of the AuthHandler struct
@param c *gin.Context: the current request context

@return none
*/
func (authHandler *AuthHandler) Login(c *gin.Context) {
	var loginDTO *model.LoginDTO

	if err := c.ShouldBindJSON(&loginDTO); err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := authHandler.UserService.GetUserByEmail(loginDTO.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = user.CheckPassword(loginDTO.Password)
	if err != nil {
		fmt.Println(err)
		if err == bcrypt.ErrMismatchedHashAndPassword {
			c.JSON(400, gin.H{
				"error": "incorrect password",
			})
		} else {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	jwt, err := authHandler.GenerateToken(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	rt, err := authHandler.RTService.CreateRT(c.ClientIP(), int(user.ID))
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.SetCookie("jwt", jwt, 3600, "/", "*", true, true)
	c.SetCookie("rt", rt.Hash, 3600, "/", "*", true, true)

	c.JSON(200, gin.H{
		"token":        jwt,
		"refreshToken": rt.Hash,
		"user":         user,
	})
}
