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

	db.AutoMigrate(&model.User{})

	userService := service.NewUserService(db)
	userHandler := handler.NewUserHandler(userService)

	r := gin.Default()

	userApi := r.Group("/api/v1/user")
	userApi.GET("/:id", userHandler.GetUser)

	r.Run()
}
