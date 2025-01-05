package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/apicalls"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
	"google.golang.org/api/forms/v1"
)

type StudentService struct {
	queries *sqlc.Queries
	RedisClient *redis.Client
	ApiCalls *apicalls.Caller
}
func NewStudentService(queriespool *sqlc.Queries, redisclient *redis.Client, apicalls *apicalls.Caller) *StudentService {
	return &StudentService{
		queries: queriespool,
		RedisClient: redisclient,
		ApiCalls: apicalls,
	}
}

func (s *StudentService) StudentFunc() {
	fmt.Println("student func")
}


func (s *StudentService) GetApplicableJobs(ctx *gin.Context, jobType string) ([]sqlc.GetApplicableJobsTypeFilterRow, error) {

	// TODO: apply all filters here
	userID, exists := ctx.Get("ID")
	if !exists {
		return []sqlc.GetApplicableJobsTypeFilterRow{}, errors.New("error getting user ID")
	}

	allapplicablejobsData, err := s.queries.GetApplicableJobsTypeFilter(ctx, sqlc.GetApplicableJobsTypeFilterParams{
		UserID: userID.(int64),
		Type: jobType,
	})
	if err != nil {
		return []sqlc.GetApplicableJobsTypeFilterRow{}, errors.New("unable to get all jobs from database")
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

func (s *StudentService) MyApplications(ctx *gin.Context, userId any, status string) ([]sqlc.GetMyApplicationsStatusFilterRow, error) {

	applicationsData, err := s.queries.GetMyApplicationsStatusFilter(ctx, sqlc.GetMyApplicationsStatusFilterParams{
		UserID: userId.(int64),
		Column2: status,
	})
	if err != nil {
		return []sqlc.GetMyApplicationsStatusFilterRow{}, err
	}

	return applicationsData, nil
}

func (s *StudentService) UpcomingData(ctx *gin.Context, userID int64, eventtype string)  (dto.AllUpcomingData, error) {

	var allData dto.AllUpcomingData
	var upcomingInter []sqlc.GetInterviewsForUserIDRow
	var upcomingTests []sqlc.GetTestsForUserIDRow
	var err error
	// if filter is 'all' or 'interview'
	if eventtype == "all" || eventtype == "interview" {
		upcomingInter, err = s.queries.GetInterviewsForUserID(ctx, userID)
		if err != nil {
			return allData, fmt.Errorf("error getting upcoming interviews for user : %s", err)
		}
		allData.InterviewsData = upcomingInter
	}
	// if filter is 'all' or 'test'
	if eventtype == "all" || eventtype == "test" {
		upcomingTests, err = s.queries.GetTestsForUserID(ctx, sqlc.GetTestsForUserIDParams{
			UserID: userID,
		})
		if err != nil {
			return allData, fmt.Errorf("error getting upcoming tests for user : %s", err)
		}
		allData.TestsData = upcomingTests
	}

	return allData, nil
}

func (s *StudentService) TestMetadata(ctx *gin.Context, userID int64, testid string) (sqlc.GetTestsForUserIDRow, error) {
	
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return sqlc.GetTestsForUserIDRow{}, fmt.Errorf("error parsing test ID : %s", err)
	}

	testMetadata, err := s.queries.GetTestsForUserID(ctx, sqlc.GetTestsForUserIDParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		return sqlc.GetTestsForUserIDRow{}, err
	}

	return testMetadata[0], nil
}

func (s *StudentService) TakeTest(ctx *gin.Context, userID int64, testid string, currentItemId string) (dto.TestQuestion, error) {

	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return dto.TestQuestion{}, fmt.Errorf("error parsing test ID : %s", err)
	}

	itemIdOrder := fmt.Sprintf("%sorder", testid)
	itemIdData := fmt.Sprintf("%sdata", testid)
	itemIdExpire := fmt.Sprintf("%sexpire%d", testid, userID)

	testData, err := s.queries.TakeTest(ctx, sqlc.TakeTestParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			return dto.TestQuestion{}, fmt.Errorf("unauthorized test ID for user ID : %d", userID)
		}
		return dto.TestQuestion{}, fmt.Errorf("error while fetching test metadata : %s", err)
	}

	exists, err := s.RedisClient.Exists(ctx, itemIdData).Result()
	if err != nil {
		return dto.TestQuestion{}, err
	} else if exists == 0 {
		gForm, err := s.ApiCalls.GetCompleteForm(testData.FileID)
		if err != nil {
			return dto.TestQuestion{}, err
		}

		for _, b := range gForm.Items {

			err := s.RedisClient.RPush(ctx, itemIdOrder, b.ItemId).Err()
			if err != nil {
				return dto.TestQuestion{}, err
			}

			values, err := json.Marshal(b)
			if err != nil {
				return dto.TestQuestion{}, err
			}
			err = s.RedisClient.HSet(ctx, itemIdData, b.ItemId, values).Err()
			if err != nil {
				return dto.TestQuestion{}, err
			}
		}
	}

	keysArray, err := s.RedisClient.LRange(ctx, itemIdOrder, 0, -1).Result()
	if err != nil {
		return dto.TestQuestion{}, err
	}

	var deserial *forms.Item 
	var index int
	var key string
	if currentItemId != "cover" {
		for index, key = range keysArray {
			if key == currentItemId {
				break
			}
		}
	}

	result, err := s.RedisClient.HGet(ctx, itemIdData, keysArray[index]).Bytes()
	if err != nil {
		return dto.TestQuestion{}, err
	}

	err = json.Unmarshal(result, &deserial)
	if err != nil {
		fmt.Println(err)
		return dto.TestQuestion{}, err
	}

	toSend := dto.TestQuestion{
		Item: deserial,
	}
	if len(keysArray) > 0  && index - 1 >= 0 {
		toSend.PrevId = keysArray[index - 1]
	}
	if index + 1 < len(keysArray) {
		toSend.NextId = keysArray[index + 1]
	}

	



	//

	// ttl, err := s.RedisClient.TTL(ctx, fmt.Sprintf("%seqwe", itemIdExpire)).Result()
	// if err != nil {
	// 	return dto.TestQuestion{}, err
	// }

	// if ()

	sec := (testData.Duration * 60) * int64(time.Second)	
	err = s.RedisClient.SetEx(ctx, itemIdExpire, "", time.Duration(sec)).Err()
	if err != nil {
		return dto.TestQuestion{}, err
	}

	return toSend, nil
}
