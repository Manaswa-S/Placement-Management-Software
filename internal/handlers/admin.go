package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type AdminHandler struct {
	UserService *services.AdminService
}

func NewAdminHandler(userService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		UserService: userService,
	}
}

func (h *AdminHandler) RegisterRoute(adminRoute *gin.RouterGroup) {
	adminRoute.GET("/getall", h.AdminHandlerFunc)
}

func (h *AdminHandler) AdminHandlerFunc(c *gin.Context) {

	ctx := c.Request.Context()

	data, err := h.UserService.AdminS(ctx)
	if err != nil {
		data = "error"
	}
	c.JSON(200, gin.H{
		"handler": "AdminHandlerFunc",
		"service": data,
	})
}