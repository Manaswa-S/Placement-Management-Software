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
	adminRoute.GET("/dashboard", h.AdminDashboard)
}

func (h *AdminHandler) AdminDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/admindashboard.html")

}