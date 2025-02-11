package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
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
	studentRoute.GET("/dashboarddata", h.DashboardData)

	// get the notifications data
	studentRoute.GET("/notifications", h.GetNotifications)


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
	studentRoute.POST("/taketestdata", h.TakeTest)
	// submit test responses
	studentRoute.GET("/submittest", h.SubmitTest) // TODO:

	// get the completed page template
	studentRoute.GET("/completed", h.CompletedStatic)
	studentRoute.GET("/completeddata", h.Completed)

	
	// get profile template
	studentRoute.GET("/profile", h.GetProfile)
	// get the complete profile data
	studentRoute.GET("/profiledata", h.ProfileData) 

	// get the file specified as query for the user id 
	studentRoute.GET("/getfile", h.GetFile)

	// update the student's details
	studentRoute.POST("/updatedetails", h.UpdateDetails)
	// update student's documents/files 
	studentRoute.POST("/updatefile", h.UpdateFile)
	
}


func (h *StudentHandler) StudentDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/studentdashboard.html")
}

func (h *StudentHandler) DashboardData(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing user ID in request.",
		})
		return
	}

	data, err := h.StudentService.DashboardData(ctx, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: err.Error(),
		})
		fmt.Println(err)
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (h *StudentHandler) GetNotifications(ctx *gin.Context) {

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

	notifs, errf := h.StudentService.Notify.GetNotifications(ctx, userid.(int64), start, end)
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

func (h *StudentHandler) JobsList(ctx *gin.Context) {
	ctx.File("./template/student/alljobslist.html")
}
func (h *StudentHandler) MyAppsStatic(ctx *gin.Context) {
	ctx.File("./template/student/myapplications.html")
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
	ctx.JSON(http.StatusOK, gin.H{
		"JobsList": alljobs,
	})
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
		return
	}

	userID, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user ID not found in token",
		})
		return
	}

	// call service 
	applicationsData, err := h.StudentService.MyApplications(ctx, userID, status)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Applications": applicationsData,
	})
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
	uData, err := h.StudentService.UpcomingData(ctx, userid.(int64), eventtype)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// return data
	ctx.JSON(http.StatusOK, uData)
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
		"TestEndTime": testMetadata.EndTime,
		"QCount": testMetadata.QCount,
		"TestType": testMetadata.Type,
	})
}
func (h *StudentHandler) TakeTest(ctx *gin.Context) {
	var data dto.TestResponse
	err := ctx.Bind(&data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "response data not found",
		})
		return
	}

	userid, exists := ctx.Get("ID")
	testid := ctx.Query("testid")
	currentItemId := ctx.Query("itemid")
	if !exists || testid == "" || currentItemId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing ID or eventtype",
		})
		return
	}

	result, errf := h.StudentService.TakeTest(ctx, userid.(int64), testid, currentItemId, data)
	if errf != nil {
		if (errf.Type != errs.Internal) {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
			ctx.JSON(http.StatusInternalServerError, errf)
		}
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *StudentHandler) SubmitTest(ctx *gin.Context) {
	// get the userid and testid from request
	userid, exists := ctx.Get("ID")
	testid := ctx.Query("testid")
	if testid == "" || !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Message": "Bad Request : Missing userid or testid ",
		})
		return
	}
	// service delegation
	errf := h.StudentService.SubmitTest(ctx, userid.(int64), testid)
	if errf != nil {
		ctx.JSON(http.StatusInternalServerError, errf)
		return 
	}
	// all good
	ctx.Status(http.StatusOK)
}

func (h *StudentHandler) CompletedStatic(ctx *gin.Context) {
	ctx.File("./template/student/completed.html")
}
func (h *StudentHandler) Completed(ctx *gin.Context) {
	// get user id from request
	userid, exists := ctx.Get("ID")
	tab := ctx.Query("tab")
	if !exists || tab == "" {
		fmt.Println("user id or tab value not found")
		return
	}	
	
	// service delegation
	cData, err := h.StudentService.Completed(ctx, userid.(int64), tab)
	if err != nil {
		fmt.Println(err)
		return
	}

	// send data
	ctx.JSON(http.StatusOK, cData)

}



func (h *StudentHandler) GetProfile(ctx *gin.Context) {
	ctx.File("./template/student/myProfile.html")
}
func (h *StudentHandler) ProfileData(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	if (!exists) {
		fmt.Println("Please provide user ID")
		return
	}

	data, err := h.StudentService.ProfileData(ctx, userid.(int64))
	if err != nil {
		fmt.Println(err)
		return
	}
	

	ctx.JSON(http.StatusOK, data)
}

func (h *StudentHandler) GetFile(ctx *gin.Context) {
	// get user if and file type requested
	userid, exists := ctx.Get("ID")
	fileType := ctx.Query("type")
	if (!exists || fileType == "") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing user ID or query parameter",
		})
		return
	}
	// get the file path 
	filePath, errf := h.StudentService.GetStudentFile(ctx, userid.(int64), fileType)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return 
	}
	// respond with file
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	ctx.File(filePath)
}

func (h *StudentHandler) UpdateDetails(ctx *gin.Context) {
	details := new(dto.UpdateStudentDetails)

	userid, exists := ctx.Get("ID")
	err := ctx.Bind(details)
	if err != nil || !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing user ID or failed to bind incoming data",
		})
		return
	}	

	err = h.StudentService.UpdateDetails(ctx, userid.(int64), details)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)		
}


func (h *StudentHandler) UpdateFile(ctx *gin.Context) {
	
	fileType := ctx.Query("type")
	userid, exists := ctx.Get("ID")
	if fileType == "" || !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "missing user ID or file type query",
		})
		return
	}
	
	file, err := ctx.FormFile(fileType)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get file",
		})
		return
	}

	errf := h.StudentService.UpdateFile(ctx, userid.(int64), file, fileType)
	if errf != nil {
		if (errf.Type != errs.Internal) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
		}
		return
	}

	ctx.Status(http.StatusOK)


}