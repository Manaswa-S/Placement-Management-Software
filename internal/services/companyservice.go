package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/apicalls"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
)


type CompanyService struct {
	queries *sqlc.Queries
	GAPIService *apicalls.Caller
	RedisClient *redis.Client
}

func NewCompanyService(queriespool *sqlc.Queries, gapiService *apicalls.Caller, redisClient *redis.Client) *CompanyService {
	return &CompanyService{
		queries: queriespool,
		GAPIService: gapiService,
		RedisClient: redisClient,
	}
}

func (c *CompanyService) NewJobPost(ctx *gin.Context, jobdata *dto.NewJobData, userID int64) (error) {	
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
	// jobdata.Extras = extras

	extraJson, err := json.Marshal(extras)
	if err != nil {
		return errors.New("unable to marshal extras to json")
	}

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
			return err
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
		return err
	}

	return nil
}

func (c *CompanyService) ApplicantsData(ctx *gin.Context, userID int64, jobid string, appid string) (*[]sqlc.GetApplicantsRow, error){


	// parse jobid to int64
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return nil, err
	}
	appID, err := strconv.ParseInt(appid, 10, 64)
	if err != nil {
		return nil, err
	}
	
	// db call
	applicantsData, err := c.queries.GetApplicants(ctx, sqlc.GetApplicantsParams{
		UserID: userID,
		JobID: jobID,
		ApplicationID: appID,
	})
	if err != nil {
		return nil, err
	}

	// return 
	return &applicantsData, nil
}

func (c *CompanyService) GetResumeOrResultFilePath(ctx *gin.Context, userID int64, applicationid string, filetype string) (string, error) {
	// parse applicationid to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid application ID '%s': %w", applicationid, err)
	}

	// get both file paths (resume and result)
	filePaths, err := c.queries.GetResumeAndResultPath(ctx, sqlc.GetResumeAndResultPathParams{
		UserID: userID,
		ApplicationID: applicationId,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch file paths for application ID %d: %w", applicationId, err)
	}

	// check what type is requested
	filepath := filePaths.ResultUrl
	if filetype == "resume" {
		filepath = filePaths.ResumeUrl.String
	}

    // Check if the file exists
    if _, err := os.Stat(filepath); err != nil {
        return "", fmt.Errorf("file not found at path '%s': %w", filepath, err)
    }

	// by default updates the application status to UnderReview from Applied
	err = c.queries.ApplicationStatusToAnd(ctx, sqlc.ApplicationStatusToAndParams{
		Status: "UnderReview",
		ApplicationID: applicationId,
		Status_2: "Applied",
	})
	if err != nil {
		return "", fmt.Errorf("failed to update application status for ID %d: %w", applicationId, err)
	}

	// return file path 
	return filepath, nil
}

func (c *CompanyService) JobListings(ctx *gin.Context, userID int64) (*[]sqlc.GetJobListingsRow, error){

	// get the listings data 
	jobListings, err := c.queries.GetJobListings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job listings for user %d: %w", userID, err)
	}

	return &jobListings, nil
}

func (c *CompanyService) CloseJob(ctx *gin.Context, jobid string, userID int64) (error){

	// parse jobid from string to int64
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid job ID: %v", err)
	}

	// db query to change jobs.active_status to false
	err = c.queries.CloseJob(ctx, sqlc.CloseJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to close job with ID %d: %v", jobID, err)
	}

	return nil
}

func (c *CompanyService) DeleteJob(ctx *gin.Context, jobid string, userID int64) (error){
	// parse jobid from string to int64
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid job ID: %v", err)
	}
	// db query to delete job if exists
	err = c.queries.DeleteJob(ctx, sqlc.DeleteJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job with ID %d: %v", jobID, err)	}

	return nil
}

func (c *CompanyService) ShortList(ctx *gin.Context, applicationid string, userID int64) (error){

	// parse applicationid from string to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse appication id : %s", err)
	}

	// get the userID of the owner of the job
	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {
		return fmt.Errorf("failed to get user ID : %s", err)
	}
	// check if ID is same as the jobOwnerID to mitigate unauthorized reqs
	// so that only the actual owner of the job listing can shortlist the application
	if userID != jobOwnerID {
		return fmt.Errorf("unauthorized access with user ID %d", userID)
	}

	// change status to shortlisted
	err = c.queries.ApplicationStatusToAnd(ctx, sqlc.ApplicationStatusToAndParams{
		Status: "ShortListed",
		ApplicationID: applicationId,	
		Status_2: "UnderReview",
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *CompanyService) Reject(ctx *gin.Context, applicationid string, userID int64) (error){
	// parse applicationid from string to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return err
	}

	// get the userID of the owner of the job
	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil {
		return fmt.Errorf("failed to get user ID : %s", err)
	}
	// check if ID is same as the jobOwnerID to mitigate unauthorized reqs
	// so that only the actual owner of the job listing can shortlist the application
	if userID != jobOwnerID {
		return fmt.Errorf("unauthorized access with user ID %d", userID)
	}


	err = c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		Status: "Rejected",
		ApplicationID: applicationId,
	})
	if err != nil {
		return err
	}

	err = c.queries.InterviewStatusTo(ctx, sqlc.InterviewStatusToParams{
		Status: "Completed",
		ApplicationID: applicationId,
	})
	if err != nil {
		return err
	}
	
	return nil
}

func (c *CompanyService) ScheduleInterview(ctx *gin.Context, data *dto.NewInterview) (*errs.Error) {

	if (data.DateTime.Compare(time.Now()) != 1) {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "Interview Date-Time cannot be in the past.",
		}
	}

	// new insert into interviews table
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
			Message: err.Error(),
		}
	}

	// student name and email and job title and company name for email template
	studentData, err := c.queries.GetScheduleInterviewData(ctx, data.ApplicationId)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: err.Error(),
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
			Message: err.Error(),
		}
	}
	// send new interview email to student
	go utils.SendEmailHTML(template, []string{studentData.StudentEmail})

	// no error
	return nil
}

func (c *CompanyService) Offer(ctx *gin.Context, userID int64, applicationid string, offerLetter *multipart.FileHeader) (error) {
	// parse appliocation id from string to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return err
	}

	// get userID for the applicationId and check unauthorized req
	jobOwnerID, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil  || jobOwnerID != userID {
		return fmt.Errorf("you must be the job owner to offer applicant : %s", err)
	}

	// TODO: atomicity problem 
	// update interview status to 'Completed'
	err = c.queries.InterviewStatusTo(ctx, sqlc.InterviewStatusToParams{
		ApplicationID: applicationId,
		Status: "Completed",
	})
	if err != nil {
		return fmt.Errorf("error updating interview status : %s", err)
	}
	// update application status to 'Offered'
	err = c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		ApplicationID: applicationId,
		Status: "Offered",
	})
	if err != nil {
		return fmt.Errorf("error updating application status : %s", err)
	}

	// get company name, rep-name, etc
	offerData, err := c.queries.GetOfferLetterData(ctx, applicationId)
	if err != nil {
		return fmt.Errorf("error fetching data : %s", err)
	}
	// send offer email with offer letter attached
	template, err := utils.DynamicHTML("./template/emails/offerEmail.html", offerData)
	if err != nil {
		return fmt.Errorf("not able to execute email template : %s", err)
	}
	go utils.SendEmailHTMLWithAttachmentFileHeader(template, []string{offerData.StudentEmail}, offerLetter)

	// no error
	return nil
}

func (c *CompanyService) CancelInterview(ctx *gin.Context, userID int64, applicationid string) (error) {
	// parse application id from string to int64
	applicationId, err := strconv.ParseInt(applicationid, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing application ID : %s", err)
	}

	// check if user is owner
	jobOwnerId, err := c.queries.GetUserIDCompanyIDJobIDApplicationID(ctx, applicationId)
	if err != nil || jobOwnerId != userID {	
		return fmt.Errorf("you are not the owner of	this job: %s", err)
	}

	// get interview details for sending email
	data, err := c.queries.CancelInterviewEmailData(ctx, applicationId)
	if err != nil {
		return fmt.Errorf("unable to get interview data: %s", err)
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

	// execute email template and send email
	template, err := utils.DynamicHTML("./template/emails/interviewCancelled.html", newData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	go utils.SendEmailHTML(template, []string{data.StudentEmail})


	// TODO: dont delete interview, make it cancelled
	// delete interview
	err = c.queries.DeleteInterview(ctx, applicationId)
	if err != nil {
		return fmt.Errorf("error updating interview status : %s", err)
	}

	return nil
}

const (
	GForm = "GForms"
	CSVJSON = "CSVJSON"
	Manual = "Manual"
)

func (c *CompanyService) NewTestPost(ctx *gin.Context, newtestData dto.NewTestPost) (errs.Error) {

	var formID string
	var errf errs.Error
	// switch between upload method and call appropriate method
	switch newtestData.UploadMethod {
	case GForm :
		var gformData dto.NewTestGForms
		err := ctx.Bind(&gformData)
		if err != nil {
			return errs.Error{
				Type: errs.MissingRequiredField,
				Message: fmt.Sprintf("failed to bind the gform data: %s", err),
			}
		}
		formID, errf = c.NewTestPostGForm(ctx, gformData)
		if errf.Message != "" {
			return errf
		}
	case CSVJSON:
	case Manual:
	default:
	}	

	// get userID from context for companyID.
	userID := ctx.GetInt64("ID")
	// insert new test into db
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
	// switch between errors as needed
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			switch pgerr.Code {
				case errs.UniqueViolation:
					return errs.Error{
						Type: errs.UniqueViolation,
						Message: "The test already exists !",
					}
				default:
					return errs.Error{
						Type: errs.Internal,
						Message: fmt.Sprintf("error creating new test in db: %v", err),
					}
			}
		}
	}
	// get emails of all applicants to the job id that the test has been binded to 
	allEmails, err := c.queries.GetAllApplicantsEmailsForJob(ctx, newtestData.BindedJobId)
	if err != nil {
		ctx.Set("error", fmt.Sprintf("error fetching all applicants emails binded to the job id from db: %s", err.Error()))
	} else {
		// get job details for dynamic email
		jobDetails, err := c.queries.GetJobDetails(ctx, newtestData.BindedJobId)
		if err != nil {
			ctx.Set("error", fmt.Sprintf("error fetching job details from db: %s", err.Error()))
		} else {
			// generate template and send email to all applicants 
			newtestData.JobTitle = jobDetails.Title
			newtestData.CompanyName = jobDetails.CompanyName
			newtestData.FormattedEndDate = newtestData.EndDateTime.Format("2006-01-02")
			newtestData.FormattedEndTime = newtestData.EndDateTime.Format("15:04")

			template, err := utils.DynamicHTML("./template/emails/newTestEmail.html", newtestData)
			if err != nil {
				ctx.Set("error", err.Error())
			} else {
				go utils.SendEmailHTML(template, allEmails)
			}
		}
	}

	// done
	return errf
}

func (c *CompanyService) NewTestPostGForm(ctx *gin.Context, gformData dto.NewTestGForms) (string, errs.Error) {

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
			return "", errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
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
					return "", errs.Error{
						Type: errs.Internal,
						Message: err.Error(),
					}
				}
				// Set the key:value in the Redis Cache {responderUri : formId}
				err = c.RedisClient.Set(ctx, formData.ResponderUri, formData.FormId, 0).Err()
				if err != nil {
					return "", errs.Error{
						Type: errs.Internal,
						Message: fmt.Sprintf("error while inserting into redis: %v", err),
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
		return "", errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("error getting results from Redis : %v", err),
		}
	}

	formID, err := c.RedisClient.Get(ctx, gformData.ResponderLink).Result()
	if err == redis.Nil {
		// the result still does not exist
		// the user has not provided you with the access
		return "", errs.Error{
			Type: errs.IncompleteAction,
			Message: "Form not shared with email address.",
		}
	} else if err != nil {
		return "", errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("error getting results from Redis : %v", err),
		}	
	}
	
	return formID, errs.Error{}
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
	// switch betweem evemt type
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


func (c *CompanyService) UpdateInterview(ctx *gin.Context, userID int64, data *dto.UpdateInterview) (error) {

	newData, err := c.queries.UpdateInterview(ctx, sqlc.UpdateInterviewParams{
		UserID: userID,
		InterviewID: data.InterviewID,
		DateTime: pgtype.Timestamptz{Time: data.DateTime, Valid: true},
		Type: data.Type,
		Notes: pgtype.Text{String: data.Notes, Valid: true},
		Location: data.Location,
	})
	if err != nil {
		return err
	}

	stdData, err := c.queries.GetScheduleInterviewData(ctx, newData.ApplicationID)
	if err != nil {
		return err
	}
	data.StudentName = stdData.StudentName
	data.CompanyName = stdData.CompanyName
	data.JobTitle = stdData.Title
	data.DT = newData.DateTime

	// execute email template
	template, err := utils.DynamicHTML("./template/emails/interviewRescheduled.html", data)
	if err != nil {
		return err
	}
	// send new interview email to student
	go utils.SendEmailHTML(template, []string{stdData.StudentEmail})

	return nil
}

func (c *CompanyService) EditCutOff(ctx *gin.Context, userID int64, newData *dto.UpdateTest) (error) {

	// update the new thresholds in the db
	err := c.queries.UpdateTest(ctx, sqlc.UpdateTestParams{
		TestID: newData.TestID,
		UserID: userID,
		Threshold: int32(newData.Threshold),
	})
	if err != nil {
		return err
	}

	// call utils to generate the cumulative result draft
	_, err = utils.GenerateResultDraft(c.queries, c.GAPIService, newData.TestID)
	if err != nil {
		return err
	}

	// all ok
	return nil
}







func (c *CompanyService) StudentProfileData(ctx *gin.Context, userID int64, studentid string) (*sqlc.StudentProfileForCompanyRow, error) {

	studentID, err := strconv.ParseInt(studentid, 10, 64)
	if err != nil {
		return nil, err
	}

	data, err := c.queries.StudentProfileForCompany(ctx, sqlc.StudentProfileForCompanyParams{
		UserID: userID,
		StudentID: studentID,
	})
	if err != nil {
		return nil, err
	}

	return &data, nil
}
