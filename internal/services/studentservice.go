package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
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

func (s *StudentService) TakeTest(ctx *gin.Context, userID int64, testid string, currentItemId string) (dto.TestQuestion, *errs.Error) {

	// parse the test ID
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("error parsing test ID : %v", err),
		}
	}


	// check if column exists in testResult with testID and userID ie. test is already given
	resultData, err := s.queries.IsTestGiven(ctx, sqlc.IsTestGivenParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		if err.Error() == errs.NoRowsMatch {
			// the user has not yet given the test
			// add a new entry with the now timestamp
			err := s.queries.NewTestResult(ctx, sqlc.NewTestResultParams{
				UserID: userID,
				TestID: testID,
				StartTime: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			})
			if err != nil {
				return dto.TestQuestion{}, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to update test result : %v", err),
				}
			}
		} else {
			return dto.TestQuestion{}, &errs.Error{
				Type: errs.Internal,
				Message: fmt.Sprintf("failed to check if entry exists in testresult : %v", err),
			}
		}
	}
	// the user has started the test, check if finished
	if (resultData.EndTime.Valid) {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.ObjectExists,
			Message: "The user has already given the test.",
		}	
	}

	// define the redis keys
	itemIdOrder := fmt.Sprintf("%sorder", testid)
	itemIdData := fmt.Sprintf("%sdata", testid)
	itemIdExpire := fmt.Sprintf("%sexpire%d", testid, userID)

	// get the file id and duration for the particular test
	testData, err := s.queries.TakeTest(ctx, sqlc.TakeTestParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		// if it returns zero rows, that means the userID cannot give the requested test ie. he needs to first apply to the job  
		if err.Error() == errs.NoRowsMatch {
			return dto.TestQuestion{}, &errs.Error{
				Type: errs.Unauthorized,
				Message: "You are not authorized to access this Test. Apply to the job first.",
			}
		}
		// random error
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get test metadata : %v", err),
		}
	}

	// the user is authenticated and has not completed the test yet
	// so check if the test data exists in cache
	exists, err := s.RedisClient.Exists(ctx, itemIdData).Result()
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to check cache for test data existence : %v", err),
		}
	} else if exists == 0 {
		// the test data does not exist in cache
		// call the test api to get complete form data
		gForm, err := s.ApiCalls.GetCompleteForm(testData.FileID)
		if err != nil {
			return dto.TestQuestion{}, &errs.Error{
				Type: errs.Internal,
				Message: fmt.Sprintf("failed during API calls : %v", err),
			}
		}
		// loop over every item in the response
		for _, b := range gForm.Items {
			// check if the test has any media like images
			// if yes, get them and convert to []byte then encode in base64
			// set the image attribute back to this media file
			var attr *forms.Image
			if (b.ImageItem != nil || (b.QuestionItem != nil && b.QuestionItem.Image != nil)) {
				switch {
				case b.ImageItem != nil:
					attr = b.ImageItem.Image
				case b.QuestionItem != nil && b.QuestionItem.Image != nil:
					attr = b.QuestionItem.Image
				}
				fileByte, err := utils.GetFileFromPath(attr.ContentUri , config.TempFileStorage)
				if err != nil {
					return dto.TestQuestion{}, &errs.Error{
						Type: errs.Internal,
						Message: fmt.Sprintf("failed to get file from url : %v", err),
					}
				}
				encodedImage := base64.StdEncoding.EncodeToString(fileByte)
				attr.ContentUri = encodedImage
			}

			// the answers are also included in the response
			// so remove them and set their fields to NULL
			qItem := b.QuestionItem
			if (qItem != nil && qItem.Question != nil && qItem.Question.Grading != nil) {
				qItem.Question.Grading = nil
			}

			// a list of the items is maintained in cache too
			// push the current items id to the list
			err = s.RedisClient.RPush(ctx, itemIdOrder, b.ItemId).Err()
			if err != nil {
				return dto.TestQuestion{}, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to push itemid to cache list : %v", err),
				}
			}

			// marshal the item
			values, err := json.Marshal(b)
			if err != nil {
				return dto.TestQuestion{}, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to marshal item : %v", err),
				}
			}
			// a hash is maintained for each test, key is testid and the fields are itemid
			// set the marshalled in the hash
			err = s.RedisClient.HSet(ctx, itemIdData, b.ItemId, values).Err()
			if err != nil {
				return dto.TestQuestion{}, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to set item in hash in cache : %v", err),
				}
			}
		}
	}

	// here, either the test data has been called and set in the cache or it was already present
	// either way, we have the complete test data here and its order list 
	// get the entire order list
	keysArray, err := s.RedisClient.LRange(ctx, itemIdOrder, 0, -1).Result()
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get list of itemid : %v", err),
		}
	}

	var deserial *forms.Item 
	var index int
	var key string
	// get the index of the itemid requested
	if currentItemId != "cover" {
		for index, key = range keysArray {
			if key == currentItemId {
				break
			}
		}
	}
	// get the entire item data from cache in bytes
	result, err := s.RedisClient.HGet(ctx, itemIdData, keysArray[index]).Bytes()
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get item data from cache : %v", err),
		}
	}
	// unmarshal it to json (deserial)
	err = json.Unmarshal(result, &deserial)
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to unMarshal item : %v", err),
		}
	}
	// generate the struct to send as response
	toSend := dto.TestQuestion{
		Item: deserial,
	}
	// include the previous and next item id's so to call the prev/next questions
	if len(keysArray) > 0  && index - 1 >= 0 {
		toSend.PrevId = keysArray[index - 1]
	}
	if index + 1 < len(keysArray) {
		toSend.NextId = keysArray[index + 1]
	}



	// we check if the timer for the test (testid) for user (userid) has expired 
	// to avoid submissions/item requests after timeout if the user messed with frontend
	ttlExists, err := s.RedisClient.Exists(ctx, itemIdExpire).Result()
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to check if timer exists : %v", err),
		}
	}
	if (ttlExists == 0) {
		// 2 cases:
		// the timer has ended
		// this is the start of the test for the user and hence first request and the timer has yet not been set

		// if startTime exists (the user had started the test and hence has an entry in the db) 
		// but endTime deos not exist (the user has not submitted the test)
		// ie. something went wrong on the frontend or the user messed up
		if (!resultData.EndTime.Valid && resultData.StartTime.Valid) {
			// we submit the test, return an error, and probably redirect to dashboard to prevent further requests
			return dto.TestQuestion{}, &errs.Error{
				Type: errs.ObjectExists,
				Message: "The time is up for the test. The test will now be auto-submitted",
			}
		} else if (!resultData.StartTime.Valid) {
			// start of the test
			// create new timer with ttl set to duration from db
			sec := (testData.Duration * 60) * int64(time.Second)	
			err = s.RedisClient.SetEx(ctx, itemIdExpire, "", time.Duration(sec)).Err()
			if err != nil {
				return dto.TestQuestion{}, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to set timer for test : %v", err),
				}
			}
		}
	}
	// the timer was already available or has been set
	// get it
	ttl, err := s.RedisClient.TTL(ctx, itemIdExpire).Result()
	if err != nil {
		return dto.TestQuestion{}, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get timer for test : %v", err),
		}
	}
	// send ttl with response changed from nanoseconds to seconds
	toSend.TTL = ttl / 1e9

	// no error, send response
	return toSend, nil
}
