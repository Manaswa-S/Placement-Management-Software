package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type PublicHandler struct {
	UserService *services.PublicService
}

func NewPublicHandler(userService *services.PublicService) *PublicHandler {
	return &PublicHandler{
		UserService: userService,
	}
}

func (h *PublicHandler) RegisterRoute(publicRoute *gin.RouterGroup) {
	publicRoute.GET("/getpub", h.PublicHandlerFunc)
}

func (h *PublicHandler) PublicHandlerFunc(c *gin.Context) {

	ctx := c.Request.Context()

	data, err := h.UserService.PublicS(ctx)
	if err != nil {
		data = "error"
	}
	c.JSON(200, gin.H{
		"handler": "PublicHandlerFunc",
		"service": data,
	})
}