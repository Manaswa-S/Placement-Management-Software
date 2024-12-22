package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
)

type StudentHandler struct {
	StudentService *services.StudentService
}

func NewStudentHandler(studentService *services.StudentService) *StudentHandler {
	return &StudentHandler{
		StudentService: studentService,
	}
}
func (h *StudentHandler) RegisterRoute(studentRoute *gin.RouterGroup) {
	// get the dashboard
	studentRoute.GET("/dashboard", h.StudentDashboard)
	// get the template for jobs list
	studentRoute.GET("/jobslist", h.JobsList)
	// get list of applicable jobs as JSON
	studentRoute.GET("/alljobs", h.ApplicableJobs)
	// post and apply to a job
	studentRoute.POST("/applytojob", h.ApplyToJob)
	studentRoute.GET("/cancelapplication", h.CancelApplication)
	// get template
	studentRoute.GET("/myappsstatic", h.MyAppsStatic)
	// get applied job list
	studentRoute.GET("/myapplications", h.MyApplications)
}


func (h *StudentHandler) StudentDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/studentdashboard.html")
}
func (h *StudentHandler) JobsList(ctx *gin.Context) {
	ctx.File("./template/jobs/alljobslist.html")
}
func (h *StudentHandler) MyAppsStatic(ctx *gin.Context) {
	ctx.File("./template/jobs/myapplications.html")
}
func (h *StudentHandler) ApplicableJobs(ctx *gin.Context) {
	// TODO: get the filters of the request body  


	// call the service that sends all job listings that the user has not yet applied for
	alljobs, err := h.StudentService.GetApplicableJobs(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// return
	ctx.JSON(http.StatusOK, alljobs)
}
func (h *StudentHandler) ApplyToJob(ctx *gin.Context) {

	// get jobid off the request
	jobId:= ctx.Query("jobid")
	if jobId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid job id, try again or contact admin",
		})
		return
	}
	// get ID off the context/token
	userID, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user id, try logging in again",
		})
		return
	}
	// call service to add application to the database
	err := h.StudentService.NewApplication(ctx, userID.(int64), jobId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// 200OK code 
	ctx.JSON(http.StatusOK, gin.H{
		"status": "applied to job successfully",
	})
}
func (h *StudentHandler) CancelApplication(ctx *gin.Context) {
	// get user id from context
	userID, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "error getting user ID", 
		})
		return
	}
	// get job id from request body
	jobID := ctx.Query("jobid")
	if jobID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "no job ID specified in request body", 
		})
		return
	}
	// call service
	err := h.StudentService.CancelApplication(ctx, userID.(int64), jobID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// respond with 200
	ctx.JSON(http.StatusOK, gin.H{
		"status": "successfully canceled application",
	})
}
func (h *StudentHandler) MyApplications(ctx *gin.Context) {
	// pre-defined filters map
	filters := map[string]bool{
		"All":true,
		"Applied": true,
		"UnderReview": true,
		"ShortListed": true,
		"Rejected": true,
		"Offered": true,
		"Hired": true,
	}
	// get filters off the request body
	status := ctx.Query("status")
	if _, exists := filters[status]; !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "unknown filter",
		})
		return
	}

	// call service 
	applicationsData, err := h.StudentService.MyApplications(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var filteredData []sqlc.GetMyApplicationsRow
	if status != "All" {
		filteredData, err = utils.FilterFromSliceOf(applicationsData, "Status", status)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, filteredData)
		return
	}

	ctx.JSON(http.StatusOK, applicationsData)
}