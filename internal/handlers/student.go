package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mod/internal/services"
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
	// get upcoming events template
	studentRoute.GET("/upcoming", h.UpcomingStatic)
	// upcoming events data with a filter
	studentRoute.GET("/upcomingdata", h.UpcomingData)
	// get take test template
	studentRoute.GET("/taketest", h.TakeTestStatic)
	// sends data for a question given the testid, and itemid
	studentRoute.GET("/taketestdata", h.TakeTest)
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

	jobType := ctx.Query("jobType")
	if jobType == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Job type not specified",
		})
		return
	}

	// call the service that sends all job listings that the user has not yet applied for
	alljobs, err := h.StudentService.GetApplicableJobs(ctx, jobType)
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

	// get filters off the request body
	status := ctx.Query("status")
	if status == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "no status specified in request body",
		})
		fmt.Println("error")
		return
	}

	userID, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user ID not found in token",
		})
		fmt.Println("erro")
		return
	}

	// call service 
	applicationsData, err := h.StudentService.MyApplications(ctx, userID, status)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, applicationsData)
}
func (h *StudentHandler) UpcomingStatic(ctx *gin.Context) {
	ctx.File("./template/student/upcoming.html")
}
func (h *StudentHandler) UpcomingData(ctx *gin.Context) {
	// get ID and eventttype
	userid, exists := ctx.Get("ID")
	eventtype := ctx.Query("eventtype")
	if !exists || eventtype == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing ID or eventtype",
		})
		return
	}
	// service delegation
	allUpcomingData, err := h.StudentService.UpcomingData(ctx, userid.(int64), eventtype)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// return data
	ctx.JSON(http.StatusOK, allUpcomingData)
}

func (h *StudentHandler) TakeTestStatic(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	testID := ctx.Query("testid")
	if testID == "" || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing test ID or user ID",
		})
		return
	}

	testMetadata, err := h.StudentService.TestMetadata(ctx, userid.(int64), testID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	
	// send the template
	ctx.HTML(200, "takeTest.html", gin.H{
		"TestID": testID,
		"TestName": testMetadata.TestName,
		"TestDescription": testMetadata.Description.String,
		"TestDuration": testMetadata.Duration,
		"TestEndDate": strings.Split(testMetadata.EndTime, " ")[0],
		"TestEndTime": strings.Split(testMetadata.EndTime, " ")[1],
		"QCount": testMetadata.QCount,
		"TestType": testMetadata.Type,
	})
}
func (h *StudentHandler) TakeTest(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	testid := ctx.Query("testid")
	currentItemId := ctx.Query("itemid")
	if !exists || testid == "" || currentItemId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing ID or eventtype",
		})
		return
	}

	result, err := h.StudentService.TakeTest(ctx, userid.(int64), testid, currentItemId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}

	ctx.JSON(200, result)
}