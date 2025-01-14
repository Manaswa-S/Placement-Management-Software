package handlers

import (
	"fmt"

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
	adminRoute.GET("/testdatadeletelater", h.TestResponsesGetDataCloud)
}

func (h *AdminHandler) AdminDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/admindashboard.html")
}


func (h *AdminHandler) TestResponsesGetDataCloud(ctx *gin.Context) {
	
	
	changeList, err := h.AdminService.AdminFunc(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	

	ctx.JSON(200, changeList)

}