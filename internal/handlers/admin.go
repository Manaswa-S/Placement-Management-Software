package handlers

import (
	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
)

type AdminHandler struct {
	AdminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		AdminService: adminService,
	}
}

func (h *AdminHandler) RegisterRoute(adminRoute *gin.RouterGroup) {
	adminRoute.GET("/getall", h.AdminHandlerFunc)
}

func (h *AdminHandler) AdminHandlerFunc(c *gin.Context) {

	h.AdminService.AdminFunc()

	c.JSON(200, gin.H{
		"handler": "AdminHandlerFunc",
	})
}