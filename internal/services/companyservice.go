package services

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
)


type CompanyService struct {
	queries *sqlc.Queries
}

func NewCompanyService(queriespool *sqlc.Queries) *CompanyService {
	return &CompanyService{queries: queriespool}
}

func (c *CompanyService) NewJobPost(ctx *gin.Context, jobdata dto.NewJobData) (sqlc.Job, error) {
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
			"CompanyName": true,
			"CompanyEmail": true,
			"CompanyLocation": true,
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
		return sqlc.Job{}, errors.New("unable to marshal extras to json")
	}
	// TODO: need to better validate incoming data 
	// add job data to db
	jobData, err := c.queries.InsertNewJob(ctx, sqlc.InsertNewJobParams{
		DataUrl: pgtype.Text{String: "", Valid: true},
		RepresentativeEmail: jobdata.CompanyEmail,
		Title: jobdata.JobTitle,
		Location: jobdata.JobLocation,
		Type: jobdata.JobType,
		Salary: jobdata.JobSalary,
		Skills: skills,
		Position: jobdata.JobPosition,
		Extras: extraJson,
	})
	if err != nil {
		return sqlc.Job{}, err
	}

	return jobData, nil
}

func (c * CompanyService) MyApplicants(ctx *gin.Context, userID int64, jobid string) ([]sqlc.GetApplicantsRow, error){

	var jobID int64
	var err error
	if jobid != "null" {
		jobID, err = strconv.ParseInt(jobid, 10, 64)
		if err != nil {
			return []sqlc.GetApplicantsRow{}, err
		}
	}

	applicantsData, err := c.queries.GetApplicants(ctx, sqlc.GetApplicantsParams{
		UserID: userID,
		JobID: jobID,
	})
	if err != nil {
		return []sqlc.GetApplicantsRow{}, err
	}

	return applicantsData, nil
}

func (c * CompanyService) GetFilePath(ctx *gin.Context, studentid string, jobid string, filetype string) (string, error){

	studentID, err := strconv.ParseInt(studentid, 10, 64)
	if err != nil {
		return "", err
	}
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return "", err
	}

	filePaths, err := c.queries.GetResumePath(ctx, studentID)
	if err != nil {
		return "", err
	}

	filepath := filePaths.ResultUrl
	if filetype == "resume" {
		filepath = filePaths.ResumeUrl.String
	}

	// Open the file using os.Open
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	err = c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		Status: "UnderReview",
		JobID: jobID,
		StudentID: studentID,
		Status_2: "Applied",
	})
	if err != nil {
		return "", err
	}

	return filepath, nil
}

func (c * CompanyService) MyJobListings(ctx *gin.Context, userID int64) ([]sqlc.GetJobListingsRow, error){

	joblistings, err := c.queries.GetJobListings(ctx, userID)
	if err != nil {
		return []sqlc.GetJobListingsRow{}, err
	}

	return joblistings, nil
}

func (c * CompanyService) CloseJob(ctx *gin.Context, jobid string, userID int64) (error){

	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return err
	}

	err = c.queries.CloseJob(ctx, sqlc.CloseJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c * CompanyService) DeleteJob(ctx *gin.Context, jobid string, userID int64) (error){

	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return err
	}

	err = c.queries.DeleteJob(ctx, sqlc.DeleteJobParams{
		JobID: jobID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c * CompanyService) ShortList(ctx *gin.Context, studentid string, jobid string) (error){

	studentID, err := strconv.ParseInt(studentid, 10, 64)
	if err != nil {
		return err
	}
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return err
	}

	err = c.queries.ApplicationStatusTo(ctx, sqlc.ApplicationStatusToParams{
		Status: "ShortListed",
		JobID: jobID,
		StudentID: studentID,
		Status_2: "UnderReview",
	})
	if err != nil {
		return err
	}

	return nil
}

func (c * CompanyService) Reject(ctx *gin.Context, studentid string, jobid string) (error){

	studentID, err := strconv.ParseInt(studentid, 10, 64)
	if err != nil {
		return err
	}
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return err
	}

	err = c.queries.ApplicationStatusToRejected(ctx, sqlc.ApplicationStatusToRejectedParams{
		Status: "Rejected",
		JobID: jobID,
		StudentID: studentID,
	})
	if err != nil {
		return err
	}

	err = c.queries.InterviewStatusTo(ctx, sqlc.InterviewStatusToParams{
		Status: "Completed",
		JobID: jobID,
		StudentID: studentID,
	})
	if err != nil {
		return err
	}
	

	return nil
}

func (c * CompanyService) ScheduleInterview(ctx *gin.Context, data dto.NewInterview) (sqlc.ScheduleInterviewRow, error) {

	// Convert time to microseconds since midnight
    microsecondsSinceMidnight := int64(data.Time.Hour())*3600000000 + int64(data.Time.Minute())*60000000 + int64(data.Time.Second())*1000000
	
	intData, err := c.queries.ScheduleInterview(ctx, sqlc.ScheduleInterviewParams{
		JobID: data.JobId,
		StudentID: data.StudentId,
		UserID: data.UserId,
		Date: pgtype.Date{Time: data.Date, Valid: true},
		Time: pgtype.Time{Microseconds: microsecondsSinceMidnight, Valid: true},
		Type: data.Type,
		Notes: pgtype.Text{String: data.Notes, Valid: true},
	})
	if err != nil {
		return sqlc.ScheduleInterviewRow{}, err
	}
	
	return intData, nil
}