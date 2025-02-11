package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	"go.mod/internal/services"
)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type CompanyHandler struct {
	CompanyService *services.CompanyService
}
func NewCompanyHandler(companyService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		CompanyService: companyService,
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// RegisterRoute initializes all the routes for the company role, see routesDoc.txt for details
func (h *CompanyHandler) RegisterRoute(companyRoute *gin.RouterGroup) {
	// get dashboard template
	companyRoute.GET("/dashboard", h.CompanyDashboard)
	// get dashboard data
	companyRoute.GET("/dashboarddata", h.DashboardData)
	// get the notifications data
	companyRoute.GET("/notifications", h.GetNotifications)


	// get new job posting form
	companyRoute.GET("/newjob", h.NewJob)
	// post new job form
	companyRoute.POST("/newjobpost", h.NewJobPost)

	// get the template for all applicants
	companyRoute.GET("/applicants", h.ApplicantsStatic)
	// get all applicants data
	companyRoute.GET("/applicantsdata", h.ApplicantsData)

	// get any student's file (resume, result)
	companyRoute.GET("/getstudentfile", h.GetResumeOrResultFile)
	// get my job listings template
	companyRoute.GET("/joblistings", h.JobListingsStatic)
	// get my job listings
	companyRoute.GET("/joblistingsdata", h.JobListingsData)
	// close job listing
	companyRoute.GET("/closejob", h.CloseJob)
	// delete job listing
	companyRoute.GET("/deletejob", h.DeleteJob)

	// shortlist given application
	companyRoute.POST("/shortlist", h.ShortList)
	// reject given application
	companyRoute.POST("/reject", h.Reject)
	// offer given application
	companyRoute.POST("/offer", h.Offer)
	// schedule interview for given application
	companyRoute.POST("/scheduleinterview", h.ScheduleInterview)
	// cancel interview for given application
	companyRoute.POST("/cancelinterview", h.CancelInterview)

	// get new test form or template
	companyRoute.GET("/newtest", h.NewTestStatic)
	// post new test data
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

	// get profile template
	companyRoute.GET("/profile", h.GetProfile)
	// get profile data
	companyRoute.GET("/profiledata", h.ProfileData)
	// get any profile file like profile pic, etc
	companyRoute.GET("/getcompanyfile", h.GetFile)
	// post new profile details
	companyRoute.POST("/updatedetails", h.UpdateProfileDetails)
	// post new file
	companyRoute.POST("/updatefile", h.UpdateFile)



	

	// TODO:
	companyRoute.GET("/studentprofiledata", h.StudentProfileData)


}
// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func (h *CompanyHandler) extractUserID(ctx *gin.Context) (int64, *errs.Error) {

	userid, exists := ctx.Get("ID")
	if !exists {
		return 0, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing user ID in request.",
			ToRespondWith: true,
		}
	}

	userID, ok := userid.(int64)
	if !ok {
		return 0, &errs.Error{
			Type: errs.InvalidFormat,
			Message: "User ID of improper format.",
			ToRespondWith: true,
		}
	}

	return userID, nil 
}
// checkFile checks file validity/existence for the given filePath.
// It also does ctx.Set(error) and returns a structured *errs.Error object too for any errors
func (h *CompanyHandler) checkFile(ctx *gin.Context, filePath string) *errs.Error {

	if filePath == "" {
		ctx.Set("error", "filePath is empty string.")
		return &errs.Error{
			Type: errs.NotFound,
			Message: "filePath is empty.",
		}
	}

	_, err := os.Stat(filePath)
	if err != nil {
		ctx.Set("error", "cannot get file data for path : " + filePath)
		return &errs.Error{
			Type: errs.NotFound,
			Message: "Failed to get metadata/ access file : " + filePath,
		}
	}

	return nil
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// CompanyDashboard returns the dashboard template for the company role
func (h *CompanyHandler) CompanyDashboard(ctx *gin.Context) {

	filePath := config.Paths.CompanyDashboardPath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// DashboardData returns the dashboard data for the company role
func (h *CompanyHandler) DashboardData(ctx *gin.Context) {

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	data, errf := h.CompanyService.DashboardData(ctx, userID)
	if errf != nil {
		ctx.Set("error", errf.Message)
		return
	}

	ctx.JSON(http.StatusOK, data)
}
// GetNotifications returns the notifications for user ID, uses start and end as params for limits
func (h *CompanyHandler) GetNotifications(ctx *gin.Context) {

	start := ctx.Query("start")
	end := ctx.Query("end")
	if start == "" || end == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.InvalidFormat,
			Message: "Missing start or end parameter in request url.",
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	notifs, errf := h.CompanyService.Notify.GetNotifications(ctx, userID, start, end)
	if errf != nil {
		ctx.Set("error", errf.Message)
		return
	}

	ctx.JSON(http.StatusOK, notifs)
}
// NewJob returns the form or template for posting a new job
func (h *CompanyHandler) NewJob(ctx *gin.Context) {

	GoogleFormLink := os.Getenv("NewJobFormLink")

	if GoogleFormLink == "" {
		filePath := config.Paths.NewJobFormTemplatePath
		errf := h.checkFile(ctx, filePath)
		if errf != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.File(filePath)
	} else {
		ctx.Redirect(http.StatusSeeOther, GoogleFormLink)
	}
}
// NewJobPost takes a post request and creates a new job
func (h *CompanyHandler) NewJobPost(ctx *gin.Context) {
	
	jobdata := new(dto.NewJobData)

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.NewJobPost(ctx, jobdata, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}
	
	// TODO: the user can just go back and submit form again which is dangerous
	ctx.Status(http.StatusOK)
}
// ApplicantsStatic returns the MyApplicants template for company role
func (h *CompanyHandler) ApplicantsStatic(ctx *gin.Context) {
	filePath := config.Paths.CompanyMyApplicantsTemplatePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// ApplicantsData retrieves applicant records based on the specified criteria:
// - Returns all applicants if jobid and appid == 0.
// - Returns applicants for a specific job if a job ID is provided and appid == 0.
// - Returns details of a specific application if an application ID is provided and jobid == 0.
func (h *CompanyHandler) ApplicantsData(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	appid := ctx.Query("appid")
	if jobid == "" || appid == "" { 
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing required fields in request url.",
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	applicantsData, errf := h.CompanyService.ApplicantsData(ctx, userID, jobid, appid)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Applicants": applicantsData,
	})
}
// GetResumeOrResultFile is specifically used to get student files like resume and result and cover letter in upcoming versions
func (h *CompanyHandler) GetResumeOrResultFile(ctx *gin.Context) {

	// get applicationid and type ('resume' or 'result') of file 
	applicationid := ctx.Query("applicationid")
	filetype := ctx.Query("type")
	if applicationid == "" || filetype == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing required fields in request url.",
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	filePath, errf := h.CompanyService.GetResumeOrResultFilePath(ctx, userID, applicationid, filetype)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	// browser heavily caches these files
	// prevent it from doing that by setting this header
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	ctx.File(filePath)
}
// JobListingsStatic returns the JobListings page for the company role
func (h *CompanyHandler) JobListingsStatic(ctx *gin.Context) {

	filePath := config.Paths.CompanyMyJobListingsTemplatePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// JobListingsData gets the JobListings data for the company role
func (h *CompanyHandler) JobListingsData(ctx *gin.Context) {

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	jobListings, errf := h.CompanyService.JobListings(ctx, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"Listings": jobListings,
	})
}
// CloseJob marks the job listing active_status to false, effectively closing the listing for further applicants
func (h *CompanyHandler) CloseJob(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	if jobid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing job ID parameter in request url.",
			ToRespondWith: true, 
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.CloseJob(ctx, jobid, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Closed job successfully.",
	})
}
// DeleteJob directly deletes the requested job, job will be archived instead of deletion in upcoming revisions. 
func (h *CompanyHandler) DeleteJob(ctx *gin.Context) {

	jobid := ctx.Query("jobid")
	if jobid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing job ID parameter in request url.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.DeleteJob(ctx, jobid, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Deleted job successfully.",
	})
}
// ShortList changes the application status to 'ShortListed', the status has to be 'Under Review'
func (h *CompanyHandler) ShortList(ctx *gin.Context) {

	applicationid := ctx.Query("applicationid")
	if applicationid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing application ID parameter in request url.", 
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.ShortList(ctx, applicationid, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Application shortlisted successfully",
	})
}
// Reject changes the application status to 'Rejected', also changes interview status to 'Completed'.
func (h *CompanyHandler) Reject(ctx *gin.Context) {

	applicationid := ctx.Query("applicationid")
	if applicationid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing application ID parameter in request url.", 
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.Reject(ctx, applicationid, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Application rejected successfully",
	})
}
// ScheduleInterview schedules a new interview from the submitted form, uses dto.NewInterview
func (h *CompanyHandler) ScheduleInterview(ctx *gin.Context) {

	data :=  new(dto.NewInterview)

	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "Invalid format of submitted form.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}
	data.UserId = userID

	errf = h.CompanyService.ScheduleInterview(ctx, data)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Interview scheduled successfully.",
	})
}
// Offer changes the application status to 'Offered', sends an email with offer letter, updates interview status to 'Completed'.
func (h *CompanyHandler) Offer(ctx *gin.Context) {

	applicationid := ctx.PostForm("OfferApplicationId")
	offerLetter, err := ctx.FormFile("OfferLetter")

	// TODO: validate and check file for size, type, etc
	if err != nil || applicationid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing application ID or Offer Letter.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.Offer(ctx, userID, applicationid, offerLetter)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Application Offered successfully.",
	})
}
// CancelInterview cancels the interview for given application ID, sends an email to student
func (h *CompanyHandler) CancelInterview(ctx *gin.Context) {

	applicationid := ctx.Query("applicationid")
	if applicationid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing application ID in request url.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.CancelInterview(ctx, userID, applicationid)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Cancelled interview successfully",
	})
}
// NewTestStatic returns the New Test template form, initial data is embedded here
func (h *CompanyHandler) NewTestStatic(ctx *gin.Context) {

	collaboratorEmail, exists := os.LookupEnv("NewTestGoogleFormsCollaboratorEmail")
	if !exists || collaboratorEmail == "" {
		ctx.JSON(http.StatusInternalServerError, errs.Error{
			Type: errs.NotFound,
			Message: "Test Collaborator Email Not found in enviornment variables.",
			ToRespondWith: true,
		})
		ctx.Set("error", "Test Collaborator Email Not found in enviornment variables.")
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	jobidtoBind, errf := h.CompanyService.JobListings(ctx, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.HTML(http.StatusOK, config.Paths.NewTestFormTemplateName, gin.H{
		"NewTestGoogleFormsCollaboratorEmail": collaboratorEmail,
		"JobIDToBind": jobidtoBind,
	})
}
// NewTestPost posts a new test, uses dto.NewTestPost
func (h *CompanyHandler) NewTestPost(ctx *gin.Context) {
	
	newtestData := new(dto.NewTestPost)
	err := ctx.Bind(&newtestData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "Invalid format of submitted form.",
			ToRespondWith: true,
		})
		// TODO: maybe also send this error on the channel
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.NewTestPost(ctx, userID, newtestData)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "New test posted successfully.",
	})
}
// ScheduledStatic responds with the 'Scheduled' page for company role
func (h *CompanyHandler) ScheduledStatic(ctx *gin.Context) {

	filePath := config.Paths.CompanyScheduledTemplatePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// ScheduledData returns 'Scheduled' events for company role, needs 'eventtype' as param
func (h *CompanyHandler) ScheduledData(ctx *gin.Context) {

	eventtype := ctx.Query("eventtype")
	if eventtype == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing event type in request url.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	uData, errf := h.CompanyService.ScheduledData(ctx, userID, eventtype)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, uData)
}
// UpdateInterview updates the interview details for given interview_ID, uses dto.UpdateInterview
func (h *CompanyHandler) UpdateInterview(ctx *gin.Context) {
	
	data := new(dto.UpdateInterview)
	err := ctx.Bind(data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "Update Interview form data is incomplete or invalid.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.UpdateInterview(ctx, userID, data)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Interview details updated successfully.",
	})
}
// CompletedStatic returns the 'Completed' page for company role
func (h *CompanyHandler) CompletedStatic(ctx *gin.Context) {

	filePath := config.Paths.CompanyCompletedTemplatePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// CompletedData returns the 'Completed' events for company role
func (h *CompanyHandler) CompletedData(ctx *gin.Context) {
	
	eventtype := ctx.Query("eventtype")
	if eventtype == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing event type in request url.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	uData, errf := h.CompanyService.CompletedData(ctx, userID, eventtype)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, uData)
}
// EditCutOff edits test cutoff threshold, and then starts the 'Draft Result' process 
// this can later be used to edit more details of the test
func (h *CompanyHandler) EditCutOff(ctx *gin.Context) {
	
	newData := new(dto.UpdateTest)
	err := ctx.Bind(newData)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "Edit cutoff form data is incomplete or invalid.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.EditCutOff(ctx, userID, newData)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Cutoff has been updated successfully.",
	})
}
// PublishTestResults starts the process of publishing test results for given test_id, returns immediately
func (h *CompanyHandler) PublishTestResults(ctx *gin.Context) {

	testid := ctx.Query("testid")
	if testid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing test ID in request url.",
			ToRespondWith: true,
		})
		return 
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.PublishTestResults(ctx, userID, testid)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Started publishing results.",
	})
}
// GetProfile returns the 'MyProfile' page for company role
func (h *CompanyHandler) GetProfile(ctx *gin.Context) {

	filePath := config.Paths.CompanyMyProfileTemplatePath
	errf := h.checkFile(ctx, filePath)
	if errf != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.File(filePath)
}
// ProfileData returns the data for the profile page, all at once
func (h *CompanyHandler) ProfileData(ctx *gin.Context) {

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	data, errf := h.CompanyService.ProfileData(ctx, userID)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, data)
}
// GetFile returns the user's file for the requested type
func (h *CompanyHandler) GetFile(ctx *gin.Context) {

	fileType := ctx.Query("type")
	if fileType == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing file type parameter in request url.", 
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	filePath, errf := h.CompanyService.GetCompanyFile(ctx, userID, fileType)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return 
	}

	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	ctx.File(filePath)
}
// UpdateProfileDetails updates the profile details for the user, uses dto.UpdateCompanyDetails
func (h *CompanyHandler) UpdateProfileDetails(ctx *gin.Context) {
	
	details := new(dto.UpdateCompanyDetails)
	err := ctx.Bind(details)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.IncompleteForm,
			Message: "New details form is incomplete or invalid.",
			ToRespondWith: true,
		})
		return
	}	

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.UpdateProfileDetails(ctx, userID, details)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Profile details updated successfully.",
	})		
}
// UpdateFile updates the given file, for the given file type, for the user
func (h *CompanyHandler) UpdateFile(ctx *gin.Context) {
	
	fileType := ctx.Query("type")
	if fileType == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing file type in request url.", 
			ToRespondWith: true,
		})
		return
	}
	
	file, err := ctx.FormFile(fileType)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing file in request.",
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	errf = h.CompanyService.UpdateFile(ctx, userID, file, fileType)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "File updated successfully.",
	})
}





// CLEANUP UNDERWAY >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>










func (h *CompanyHandler) StudentProfileData(ctx *gin.Context) {
	
	studentid := ctx.Query("id")
	if studentid == "" {
		ctx.JSON(http.StatusBadRequest, errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing student ID parameter in request url.", 
			ToRespondWith: true,
		})
		return
	}

	userID, errf := h.extractUserID(ctx)
	if errf != nil {
		ctx.JSON(http.StatusBadRequest, errf)
		return
	}

	data, errf := h.CompanyService.StudentProfileData(ctx, userID, studentid)
	if errf != nil {
		if errf.ToRespondWith {
			ctx.JSON(http.StatusBadRequest, errf)
		} else {
			ctx.Set("error", errf.Message)
		}
		return
	}
	
	ctx.JSON(http.StatusOK, data)
}	


