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
	companyRoute.GET("/getfile", h.GetResumeOrResultFile)
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
	companyRoute.POST("/scheduleinterview", h.ScheduleInterview)
	companyRoute.POST("/cancelinterview", h.CancelInterview)

	companyRoute.GET("/newtest", h.NewTestStatic)
	companyRoute.POST("/newtestpost", h.NewTestPost)

}

func (h *CompanyHandler) CompanyDashboard(ctx *gin.Context) {
	ctx.File("./template/dashboard/companydashboard.html")
}

func (h *CompanyHandler) NewJobGet(ctx *gin.Context) {

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


func (h *CompanyHandler) MyApplicantsStatic(ctx *gin.Context) {
	// return the html template
	// the template then indirectly calls /myapplicants
	ctx.File("./template/company/myapplicants.html")
}

func (h *CompanyHandler) MyApplicants(ctx *gin.Context) {
	// get jobid and userid off the request
	jobid := ctx.Query("jobid")
	userID, exists := ctx.Get("ID")
	if !exists { 
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "user ID not found in token",
		})
		return
	}

	// service call
	applicantsData, err := h.CompanyService.MyApplicants(ctx, userID.(int64), jobid)
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


func (h *CompanyHandler) MyJobListStatic(ctx *gin.Context) {
	ctx.File("./template/company/myjoblistings.html")
}

func (h *CompanyHandler) MyJobListings(ctx *gin.Context) {
	// get userid off token, to identify company
	userid, exists := ctx.Get("ID")
	if !exists || userid == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "unable to get user ID",
		})
		return
	}

	//service delegation
	jobListings, err := h.CompanyService.MyJobListings(ctx, userid.(int64))
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
	var data dto.NewInterview
	err := ctx.Bind(&data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
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
	err = h.CompanyService.ScheduleInterview(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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
	jobidtoBind, err := h.CompanyService.MyJobListings(ctx, userid)
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
