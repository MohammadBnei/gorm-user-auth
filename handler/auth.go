package handler

import (
	"github.com/MohammadBnei/gorm-user-auth/service"
	"github.com/gin-gonic/gin"
)

type rtHandler struct {
	RTService *service.RTService
}

func NewRefreshTokenHandler(rtService *service.RTService) *rtHandler {
	return &rtHandler{
		RTService: rtService,
	}
}

func (h *rtHandler) RefreshToken(c *gin.Context) {

}
