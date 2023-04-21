package main

import (
	"log"

	"github.com/MohammadBnei/gorm-user-auth/config"
	"github.com/MohammadBnei/gorm-user-auth/handler"
	"github.com/MohammadBnei/gorm-user-auth/model"
	"github.com/MohammadBnei/gorm-user-auth/service"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.InitConfig()
	db, err := config.InitDB(conf)
	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&model.User{}, &model.RefreshToken{})

	userService := service.NewUserService(db)
	rtService := service.NewRTService(db)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(rtService, userService, conf)

	r := gin.Default()

	userApi := r.Group("/api/v1/user")
	userApi.GET("/:id", userHandler.GetUser)
	userApi.GET("/", userHandler.GetUsers)
	userApi.POST("/", userHandler.CreateUser)
	userApi.PUT("/:id", userHandler.UpdateUser)
	userApi.DELETE("/:id", userHandler.DeleteUser)

	authApi := r.Group("/api/v1/auth")
	authApi.POST("/login", authHandler.Login)

	r.GET("/test/auth", authHandler.AuthMiddleware(), func(c *gin.Context) {
		user, exist := c.Get("user")

		if !exist {
			c.JSON(401, gin.H{
				"error": "no user in the context",
			})
			return
		}
		c.JSON(200, gin.H{
			"user": user,
		})
	})

	r.Run()
}
