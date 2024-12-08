package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type SuperUserHandler struct {
	UserService *services.SuperUserService
}

func NewSuperUserHandler(userService *services.SuperUserService) *SuperUserHandler {
	return &SuperUserHandler{
		UserService: userService,
	}
}

func (h *SuperUserHandler) RegisterRoute(superuserRoute *gin.RouterGroup) {
	superuserRoute.GET("/getsup", h.SuperUserHandlerFunc)
}

func (h *SuperUserHandler) SuperUserHandlerFunc(c *gin.Context) {

	ctx := c.Request.Context()

	data, err := h.UserService.SuperUserS(ctx)
	if err != nil {
		data = "error"
	}
	c.JSON(200, gin.H{
		"handler": "SuperUserHandlerFunc",
		"service": data,
	})
}