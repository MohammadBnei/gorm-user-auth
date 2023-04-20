package main

import (
	"log"

	"github.com/MohammadBnei/gorm-user-auth/config"
	"github.com/MohammadBnei/gorm-user-auth/model"
)

func main() {
	conf := config.InitConfig()
	db, err := config.InitDB(conf)
	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&model.User{})
}
