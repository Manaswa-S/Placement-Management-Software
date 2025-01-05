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
	
	// query := "mimeType='application/vnd.google-apps.form' and sharedWithMe=true"

	// allFormsData, err := driveService.Files.List().Q(query).Fields("files(id, webViewLink, name)").Do()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// respondersURL := "https://docs.google.com/forms/d/e/1FAIpQLSfHcaoaaMcCNTWNY83bccAFcG2mhG4izHtKu5aODu8lwTeB7A/viewform"

	// form, err := formsService.Forms.Get(file.Id).Fields("responderUri", "formId").Context(ctx).Do()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	changeList, err := h.AdminService.AdminFunc()
	if err != nil {
		fmt.Println(err)
		return
	}
	

	ctx.JSON(200, changeList.Changes)

}