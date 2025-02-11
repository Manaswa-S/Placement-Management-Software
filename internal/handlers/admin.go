package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
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
	// get the notifications data
	adminRoute.GET("/notifications", h.GetNotifications)

	// get all students info
	adminRoute.GET("/studentinfo", h.StudentInfo)

	// get the static 'manage students' template
	adminRoute.GET("/managestudents", h.ManageStudentsStatic)
	// get the 'manage students' data
	adminRoute.GET("/managestudentsdata", h.ManageStudents)

	adminRoute.GET("/verifyst", h.VerifyStudent)

	// generates the test results, returns them, and triggers other funcs
	adminRoute.GET("/testresult", h.GenerateTestResult)

}



func (h *AdminHandler) AdminDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/admindashboard.html")
}

func (h *AdminHandler) GetNotifications(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing user ID",
		})
		return
	}

	start := ctx.Query("start")
	end := ctx.Query("end")
	if start == "" || end == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing query params start and end",
		})
		return
	}

	notifs, errf := h.AdminService.Notify.GetNotifications(ctx, userid.(int64), start, end)
	if errf != nil {
		if errf.Type != errs.Internal {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
			return
		}
		// TODO: add the internal errors logic 
		return
	}

	ctx.JSON(http.StatusOK, notifs)
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


func (h *AdminHandler) GenerateTestResult(ctx *gin.Context) {

	err := h.AdminService.GenerateTestResult(ctx, "10084")
	if err != nil {
		fmt.Println(err)
	}

	ctx.File(os.Getenv("ResultDraftStorage"))
}