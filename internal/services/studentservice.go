package services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	sqlc "go.mod/internal/sqlc/generate"
)

type StudentService struct {
	queries *sqlc.Queries
}
func NewStudentService(queriespool *sqlc.Queries) *StudentService {
	return &StudentService{queries: queriespool}
}

func (s *StudentService) StudentFunc() {
	fmt.Println("student func")
}


func (s *StudentService) GetApplicableJobs(ctx *gin.Context) ([]sqlc.GetApplicableJobsRow, error) {

	// TODO: apply all filters here
	userID, exists := ctx.Get("ID")
	if !exists {
		return []sqlc.GetApplicableJobsRow{}, errors.New("error getting user ID")
	}

	allapplicablejobsData, err := s.queries.GetApplicableJobs(ctx, userID.(int64))
	if err != nil {
		return []sqlc.GetApplicableJobsRow{}, errors.New("unable to get all jobs from database")
	}

	return allapplicablejobsData, nil
}

func (s *StudentService) NewApplication(ctx *gin.Context, userId int64, jobid string) (error) {

	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return errors.New("unable to parse job id from string to int")
	}

	err = s.queries.InsertNewApplication(ctx, sqlc.InsertNewApplicationParams{
		JobID: jobID,
		UserID: userId,
		DataUrl: pgtype.Text{String: "", Valid: true},
	})
	if err != nil {
		fmt.Println(err)
		return errors.New("unable to insert new application into database")
	}

	return nil
}

func (s *StudentService) CancelApplication(ctx *gin.Context, userID int64, jobid string)  (error) {
	
	jobID, err := strconv.ParseInt(jobid, 10, 64)
	if err != nil {
		return err
	}

	err = s.queries.CancelApplication(ctx, sqlc.CancelApplicationParams{
		UserID: userID,
		JobID: jobID,
	})
	if err != nil {
		return err	
	}

	return nil
}

func (s *StudentService) MyApplications(ctx *gin.Context) ([]sqlc.GetMyApplicationsRow, error) {

	userID, exists := ctx.Get("ID")
	if !exists {
		return []sqlc.GetMyApplicationsRow{}, errors.New("error getting user ID")
	}

	applicationsData, err := s.queries.GetMyApplications(ctx, userID.(int64))
	if err != nil {
		return []sqlc.GetMyApplicationsRow{}, err
	}

	return applicationsData, nil
}