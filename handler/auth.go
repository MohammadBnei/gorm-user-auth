package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MohammadBnei/gorm-user-auth/config"
	"github.com/MohammadBnei/gorm-user-auth/model"
	"github.com/MohammadBnei/gorm-user-auth/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	c.SetCookie("jwt", jwt, 3600, "/", "*", false, true)
	c.SetCookie("rt", rt.Hash, 3600, "/", "*", false, true)

	c.JSON(200, gin.H{
		"token":        jwt,
		"refreshToken": rt.Hash,
		"user":         user,
	})
}

/*
AuthMiddleware is a middleware function that handles user authentication using JWT tokens.

Parameters:
- authHandler (*AuthHandler): A pointer to an AuthHandler instance containing JWT_SECRET.
- c (*gin.Context): A pointer to the gin.Context instance.

Returns:
- gin.HandlerFunc: A function that handles the middleware.
*/
func (authHandler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request

		// First, trying to extract the jwt from the cookie
		jwtToken, err := c.Cookie("jwt")

		// If not present, proceed to extract it from the Authorization header
		if err == http.ErrNoCookie {
			authHeader := c.GetHeader("Authorization")
			// Using Bearer prefix
			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				c.JSON(401, gin.H{
					"error": "no token provided",
				})
				c.Abort()
				return
			}
			jwtToken = splitToken[1]

			if jwtToken == "" {
				c.JSON(401, gin.H{
					"error": "no token provided",
				})
				c.Abort()
				return
			}
		}
		// If the error is anything else beside ErrNoCookie
		if err != nil {
			c.JSON(401, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// Parsing the token
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			// This is just an example of specific token verification
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Only this part is required
			return []byte(authHandler.JWT_SECRET), nil
		})

		// If the token is expired, let's trying to update it with the refresh token
		if errors.Is(err, jwt.ErrTokenExpired) {
			// This time, only getting the refresh token from the cookie. No header
			rtToken, err := c.Cookie("rt")
			// If we get a token, this part will handle all the logic. It means that it does not return to the main part.
			if err == nil {
				rt, err := authHandler.RTService.GetRT(rtToken)
				if err != nil {
					c.JSON(401, gin.H{
						"error": "token expired, unable to automatically refresh : " + err.Error(),
					})
					c.Abort()
					return
				}

				// By default, without using the Preload method, the user will be an empty struct
				if rt.User.ID == 0 {
					c.JSON(401, gin.H{
						"error": "token expired, unable to automatically refresh. Something went wrong retrieving the user",
					})
					c.Abort()
					return
				}

				c.Set("user", rt.User)

				// Regenerating the cookie and putting it in the response's cookies
				newJwt, err := authHandler.GenerateToken(&rt.User)
				if err != nil {
					fmt.Println(err)
				}

				c.SetCookie("jwt", newJwt, 3600, "/", "*", false, true)

				c.Next()

				return
			}
		}
		if err != nil {
			c.JSON(401, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		userId := token.Claims.(jwt.MapClaims)["id"].(float64)
		user, err := authHandler.UserService.GetUser(int(userId))
		if err != nil {
			c.JSON(401, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("user", user)

		c.Next()

		// after request
	}
}
