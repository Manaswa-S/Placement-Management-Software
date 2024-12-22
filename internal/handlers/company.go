package handlers

import (
	"fmt"
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
	// get the template for all applicants
	companyRoute.GET("/myappsstatic", h.MyApplicantsStatic)
	// get all applicants data
	companyRoute.GET("/myapplicants", h.MyApplicants)
	// get any file (resume, result)
	companyRoute.GET("/getfile", h.GetFile)
	// get my job listings static
	companyRoute.GET("/myjobliststatic", h.MyJobListStatic)
	// get my job listings
	companyRoute.GET("/myjoblistings", h.MyJobListings)
	// cancel job listing
	companyRoute.GET("/closejob", h.CloseJob)
	// delete job listing
	companyRoute.GET("/deletejob", h.DeleteJob)


	companyRoute.POST("/shortlist", h.ShortList)
	companyRoute.POST("/reject", h.Reject)
	companyRoute.POST("/offer", h.Offer)
	companyRoute.POST("/hire", h.Hire)
	companyRoute.POST("/scheduleinterview", h.ScheduleInterview)

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

func (h *CompanyHandler) MyApplicantsStatic(ctx *gin.Context) {
	ctx.File("./template/jobs/myapplicants.html")
}

func (h *CompanyHandler) MyApplicants(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	userID, exists := ctx.Get("ID")
	if !exists { 
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user ID not found in token",
		})
		return
	}

	applicantsData, err := h.CompanyService.MyApplicants(ctx, userID.(int64), jobid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err.Error())
		return
	}

	ctx.JSON(http.StatusOK, applicantsData)
}

func (h *CompanyHandler) GetFile(ctx *gin.Context) {

	studentid := ctx.Query("studentid")
	jobid := ctx.Query("jobid")
	filetype := ctx.Query("type")
	if studentid == "" || jobid == "" || filetype == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "path or type not specified",
		})
		return
	}

	filePath, err := h.CompanyService.GetFilePath(ctx, studentid, jobid, filetype)
	if err != nil {
		ctx. AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// browser heavily caches these files
	// prevent it from doing that by setting this header
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	
	ctx.File(filePath)
}

func (h *CompanyHandler) MyJobListStatic(ctx *gin.Context) {
	ctx.File("./template/jobs/myjoblistings.html")
}

func (h *CompanyHandler) MyJobListings(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "unable to get user ID",
		})
		fmt.Println("error")
		return
	}

	jobListings, err := h.CompanyService.MyJobListings(ctx, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, jobListings)
}

func (h *CompanyHandler) CloseJob(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user id not found",
		})
		return
	}

	err := h.CompanyService.CloseJob(ctx, jobid, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "closed job successfully",
	})
}

func (h *CompanyHandler) DeleteJob(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user id not found",
		})
		return
	}

	err := h.CompanyService.DeleteJob(ctx, jobid, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "delete job successfully",
	})
}


func (h *CompanyHandler) ShortList(ctx *gin.Context) {
	
	studentid := ctx.Query("studentid")
	jobid := ctx.Query("jobid")
	if jobid == "" || studentid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing jobid or studentid",
		})
		return
	}

	err := h.CompanyService.ShortList(ctx, studentid, jobid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) Reject(ctx *gin.Context) {

	studentid := ctx.Query("studentid")
	jobid := ctx.Query("jobid")
	if jobid == "" || studentid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing jobid or studentid",
		})
		return
	}

	err := h.CompanyService.Reject(ctx, studentid, jobid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) ScheduleInterview(ctx *gin.Context) {

	var data dto.NewInterview
	err := ctx.Bind(&data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	data.UserId = ctx.GetInt64("ID")

	intData, err := h.CompanyService.ScheduleInterview(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, intData)
}

func (h *CompanyHandler) Offer(ctx *gin.Context) {
	fmt.Println("offered")
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) Hire(ctx *gin.Context) {
	fmt.Println("hired")
	ctx.Status(http.StatusOK)
}