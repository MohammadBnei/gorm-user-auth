package service

import (
	"fmt"

	"github.com/MohammadBnei/gorm-user-auth/model"
	"github.com/kjk/betterguid"
	"gorm.io/gorm"
)

type RTService struct {
	db *gorm.DB
}

func NewRTService(db *gorm.DB) *RTService {
	return &RTService{
		db: db,
	}
}

/*
CreateRT creates a new refresh token with the provided IP address and user ID.

Args:
  - ip (string): The IP address associated with the token.
  - userId (int): The ID of the user associated with the token.

Returns:
  - (*model.RefreshToken): The newly created refresh token.
  - (error): An error if one occurred during database save.
*/
func (rt *RTService) CreateRT(ip string, userId int) (*model.RefreshToken, error) {
	hash := betterguid.New()

	token := &model.RefreshToken{
		Hash:   hash,
		Ip:     ip,
		UserId: userId,
	}

	err := rt.db.Save(token).Error
	if err != nil {
		return nil, err
	}

	var previousTokens []model.RefreshToken
	err = rt.db.Where("ip = ? AND user_id = ? AND NOT hash = ?", ip, userId, hash).Delete(previousTokens).Error
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (rt *RTService) GetRT(hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := rt.db.Where("hash = ?", hash).Preload("User").First(&token).Error
	fmt.Println(token, err)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
