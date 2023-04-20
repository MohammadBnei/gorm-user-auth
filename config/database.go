package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
InitDB initializes a GORM database connection using the provided Config.

Parameters:
- config (*Config): A pointer to the Config struct containing database connection details.

Returns:
- (*gorm.DB): A pointer to the GORM database object.
- (error): An error object if the connection fails, nil otherwise.
*/
func InitDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_PORT, config.DB_NAME)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
