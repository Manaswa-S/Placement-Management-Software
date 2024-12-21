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

func (c * CompanyService) MyApplicants(ctx *gin.Context, userID int64) ([]sqlc.GetApplicantsRow, error){

	applicantsData, err := c.queries.GetApplicants(ctx, userID)
	if err != nil {
		return []sqlc.GetApplicantsRow{}, err
	}

	return applicantsData, nil
}

func (c * CompanyService) GetFilePath(ctx *gin.Context, studentID int64, filetype string) (string, error){

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