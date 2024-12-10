package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type SuperUserHandler struct {
	SuperService *services.SuperService
}

func NewSuperUserHandler(superService *services.SuperService) *SuperUserHandler {
	return &SuperUserHandler{
		SuperService: superService,
	}
}

func (h *SuperUserHandler) RegisterRoute(superuserRoute *gin.RouterGroup) {
	superuserRoute.GET("/getsup", h.SuperUserHandlerFunc)
}

func (h *SuperUserHandler) SuperUserHandlerFunc(c *gin.Context) {

	h.SuperService.SuperFunc()

	c.JSON(200, gin.H{
		"handler": "SuperUserHandlerFunc",
	})
}