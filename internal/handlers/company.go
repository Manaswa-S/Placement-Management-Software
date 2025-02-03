package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
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
	companyRoute.GET("/dashboarddata", h.DashboardData)


	// get new job posting form
	companyRoute.GET("/newjob", h.NewJob)
	// post new job form
	companyRoute.POST("/newjobpost", h.NewJobPost)

	// get the template for all applicants
	companyRoute.GET("/applicants", h.ApplicantsStatic)
	// get all applicants data
	companyRoute.GET("/applicantsdata", h.ApplicantsData)

	// get any file (resume, result)
	companyRoute.GET("/getfile", h.GetResumeOrResultFile)
	// get my job listings static
	companyRoute.GET("/joblistings", h.JobListingsStatic)
	// get my job listings
	companyRoute.GET("/joblistingsdata", h.JobListingsData)
	// close job listing
	companyRoute.GET("/closejob", h.CloseJob)
	// delete job listing
	companyRoute.GET("/deletejob", h.DeleteJob)


	companyRoute.POST("/shortlist", h.ShortList)
	companyRoute.POST("/reject", h.Reject)
	companyRoute.POST("/offer", h.Offer)
	companyRoute.POST("/scheduleinterview", h.ScheduleInterview)
	companyRoute.POST("/cancelinterview", h.CancelInterview)

	companyRoute.GET("/newtest", h.NewTestStatic)
	companyRoute.POST("/newtestpost", h.NewTestPost)

	// get the scheduled events template
	companyRoute.GET("/scheduled", h.ScheduledStatic)
	// get the scheduled events data
	companyRoute.GET("/scheduleddata", h.ScheduledData)
	// update interview details
	companyRoute.POST("/updateinterview", h.UpdateInterview)



	// get the completed events template
	companyRoute.GET("/completed", h.CompletedStatic)
	// get the completed events data
	companyRoute.GET("/completeddata", h.CompletedData)
	// post the new test cut off
	companyRoute.POST("/editcutoff", h.EditCutOff)

	// publish individual results
	companyRoute.GET("/publishresults", h.PublishTestResults)


	

	// TODO:
	companyRoute.GET("/studentprofiledata", h.StudentProfileData)


}

func (h *CompanyHandler) CompanyDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/companydashboard.html")
}

func (h *CompanyHandler) DashboardData(ctx *gin.Context) {

	userid, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing user ID in request.",
		})
		return
	}

	data, err := h.CompanyService.DashboardData(ctx, userid.(int64))
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


func (h *CompanyHandler) NewJob(ctx *gin.Context) {

	// we can use an external form or on-site html too
	GoogleFormLink := os.Getenv("NewJobFormLink")

	if GoogleFormLink == "" {
		ctx.File("./template/company/newjobform.html")
		return
	} else {
		ctx.Redirect(http.StatusSeeOther, GoogleFormLink)
		return
	}
}

func (h *CompanyHandler) NewJobPost(ctx *gin.Context) {
	
	// bind incoming request form
	jobdata := new(dto.NewJobData)
	userid, exists := ctx.Get("ID")
	err := ctx.Bind(&jobdata)
	if err != nil || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// call appropriate service
	err = h.CompanyService.NewJobPost(ctx, jobdata, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// TODO: the user can just go back and submit form again which is dangerous
	ctx.Status(http.StatusOK)
}


func (h *CompanyHandler) ApplicantsStatic(ctx *gin.Context) {
	// return the html template
	// the template then indirectly calls /applicantsdata
	ctx.File("./template/company/myapplicants.html")
}

func (h *CompanyHandler) ApplicantsData(ctx *gin.Context) {
	// get jobid and userid and application id off the request
	jobid := ctx.Query("jobid")
	appid := ctx.Query("appid")
	userID, exists := ctx.Get("ID")
	if !exists || jobid == "" || appid == "" { 
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing user id or query parameters",
		})
		return
	}

	// service call
	applicantsData, err := h.CompanyService.ApplicantsData(ctx, userID.(int64), jobid, appid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// respond with 200
	ctx.JSON(http.StatusOK, gin.H{
		"Applicants": applicantsData,
	})
}

func (h *CompanyHandler) GetResumeOrResultFile(ctx *gin.Context) {

	// get applicationid and type ('resume' or 'result') of file 
	applicationid := ctx.Query("applicationid")
	filetype := ctx.Query("type")
	if applicationid == "" || filetype == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "applicationid or type not specified",
		})
		return
	}
	userID := ctx.GetInt64("ID")

	//service call
	filePath, err := h.CompanyService.GetResumeOrResultFilePath(ctx, userID, applicationid, filetype)
	if err != nil {
		ctx. AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// browser heavily caches these files
	// prevent it from doing that by setting this header
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	ctx.File(filePath)
}


func (h *CompanyHandler) JobListingsStatic(ctx *gin.Context) {
	ctx.File("./template/company/myjoblistings.html")
}

func (h *CompanyHandler) JobListingsData(ctx *gin.Context) {
	// get userid off token, to identify company
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "unable to get user ID",
		})
		return
	}

	//service delegation
	jobListings, err := h.CompanyService.JobListings(ctx, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 200OK response
	ctx.JSON(http.StatusOK, gin.H{
		"Listings": jobListings,
	})
}


func (h *CompanyHandler) CloseJob(ctx *gin.Context) {

	// extract jobid and ID from request
	jobid := ctx.Query("jobid")
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user id not found",
		})
		return
	}

	// service delegation
	err := h.CompanyService.CloseJob(ctx, jobid, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to close job: " + err.Error(),
		})
		return
	}

	// 200OK response
	ctx.JSON(http.StatusOK, gin.H{
		"status": "closed job successfully",
	})
}

func (h *CompanyHandler) DeleteJob(ctx *gin.Context) {

	// get jobid and ID off request
	jobid := ctx.Query("jobid")
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user id not found",
		})
		return
	}

	// service delegation
	err := h.CompanyService.DeleteJob(ctx, jobid, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete job: " + err.Error(),
		})
		return
	}

	//200OK response
	ctx.JSON(http.StatusOK, gin.H{
		"status": "delete job successfully",
	})
}


func (h *CompanyHandler) ShortList(ctx *gin.Context) {
	// applicationid off request
	userID, exists := ctx.Get("ID")
	applicationid := ctx.Query("applicationid")
	if applicationid == "" || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing jobid or studentid",
		})
		return
	}
	// service delegation
	err := h.CompanyService.ShortList(ctx, applicationid, userID.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 200OK response
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) Reject(ctx *gin.Context) {
	// get applicationid from request
	userid, exists := ctx.Get("ID")
	applicationid := ctx.Query("applicationid")
	if applicationid == "" || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing jobid or studentid",
		})
		return
	}
	// service delegation
	err := h.CompanyService.Reject(ctx, applicationid, userid.(int64))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) ScheduleInterview(ctx *gin.Context) {

	// bind the POST form for new interview (time, date, etc)
	data :=  new(dto.NewInterview)
	err := ctx.Bind(data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}
	// get userID from token
	userid, exists := ctx.Get("ID")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing ID",
		})
		return
	}
	data.UserId = userid.(int64)

	// service delegation
	errf := h.CompanyService.ScheduleInterview(ctx, data)
	if errf != nil {
		if errf.Type != errs.Internal {
			fmt.Println(errf.Message)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
		}
		return
	}

	// 200OK response
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) Offer(ctx *gin.Context) {
	// get ID, application ID and offer letter from request
	userid, exists := ctx.Get("ID")
	applicationid := ctx.PostForm("OfferApplicationId")
	offerLetter, err := ctx.FormFile("OfferLetter")
	// TODO: validate and check file for size, type, etc
	if err != nil || applicationid == "" || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("missing user id or application id or offer letter : %s", err),
		})
		return
	}
	// service delegation
	err = h.CompanyService.Offer(ctx, userid.(int64), applicationid, offerLetter)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 200OK response
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) CancelInterview(ctx *gin.Context) {
	// get userid and application id from request
	userid, exists := ctx.Get("ID")
	applicationid := ctx.Query("applicationid")
	if applicationid == "" || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user id or application id not found",
		})
		return
	}
	// service delegation
	err := h.CompanyService.CancelInterview(ctx, userid.(int64), applicationid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 200OK response
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) NewTestStatic(ctx *gin.Context) {
	collaboratorEmail := os.Getenv("NewTestGoogleFormsCollaboratorEmail")
	userid := ctx.GetInt64("ID")
	jobidtoBind, err := h.CompanyService.JobListings(ctx, userid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.HTML(200, "newtest.html", gin.H{
		"NewTestGoogleFormsCollaboratorEmail": collaboratorEmail,
		"JobIDToBind": jobidtoBind,
	})
}

func (h *CompanyHandler) NewTestPost(ctx *gin.Context) {
	var newtestData dto.NewTestPost
	// bind the metadata
	err := ctx.Bind(&newtestData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("failed to bind new test data: %s", err),
		})
		return
	}
	// service delegation
	errf := h.CompanyService.NewTestPost(ctx, newtestData)
	if errf.Message != "" {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"type": errf.Type,
			"message": errf.Message,
		})
		return
	}

	// 200OK
	ctx.Status(http.StatusOK)
}


func (h *CompanyHandler) ScheduledStatic(ctx *gin.Context) {
	ctx.File("./template/company/scheduled.html")
}

func (h *CompanyHandler) ScheduledData(ctx *gin.Context) {
	// get userid and eventtype from the request
	userid, exist := ctx.Get("ID")
	eventtype := ctx.Query("eventtype")
	if !exist || eventtype == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing user id or event type",
		})
		return
	}

	// service delegation
	uData, errf := h.CompanyService.ScheduledData(ctx, userid.(int64), eventtype)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}
	// return the data with 200OK
	ctx.JSON(http.StatusOK, uData)
}

func (h *CompanyHandler) UpdateInterview(ctx *gin.Context) {
	// get user id and bind incoming data
	data := new(dto.UpdateInterview)
	userid, exists := ctx.Get("ID")
	err := ctx.Bind(data)
	if err != nil || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"Message": err.Error(),
		})
		fmt.Println(err)
		return
	}
	// service delegation
	err = h.CompanyService.UpdateInterview(ctx, userid.(int64), data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"Message": err.Error(),
		})
		fmt.Println(err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) CompletedStatic(ctx *gin.Context) {
	ctx.File("./template/company/completed.html")
}

func (h *CompanyHandler) CompletedData(ctx *gin.Context) {
	// get user id and event type from request
	userid, exist := ctx.Get("ID")
	eventtype := ctx.Query("eventtype")
	if !exist || eventtype == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing user id or event type",
		})
		return
	}
	// service delegation
	uData, errf := h.CompanyService.CompletedData(ctx, userid.(int64), eventtype)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}
	// respond with data and 200OK
	ctx.JSON(http.StatusOK, uData)
}

func (h *CompanyHandler) EditCutOff(ctx *gin.Context) {
	
	newData := new(dto.UpdateTest)
	// get the user id and bind remaining data
	userid, exists := ctx.Get("ID")
	err := ctx.Bind(newData)
	if err != nil || !exists {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// service delegation
	errf := h.CompanyService.EditCutOff(ctx, userid.(int64), newData)
	if errf != nil {
		if errf.Type != errs.Internal {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
			return
		}
		return
	}
	// 200OK
	ctx.Status(http.StatusOK)
}

func (h *CompanyHandler) PublishTestResults(ctx *gin.Context) {

	testid := ctx.Query("testid")
	userid, exists := ctx.Get("ID")
	if !exists || testid == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Type": errs.MissingRequiredField,
			"Message": "Missing user ID or test ID",
		})
		return 
	}

	errf := h.CompanyService.PublishTestResults(ctx, userid.(int64), testid)
	if errf != nil {
		if errf.Type != errs.Internal {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"Type": errf.Type,
				"Message": errf.Message,
			})
			return
		}
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}













func (h *CompanyHandler) StudentProfileData(ctx *gin.Context) {
	studentid := ctx.Query("id")
	userid, exists := ctx.Get("ID")
	if !exists || studentid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "missing userid or query parameters",
		})
		fmt.Println("error: missing userid or query parameters")
		return
	}

	data, err := h.CompanyService.StudentProfileData(ctx, userid.(int64), studentid)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		fmt.Println(err)
		return
	}
	
	ctx.JSON(http.StatusOK, data)
}	