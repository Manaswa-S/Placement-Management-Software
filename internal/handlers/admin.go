package handlers

import (
	"fmt"
	"net/http"

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
	// get the static dashboard template
	adminRoute.GET("/dashboard", h.AdminDashboard)

	// get all students info
	adminRoute.GET("/studentinfo", h.StudentInfo)

	// get the static 'manage students' template
	adminRoute.GET("/managestudents", h.ManageStudentsStatic)
	// get the 'manage students' data
	adminRoute.GET("/managestudentsdata", h.ManageStudents)

	adminRoute.GET("/verify", h.VerifyStudent)


}

func (h *AdminHandler) AdminDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/admindashboard.html")
}

func (h *AdminHandler) StudentInfo(ctx *gin.Context) {

	userid := ctx.Query("id")
	if userid == "" {
		fmt.Println("empty id param in query")
		return
	}

	data, err := h.AdminService.StudentInfo(ctx, userid)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Data": data,
	})
}

func (h *AdminHandler) ManageStudentsStatic(ctx *gin.Context) {
	ctx.File("./template/admin/manageStudents.html")
}

func (h *AdminHandler) ManageStudents(ctx *gin.Context) {

	tab := ctx.Query("tab")
	if tab == "" {
		fmt.Println("wrong url structure or parameter")
		return
	}

	data, err := h.AdminService.ManageStudents(ctx, tab)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Data": data,
	})

}

func (h *AdminHandler) VerifyStudent(ctx *gin.Context) {

	userid := ctx.Query("id")
	if userid == "" {
		fmt.Println("invalid student id")
		return
	}

	err := h.AdminService.VerifyStudent(ctx, userid)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.Status(http.StatusOK)

}