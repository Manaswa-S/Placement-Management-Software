package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mod/internal/dto"
	"go.mod/internal/services"
)

type CompanyHandler struct {
	CompanyService *services.CompanyService
}

func NewCompanyHandler(companyService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		CompanyService: companyService,
	}
}

func (h *CompanyHandler) RegisterRoute(companyRoute *gin.RouterGroup) {
	// get dashboard
	companyRoute.GET("/dashboard", h.CompanyDashboard)
	// get new job posting form
	companyRoute.GET("/newjobget", h.NewJobGet)
	// post new job form
	companyRoute.POST("/newjobpost", h.NewJobPost)
}

func (h *CompanyHandler) CompanyDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/companydashboard.html")
}

func (h *CompanyHandler) NewJobGet(ctx *gin.Context) {

	// we can use an external form or on-site html too
	GoogleFormLink := os.Getenv("NewJobFormLink")

	if GoogleFormLink == "" {
		ctx.File("./template/jobs/newjobform.html")
		return
	} else {
		ctx.Redirect(http.StatusSeeOther, GoogleFormLink)
		return
	}
}

func (h *CompanyHandler) NewJobPost(ctx *gin.Context) {
	
	// bind incoming request form
	var jobdata dto.NewJobData
	err := ctx.Bind(&jobdata)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call appropriate service
	_, err = h.CompanyService.NewJobPost(ctx, jobdata)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// redirect to dashboard
	// TODO: the user can just go back and submit form again which is dangerous
	ctx.Redirect(http.StatusSeeOther, "/laa/company/dashboard")
}