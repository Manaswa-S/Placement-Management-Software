package services

import (
	"encoding/json"
	"errors"
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