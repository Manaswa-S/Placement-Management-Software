package services

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	gocharts "go.mod/internal/go-charts"
	"go.mod/internal/notify"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
)


type CompanyService struct {
	queries *sqlc.Queries
	GAPIService *apicalls.Caller
	RedisClient *redis.Client
	Notify *notify.Notify
}

func NewCompanyService(queriespool *sqlc.Queries, gapiService *apicalls.Caller, redisClient *redis.Client, notifyService *notify.Notify) *CompanyService {
	return &CompanyService{
		queries: queriespool,
		GAPIService: gapiService,
		RedisClient: redisClient,
		Notify: notifyService,
	}
}

func (c *CompanyService) DashboardData(ctx *gin.Context, userID int64) (*sqlc.CompanyDashboardDataRow, *errs.Error) {

	data, err := c.queries.CompanyDashboardData(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	return &data, nil
}

func (c *CompanyService) NewJobPost(ctx *gin.Context, jobdata *dto.NewJobData, userID int64) (*errs.Error) {	
	// split skills into []text
	skills := strings.Split(jobdata.SkillsRequired, ",")
	for i, skill := range skills {
		// trim off spaces
		skills[i] = strings.TrimSpace(skill)
	}

	// create map of extra params // flexiblity
	extras := make(map[string]interface{})
	for key, values := range ctx.Request.Form {
		if _, exists := map[string]bool{
			"JobId": true,
			"JobLocation": true,
			"JobTitle": true,
			"JobDescription": true,
			"JobType": true,
			"JobSalary": true,
			"SkillsRequired": true,
			"JobPosition": true,
		}[key]; !exists {
			if len(values) > 0 {
				extras[key] = values[0]
			}
		}
	}

	extraJson, err := json.Marshal(extras)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to marshal extra params : " + err.Error(),
		}
	}
	// TODO: also used to update the job, can be a vulnerability
	if (jobdata.JobId != 0) {
		err = c.queries.UpdateJob(ctx, sqlc.UpdateJobParams{
			Location: jobdata.JobLocation,
			Title: jobdata.JobTitle,
			Description: pgtype.Text{String: jobdata.JobDescription, Valid: true},
			Type: jobdata.JobType,
			Salary: jobdata.JobSalary,
			Skills: skills,
			Position: jobdata.JobPosition,
			Extras: extraJson,
			JobID: jobdata.JobId,
			UserID: userID,
		})
		if err != nil {
			if err.Error() == errs.NoRowsMatch {
				return &errs.Error{
					Type: errs.Unauthorized,
					Message: "You are not allowed to alter this job, or it does not exist.",
					ToRespondWith: true,
				}
			}
			return &errs.Error{
				Type: errs.Internal,
				Message: "Failed to update job listing : " + err.Error(),
			}
		}

		return nil
	}

	// TODO: need to better validate incoming data 
	// add job data to db
	err = c.queries.InsertNewJob(ctx, sqlc.InsertNewJobParams{
		DataUrl: pgtype.Text{String: "", Valid: true},
		UserID: userID,
		Title: jobdata.JobTitle,
		Location: jobdata.JobLocation,
		Type: jobdata.JobType,
		Salary: jobdata.JobSalary,
		Skills: skills,
		Position: jobdata.JobPosition,
		Extras: extraJson,
		Description: pgtype.Text{String: jobdata.JobDescription, Valid: true},
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to insert new job data : " + err.Error(),
		}
	}

	return nil
}

func (c *CompanyService) ApplicantsData(ctx *gin.Context, userID int64, jobid string, appid string) (*[]sqlc.GetApplicantsRow, *errs.Error){


	// parse jobid to int64
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse job ID string to int : " + err.Error(),
		}
	}
	appID, err := strconv.ParseInt(appid, 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse application ID string to int : " + err.Error(),
		}
	}
	
	// db call
	applicantsData, err := c.queries.GetApplicants(ctx, sqlc.GetApplicantsParams{
		UserID: userID,
		JobID: jobID,
		ApplicationID: appID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return nil, &errs.Error{
				Type: errs.NotFound,
				Message: "No applications found for the given user ID, job ID and application ID",
				ToRespondWith: true,
			}	
		}
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse job ID string to int : " + err.Error(),
		}
	}

	// return 
	return &applicantsData, nil
}

func (c *CompanyService) GetResumeOrResultFilePath(ctx *gin.Context, userID int64, applicationid string, filetype string) (string, *errs.Error) {
	// parse applicationid to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return "", &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Invalid application ID, failed to parse to int : " + err.Error(),
		}
	}

	// get both file paths (resume and result)
	filePaths, err := c.queries.GetResumeAndResultPath(ctx, sqlc.GetResumeAndResultPathParams{
		UserID: userID,
		ApplicationID: applicationId,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return "", &errs.Error{
				Type: errs.Unauthorized,
				Message: "The given user ID is not authorized to view the requested file.",
				ToRespondWith: true,
			}
		}
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "Could not get resume and result path : " + err.Error(),
		}
	}

	// check what type is requested
	filepath := filePaths.ResultUrl
	if filetype == "resume" {
		filepath = filePaths.ResumeUrl.String
	}

    // Check if the file exists
    if _, err := os.Stat(filepath); err != nil {
        return "", &errs.Error{
			Type: errs.Internal,
			Message: "File not found at path : " + filepath + " : " + err.Error(),
		}
    }

	// by default updates the application status to UnderReview from Applied
	_, err = c.queries.ApplicationStatusToAnd(ctx, sqlc.ApplicationStatusToAndParams{
		Status: "UnderReview",
		ApplicationID: applicationId,
		Status_2: "Applied",
	})
	if err != nil {
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "Failed to update application status : " + err.Error(),
		}
	}

	// return file path 
	return filepath, nil
}

func (c *CompanyService) JobListings(ctx *gin.Context, userID int64) (*[]sqlc.GetJobListingsRow, *errs.Error){

	jobListings, err := c.queries.GetJobListings(ctx, userID)
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return nil, &errs.Error{
				Type: errs.Unauthorized,
				Message: "The given user ID is not authorized to view requested data.",
				ToRespondWith: true,
			}
		}
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get job listings data for company : " + err.Error(),
		}
	}

	return &jobListings, nil
}

func (c *CompanyService) CloseJob(ctx *gin.Context, jobid string, userID int64) (*errs.Error){

	// parse jobid from string to int64
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Invalid job ID, failed to parse to int : " + err.Error(),
		}
	}

	// db query to change jobs.active_status to false
	err = c.queries.CloseJob(ctx, sqlc.CloseJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return &errs.Error{
				Type: errs.Unauthorized,
				Message: "The given user ID is not authorized to change status of requested job.",
				ToRespondWith: true,
			}	
		}
		return &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("Failed to close job for job ID : %d : %v", jobID, err.Error()),
		}
	}

	return nil
}

func (c *CompanyService) DeleteJob(ctx *gin.Context, jobid string, userID int64) (*errs.Error){

	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Invalid job ID, failed to parse to int : " + err.Error(),
		}
	}

	err = c.queries.DeleteJob(ctx, sqlc.DeleteJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return &errs.Error{
				Type: errs.Unauthorized,
				Message: "The given user ID is not authorized to change status of requested job.",
				ToRespondWith: true,
			}	
		}
		return &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("Failed to delete job for job ID : %d : %v", jobID, err.Error()),
		}
	}

	return nil
}

func (c *CompanyService) ShortList(ctx *gin.Context, applicationid string, userID int64) (*errs.Error){

	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Invalid applications ID, failed to parse to int : " + err.Error(),
		}
	}

	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get job owner user ID : " + err.Error(),
		}
	}

	if userID != jobOwnerID {
		return &errs.Error{
			Type: errs.Unauthorized,
			Message: "The given user ID is not authorized to access requested application.",
			ToRespondWith: true,
		}
	}

	studentUserID, err := c.queries.ApplicationStatusToAnd(ctx, sqlc.ApplicationStatusToAndParams{
		Status: "ShortListed",
		ApplicationID: applicationId,	
		Status_2: "UnderReview",
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to change application status : " + err.Error(),
		}
	}

	errf := c.Notify.NewNotification(ctx, studentUserID, &dto.NotificationData{
		Title: "Application Shortlisted",
		Description: fmt.Sprintf("Your application (ID: %s) has been shortlisted.", applicationid),
	})
	if errf != nil {
		return errf
	}

	return nil
}

func (c *CompanyService) Reject(ctx *gin.Context, applicationid string, userID int64) (*errs.Error){

	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Invalid applications ID, failed to parse to int : " + err.Error(),
		}
	}

	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get job owner user ID : " + err.Error(),
		}
	}

	if userID != jobOwnerID {
		return &errs.Error{
			Type: errs.Unauthorized,
			Message: "The given user ID is not authorized to access requested application.",
			ToRespondWith: true,
		} 
	}

	studentUserID, err := c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		Status: "Rejected",
		ApplicationID: applicationId,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to change application status : " + err.Error(),
		}
	}

	err = c.queries.InterviewStatusTo(ctx, sqlc.InterviewStatusToParams{
		Status: "Completed",
		ApplicationID: applicationId,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to change interview status : " + err.Error(),
		}
	}

	errf := c.Notify.NewNotification(ctx, studentUserID, &dto.NotificationData{
		Title: "Application Rejected",
		Description: fmt.Sprintf("Your application (ID: %s) has been Rejected.", applicationid),
	})
	if errf != nil {
		return errf
	}
	
	return nil
}

func (c *CompanyService) ScheduleInterview(ctx *gin.Context, data *dto.NewInterview) (*errs.Error) {

	if (data.DateTime.Compare(time.Now()) != 1) {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "Interview Date-Time cannot be in the past.",
			ToRespondWith: true,
		}
	}

	dt, err := c.queries.ScheduleInterview(ctx, sqlc.ScheduleInterviewParams{
		ApplicationID: data.ApplicationId,
		UserID: data.UserId,
		DateTime: pgtype.Timestamptz{Time: data.DateTime, Valid: true},
		Type: data.Type,
		Notes: pgtype.Text{String: data.Notes, Valid: true},
		Location: data.Location,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to insert new interview in db : " + err.Error(),
		}
	}

	// student name and email and job title and company name for email template
	studentData, err := c.queries.GetScheduleInterviewData(ctx, data.ApplicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get student data for email : " + err.Error(),
		}
	}

	data.StudentName = studentData.StudentName
	data.CompanyName = studentData.CompanyName
	data.JobTitle = studentData.Title
	data.DT = dt
	
	// execute email template
	template, err := utils.DynamicHTML("./template/emails/interviewScheduled.html", data)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get dynamic template for new interview email : " + err.Error(),
		}
	}
	// send new interview email to student
	go utils.SendEmailHTML(template, []string{studentData.StudentEmail})

	errf := c.Notify.NewNotification(ctx, studentData.UserID, &dto.NotificationData{
		Title: "Interview Scheduled",
		Description: fmt.Sprintf("New Interview scheduled for application (ID: %d).", data.ApplicationId),
	})
	if errf != nil {
		return errf
	}

	// no error
	return nil
}

func (c *CompanyService) Offer(ctx *gin.Context, userID int64, applicationid string, offerLetter *multipart.FileHeader) (*errs.Error) {

	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Failed to parse application ID : " + err.Error(),
		}
	}

	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get job owner user ID : " + err.Error(),
		}
	}

	if userID != jobOwnerID {
		return &errs.Error{
			Type: errs.Unauthorized,
			Message: "The given user ID is not authorized to access requested application.",
			ToRespondWith: true,
		} 
	}

	// TODO: atomicity problem 
	// update interview status to 'Completed'
	err = c.queries.InterviewStatusTo(ctx, sqlc.InterviewStatusToParams{
		ApplicationID: applicationId,
		Status: "Completed",
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to change interview status : " + err.Error(),
		}
	}

	studentUserID, err := c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		ApplicationID: applicationId,
		Status: "Offered",
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to update application status : " + err.Error(),
		}
	}

	offerData, err := c.queries.GetOfferLetterData(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to fetch offer letter data for email : " + err.Error(),
		}
	}

	template, err := utils.DynamicHTML("./template/emails/offerEmail.html", offerData)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get dynamic template for offer email : " + err.Error(),
		}
	}
	go utils.SendEmailHTMLWithAttachmentFileHeader(template, []string{offerData.StudentEmail}, offerLetter)

	errf := c.Notify.NewNotification(ctx, studentUserID, &dto.NotificationData{
		Title: "Offered !!",
		Description: fmt.Sprintf("Congratulations! New job offer received. (ID: %s)", applicationid),
	})
	if errf != nil {
		return errf
	}

	// no error
	return nil
}

func (c *CompanyService) CancelInterview(ctx *gin.Context, userID int64, applicationid string) (*errs.Error) {

	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.InvalidFormat,
			Message: "Failed to parse application ID : " + err.Error(),
		}	}

	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {	
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get job owner user ID : " + err.Error(),
		}	
	}
	if userID != jobOwnerID {
		return &errs.Error{
			Type: errs.Unauthorized,
			Message: "The given user ID is not authorized to access requested application.",
			ToRespondWith: true,
		} 
	}

	data, err := c.queries.CancelInterviewEmailData(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get data for email : " + err.Error(),
		}	
	}

	newData := dto.CancelInterview{
		StudentName: data.StudentName,
		StudentEmail: data.StudentEmail,
		JobTitle: data.Title,
		CompanyName: data.CompanyName,
		DateTime: data.DateTime,
		RepresentativeEmail: data.RepresentativeEmail,
		RepresentativeName: data.RepresentativeName,
	}
	template, err := utils.DynamicHTML("./template/emails/interviewCancelled.html", newData)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get dynamic template for offer email : " + err.Error(),
		}
	}
	go utils.SendEmailHTML(template, []string{data.StudentEmail})

	// TODO: dont delete interview, make it cancelled
	err = c.queries.DeleteInterview(ctx, applicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to delete interview : " + err.Error(),
		}	
	}

	return nil
}

const (
	GForm = "GForms"
	CSVJSON = "CSVJSON"
	Manual = "Manual"
)

func (c *CompanyService) NewTestPost(ctx *gin.Context, userID int64, newtestData *dto.NewTestPost) (*errs.Error) {

	var formID string
	var errf *errs.Error

	switch newtestData.UploadMethod {
	case GForm :
		gformData := new(dto.NewTestGForms)
		err := ctx.Bind(gformData)
		if err != nil {
			return &errs.Error{
				Type: errs.IncompleteForm,
				Message: "Failed to bind GForm : " + err.Error(),
			}
		}
		formID, errf = c.NewTestPostGForm(ctx, gformData)
		if errf != nil {
			return errf
		}
	case CSVJSON:
		// TODO: 
	case Manual:
		// TODO:
	default:
	}	

	err := c.queries.NewTest(ctx, sqlc.NewTestParams{
		TestName: newtestData.Name,
		Description: pgtype.Text{String: newtestData.Description, Valid: true},
		Duration: newtestData.Duration,
		QCount: newtestData.QuestionCount,
		EndTime: pgtype.Timestamptz{Time: newtestData.EndDateTime, Valid: true},
		Type: newtestData.Type,
		UploadMethod: newtestData.UploadMethod,
		JobID: pgtype.Int8{Int64: newtestData.BindedJobId, Valid: true},
		UserID: userID,
		FileID: formID,
		Threshold: int32(newtestData.Threshold),
	})
	if err != nil {
		var pgerr *pgconn.PgError
		// TODO: this can also fail, have a fallback (else statement)
		if errors.As(err, &pgerr) {
			if pgerr.Code == errs.UniqueViolation {
				return &errs.Error{
					Type: errs.UniqueViolation,
					Message: "The test already exists ! You cannot create multiple tests with the same test file.",
					ToRespondWith: true,
				}
			} 
			return &errs.Error{
				Type: errs.Internal,
				Message: "Failed to insert new test in db : " + err.Error(),
			}
		}
	}

	allEmails, err := c.queries.GetAllApplicantsEmailsForJob(ctx, newtestData.BindedJobId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get all emails of applicants for job to send new test email to : " + err.Error(),
		}
	} else {

		jobDetails, err := c.queries.GetJobDetails(ctx, newtestData.BindedJobId)
		if err != nil {
			return &errs.Error{
				Type: errs.Internal,
				Message: "Failed to get job details : " + err.Error(),
			}
		} else {

			newtestData.JobTitle = jobDetails.Title
			newtestData.CompanyName = jobDetails.CompanyName
			newtestData.FormattedEndDate = newtestData.EndDateTime.Format("2006-01-02")
			newtestData.FormattedEndTime = newtestData.EndDateTime.Format("15:04")

			template, err := utils.DynamicHTML("./template/emails/newTestEmail.html", newtestData)
			if err != nil {
				return &errs.Error{
					Type: errs.Internal,
					Message: "Failed to generate template for new test email : " + err.Error(),
				}
			} else {
				go utils.SendEmailHTML(template, allEmails)
			}
		}
	}

	return errf
}

func (c *CompanyService) NewTestPostGForm(ctx *gin.Context, gformData *dto.NewTestGForms) (string, *errs.Error) {

	// get the raw responders link from the link provided by the user 
	paramIndex := strings.Index(gformData.ResponderLink, "?")
	if paramIndex != -1 {
		gformData.ResponderLink = gformData.ResponderLink[:paramIndex]
	}
	// check if we already have the form metadata in the Redis Cache
	exists, err := c.RedisClient.Exists(ctx, gformData.ResponderLink).Result()
	// if not
	if exists == 0 {
		// call drive change api to get file changes from the last start token
		newList, err := c.GAPIService.DriveChanges()
		if err != nil {
			return "", &errs.Error{
				Type: errs.Internal,
				Message: "Failed to get GDrive changes : " + err.Error(),
			}
		}
		changes := newList.Changes

		// loop over every changed object
		for _, change := range changes {
			// check if it is not null and has a mimetype of apps.form
			if change.File != nil && change.File.MimeType == "application/vnd.google-apps.form" {
				// get the metadata from the forms api
				formData, err := c.GAPIService.GetFormMetadata(change.FileId)
				if err != nil {
					return "", &errs.Error{
						Type: errs.Internal,
						Message: "Failed to get GForm metadata : " + err.Error(),
					}
				}
				// Set the key:value in the Redis Cache {responderUri : formId}
				err = c.RedisClient.Set(ctx, formData.ResponderUri, formData.FormId, 0).Err()
				if err != nil {
					return "", &errs.Error{
						Type: errs.Internal,
						Message: "Failed to insert into redis : " + err.Error(),
					}
				}

				// Do this for every form 
				// We do not consider if we have already found our result
				// We continue for every updated form to keep Cache up to date
				// This can be changed later
				//TODO:
			}
		}
	} else if err != nil {
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get from redis : " + err.Error(),
		}
	}

	formID, err := c.RedisClient.Get(ctx, gformData.ResponderLink).Result()
	if err == redis.Nil {
		// the result still does not exist
		// the user has not provided you with the access
		return "", &errs.Error{
			Type: errs.IncompleteAction,
			Message: "The 'Editor Access' has not been shared with the given collaborator email.",
			ToRespondWith: true,
		}
	} else if err != nil {
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get from redis : " + err.Error(),
		}	
	}
	
	return formID, nil
}

func (c *CompanyService) ScheduledData(ctx *gin.Context, userID int64, eventtype string) (*dto.Upcoming, *errs.Error) {
	// switch between event types
	switch eventtype {
	case "interviews":
		uInts, err := c.queries.ScheduledInterviewsCompany(ctx, userID)
		if err != nil {
			return nil, &errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
			}
		}
		return &dto.Upcoming{
			Data: uInts,
		}, nil
	case "tests":
		uTests, err := c.queries.ScheduledTestsCompany(ctx, userID)
		if err != nil {
			return nil, &errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
			}
		}
		return &dto.Upcoming{
			Data: uTests,
		}, nil
	default:
		return nil, &errs.Error{
			Type: errs.NotFound,
			Message: "No such event type exists",
		}
	}
}


func (c *CompanyService) CompletedData(ctx *gin.Context, userID int64, eventtype string) (*dto.Completed, *errs.Error) {
	// switch betweem event type
	switch eventtype {
	case "interviews":
		uInts, err := c.queries.CompletedInterviewsCompany(ctx, userID)
		if err != nil {
			return nil, &errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
			}
		}
		return &dto.Completed{
			Data: uInts,
		}, nil
	case "tests":
		uTests, err := c.queries.CompletedTestsCompany(ctx, userID)
		if err != nil {
			return nil, &errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
			}
		}
		return &dto.Completed{
			Data: uTests,
		}, nil
	default:
		return nil, &errs.Error{
			Type: errs.NotFound,
			Message: "Invalid argument, no such event type exists",
		}
	}
}


func (c *CompanyService) UpdateInterview(ctx *gin.Context, userID int64, data *dto.UpdateInterview) (*errs.Error) {

	newData, err := c.queries.UpdateInterview(ctx, sqlc.UpdateInterviewParams{
		UserID: userID,
		InterviewID: data.InterviewID,
		DateTime: pgtype.Timestamptz{Time: data.DateTime, Valid: true},
		Type: data.Type,
		Notes: pgtype.Text{String: data.Notes, Valid: true},
		Location: data.Location,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return &errs.Error{
				Type: errs.Unauthorized,
				Message: "An interview for the given interview_ID and user ID was not found.",
				ToRespondWith: true,
			}
		}
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to update interview details : " + err.Error(),
		}
	}

	stdData, err := c.queries.GetScheduleInterviewData(ctx, newData.ApplicationID)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get student data for interview-updated email : " + err.Error(),
		}
	}
	data.StudentName = stdData.StudentName
	data.CompanyName = stdData.CompanyName
	data.JobTitle = stdData.Title
	data.DT = newData.DateTime

	// execute email template
	template, err := utils.DynamicHTML("./template/emails/interviewRescheduled.html", data)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to generate template for interview-updated email : " + err.Error(),
		}
	}
	// send new interview email to student
	go utils.SendEmailHTML(template, []string{stdData.StudentEmail})

	return nil
}





func (c *CompanyService) EditCutOff(ctx *gin.Context, userID int64, newData *dto.UpdateTest) (*errs.Error) {

	if newData.Threshold > 99 || newData.Threshold < 1 {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "Cutoff threshold must be between 1 and 99.",
		}
	}

	published, err := c.queries.IsTestPublished(ctx, newData.TestID)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}
	if published {
		return &errs.Error{
			Type: errs.CheckViolation,
			Message: "This test is already published. Cannot edit now.",
		}
	}

	// update the new thresholds in the db
	err = c.queries.UpdateTest(ctx, sqlc.UpdateTestParams{
		TestID: newData.TestID,
		UserID: userID,
		Threshold: pgtype.Int4{Int32: int32(newData.Threshold), Valid: true},
		Published: pgtype.Bool{Valid: false},
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	// call utils to generate the cumulative result draft
	_, err = utils.GenerateTestResultDraft(c.queries, c.GAPIService, newData.TestID)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	// all ok
	return nil
}


func (c *CompanyService) PublishTestResults(ctx *gin.Context, userID int64, testid string) (*errs.Error) {

	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("unable to parse test id : %v", err),
		}
	}

	_, err = c.queries.TestAuthorization(ctx, sqlc.TestAuthorizationParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return &errs.Error{
				Type: errs.Unauthorized,
				Message: "You are not authorized to operate on this Test. This test belongs to a different user. Or the test does not exist.",
			}
		}
	}

	published, err := c.queries.IsTestPublished(ctx, testID)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}
	if published {
		return &errs.Error{
			Type: errs.CheckViolation,
			Message: "This test is already published. Cannot publish again.",
		}
	}

	go func() {
		err = utils.PublishTestResults(c.queries, c.GAPIService, testID)
		if err != nil {
			fmt.Println(err)
		} else {
			err := c.queries.UpdateTest(ctx, sqlc.UpdateTestParams{
				TestID: testID,
				UserID: userID,
				Threshold: pgtype.Int4{Valid: false},
				Published: pgtype.Bool{Bool: true, Valid: true},
			})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("Test results for %d have been published.", testID)
		}
	} ()

	return nil
}


func (s *CompanyService) ProfileData(ctx *gin.Context, userID int64) (*dto.CompanyProfileData, *errs.Error) {
	

	// alljobIds, err := s.queries.GetAllJobsID(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }


	overData, err := s.queries.ApplicantsCount(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}
	var sankeyCharts []*charts.Sankey
	for _, o := range overData {
		sankeyChrt := gocharts.SankeyApplicants(&o)
		sankeyCharts = append(sankeyCharts, sankeyChrt)
	}


	// total jobs posted >  offered
	
	userData, err := s.queries.UsersTableData(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}	}

	data, err := s.queries.CompanyProfileData(ctx, userID)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}	}

	// appsHistory, err := s.queries.ApplicationHistory(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// intsHistory, err := s.queries.InterviewHistory(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// testHistory, err := s.queries.TestHistory(ctx, userID)
	// if err != nil {
	// 	return nil, err
	// }

	// let us try to send a 'Sankey' type graph
	// sankeyChrt := gocharts.SankeyApplicants(&overData)
	
	return &dto.CompanyProfileData{
		UsersData: &userData,
		ProData: &data,
		SankeyChrt: sankeyCharts,
	}, nil
}

func (s *CompanyService) GetCompanyFile(ctx *gin.Context, userID int64, fileType string) (string, *errs.Error) {
	// get paths to all available files
	filePaths, err := s.queries.GetAllFilePathsCompany(ctx, userID)
	if err != nil {
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "failed to fetch file paths for user ID",
		}
	}

	// check what type is requested
	filepath := filePaths.PictureUrl.String
	// TODO: returns profile file even if requested file type is invalid
	// add switches if other files types are requested

    // Check if the file exists
    if _, err := os.Stat(filepath); err != nil {
        return "", &errs.Error{
			Type: errs.NotFound,
			Message: "could not find any file for given path",
		}
    }
	// return the requested file's path
	return filepath, nil
}
func (s *CompanyService) UpdateProfileDetails(ctx *gin.Context, userID int64, details *dto.UpdateCompanyDetails) (*errs.Error) {
	// TODO: verify incoming data
	err := s.queries.UpdateCompanyDetails(ctx, sqlc.UpdateCompanyDetailsParams{
		CompanyName: details.CompanyName,
		RepresentativeEmail: details.RepresentativeEmail,
		RepresentativeContact: details.RepresentativeContact,
		RepresentativeName: details.RepresentativeName,
		Address: details.CompanyAddress,
		Website: pgtype.Text{String: details.CompanyWebsite, Valid: true},
		Description: pgtype.Text{String: details.CompanyDescription, Valid: true},
		Industry: details.IndustryType,
		UserID: userID,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	return nil
}

func (s *CompanyService) UpdateFile(ctx *gin.Context, userID int64, file *multipart.FileHeader, fileType string) (*errs.Error) {
	var err error
	// get the file size and content-type
	size := file.Size
	ext := file.Header.Get("Content-Type")
	// get the expected size for the content type 
	expected := config.FileSizeForContentType[ext] 
	if (expected == 0) {
		// invalid file content type
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "Invalid file type.",
		}
	}
	if (expected < size) {
		// file size more than expected
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "The file size exceeds the limit.",
		}
	}

	userUUID, err := s.queries.GetUserUUIDFromUserID(ctx, userID)
	if err != nil {
		return nil
	}
	strUUID := hex.EncodeToString(userUUID.Bytes[:])

	nameType := strings.ToLower(fileType)
	storageDir := fmt.Sprintf("%sStorageDir", fileType)

	fileStoragePath := fmt.Sprintf("%s%s&%d&%s%s", os.Getenv(storageDir), strUUID, time.Now().Unix(), nameType, filepath.Ext(file.Filename))
	fileSavePath, err := utils.SaveFile(ctx, fileStoragePath, file)
	if err != nil {
		return nil
	}

	switch fileType {
	case "ProfilePic":
		err = s.queries.UpdateCompanyProfilePic(ctx, sqlc.UpdateCompanyProfilePicParams{
			PictureUrl: pgtype.Text{String: fileSavePath, Valid: true},
			UserID: userID,
		})
	default:
		return &errs.Error{
			Type: errs.NotFound,
			Message: "No such file type found. Use valid file type in url.",
		}
	}
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	return nil
}




















func (c *CompanyService) StudentProfileData(ctx *gin.Context, userID int64, studentid string) (*sqlc.StudentProfileForCompanyRow, *errs.Error) {

	studentID, err := strconv.ParseInt(studentid, 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	data, err := c.queries.StudentProfileForCompany(ctx, sqlc.StudentProfileForCompanyParams{
		UserID: userID,
		StudentID: studentID,
	})
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
		}
	}

	return &data, nil
}
