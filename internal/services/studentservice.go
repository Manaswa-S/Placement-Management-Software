package services

import (
	"encoding/base64"
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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	gocharts "go.mod/internal/go-charts"
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


func (s *StudentService) GetApplicableJobs(ctx *gin.Context, jobType string) (*[]sqlc.GetApplicableJobsTypeFilterRow, error) {

	// TODO: apply all filters here
	userID, exists := ctx.Get("ID")
	if !exists {
		return nil, errors.New("error getting user ID")
	}

	allapplicablejobsData, err := s.queries.GetApplicableJobsTypeFilter(ctx, sqlc.GetApplicableJobsTypeFilterParams{
		UserID: userID.(int64),
		Type: jobType,
	})
	if err != nil {
		return nil, errors.New("unable to get all jobs from database")
	}

	return &allapplicablejobsData, nil
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

func (s *StudentService) MyApplications(ctx *gin.Context, userId any, status string) (*[]sqlc.GetMyApplicationsStatusFilterRow, error) {

	applicationsData, err := s.queries.GetMyApplicationsStatusFilter(ctx, sqlc.GetMyApplicationsStatusFilterParams{
		UserID: userId.(int64),
		Column2: status,
	})
	if err != nil {
		return nil, err
	}

	return &applicationsData, nil
}

func (s *StudentService) UpcomingData(ctx *gin.Context, userID int64, eventtype string)  (*dto.Upcoming, error) {

	switch eventtype {
	case "interviews":
		uInts, err := s.queries.UpcomingInterviewsStudent(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &dto.Upcoming{
			Data: uInts,
		}, nil
	case "tests":
		uTests, err := s.queries.UpcomingTestsStudent(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &dto.Upcoming{
			Data: uTests,
		}, nil
	default:
		return nil, nil
	}
}

func (s *StudentService) TestMetadata(ctx *gin.Context, userID int64, testid string) (*sqlc.TestMetadataRow, error) {
	
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing test ID : %s", err)
	}

	testMetadata, err := s.queries.TestMetadata(ctx, sqlc.TestMetadataParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		return nil, err
	}

	return &testMetadata, nil
}

// TODO: This function seems too inefficient, too much in one func and too many redundant db calls.
func (s *StudentService) TakeTest(ctx *gin.Context, userID int64, testid string, currentItemId string, response dto.TestResponse) (*dto.TestQuestion, *errs.Error) {

	// parse the test ID
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("error parsing test ID : %v", err),
		}
	}

	// get the file id and duration for the particular test
	testData, err := s.queries.TakeTest(ctx, sqlc.TakeTestParams{
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		// if it returns zero rows, that means the userID cannot give the requested test ie. he needs to first apply to the job  
		if err.Error() == errs.NoRowsMatch {
			return nil, &errs.Error{
				Type: errs.Unauthorized,
				Message: "You are not authorized to access this Test. Apply to the job first.",
			}
		}
		// random error
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get test metadata : %v", err),
		}
	}
	// compares the returned test end_time with time.Now() and returns error if end_time was in the past
	if testData.EndTime.Time.Compare(time.Now()) == -1 {
		return nil, &errs.Error{
			Type: errs.Unauthorized,
			Message: "The end time for the test has gone by. You cannot give the test now.",
		}
	}

	

	// define the redis keys
	itemIdOrder := fmt.Sprintf("%sorder", testid)
	itemIdData := fmt.Sprintf("%sdata", testid)
	itemIdExpire := fmt.Sprintf("%sexpire%d", testid, userID)

	// the user is authenticated and has not completed the test yet
	// so check if the test data exists in cache
	exists, err := s.RedisClient.Exists(ctx, itemIdData).Result()
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to check cache for test data existence : %v", err),
		}
	} else if exists == 0 {
		// the test data does not exist in cache
		// call the test api to get complete form data
		gForm, err := s.ApiCalls.GetCompleteForm(testData.FileID)
		if err != nil {
			return nil, &errs.Error{
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
				// TODO: this should ideally be done concurrently, for a real test with even 30 images thats like 45s of buffering for the user
				fileByte, err := utils.GetFileFromPath(attr.ContentUri , config.TempFileStorage)
				if err != nil {
					return nil, &errs.Error{
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
				return nil, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to push itemid to cache list : %v", err),
				}
			}

			// marshal the item
			values, err := json.Marshal(b)
			if err != nil {
				return nil, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to marshal item : %v", err),
				}
			}
			// a hash is maintained for each test, key is testid and the fields are itemid
			// set the marshalled in the hash
			err = s.RedisClient.HSet(ctx, itemIdData, b.ItemId, values).Err()
			if err != nil {
				return nil, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to set item in hash in cache : %v", err),
				}
			}
		}
	}

	// this should ideally be done at the start of the test, but a new row was being inserted at the very start 
	// and if the later code errors out, 
	// the row stays, effectively saying the the user has started the test whereas he received an error
	// this is still error prone, but errors further down are all native and will be detected instantly
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
				return nil, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to update test result : %v", err),
				}
			}
		} else {
			return nil, &errs.Error{
				Type: errs.Internal,
				Message: fmt.Sprintf("failed to check if entry exists in testresult : %v", err),
			}
		}
	}
	// the user has started the test, check if finished
	if (resultData.EndTime.Valid) {
		return nil, &errs.Error{
			Type: errs.ObjectExists,
			Message: "The user has already given the test.",
		}	
	}

	// here, either the test data has been called and set in the cache or it was already present
	// either way, we have the complete test data here and its order list 
	// get the entire order list
	keysArray, err := s.RedisClient.LRange(ctx, itemIdOrder, 0, -1).Result()
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get list of itemid : %v", err),
		}
	}

	var deserial *forms.Item 
	var index int
	var key string
	// get the index of the itemid requested
	if currentItemId != "cover" {
		// update the responses in the db
		if (response.ItemID != "") {
			err = s.queries.UpdateResponse(ctx, sqlc.UpdateResponseParams{
				ResultID: resultData.ResultID,
				QuestionID: response.ItemID,
				Response: response.Response,
				TimeTaken: pgtype.Int8{Int64: response.TimeTaken, Valid: true},
			})
			if err != nil {
				return nil, &errs.Error{
					Type: errs.Internal,
					Message: fmt.Sprintf("failed to update test result in db : %v", err),
				}
			}
		}

		for index, key = range keysArray {
			if key == currentItemId {
				break
			}
		}
	}
	// get the entire item data from cache in bytes
	result, err := s.RedisClient.HGet(ctx, itemIdData, keysArray[index]).Bytes()
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get item data from cache : %v", err),
		}
	}
	// unmarshal it to json (deserial)
	err = json.Unmarshal(result, &deserial)
	if err != nil {
		return nil, &errs.Error{
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
		return nil, &errs.Error{
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
			return nil, &errs.Error{
				Type: errs.ObjectExists,
				Message: "The time is up for the test. The test will now be auto-submitted",
			}
		} else if (!resultData.StartTime.Valid) {
			// start of the test
			// create new timer with ttl set to duration from db
			sec := (testData.Duration * 60) * int64(time.Second)	
			err = s.RedisClient.SetEx(ctx, itemIdExpire, "", time.Duration(sec)).Err()
			if err != nil {
				return nil, &errs.Error{
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
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: fmt.Sprintf("failed to get timer for test : %v", err),
		}
	}
	// send ttl with response changed from nanoseconds to seconds
	toSend.TTL = ttl / 1e9

	// no error, send response
	return &toSend, nil
}

func (s *StudentService) SubmitTest(ctx *gin.Context, userID int64, testid string) (*errs.Error) {
	// parse test id from string to int64
	testID, err := strconv.ParseInt(testid, 10, 64)
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse Test Id from string",
		}
	}
	// update the endtime of the result from null to time.now()
	result_id, err := s.queries.SubmitTest(ctx, sqlc.SubmitTestParams{
		EndTime: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UserID: userID,
		TestID: testID,
	})
	if err != nil {
		if (err.Error() == errs.NoRowsMatch) {
			return &errs.Error{
				Type: errs.InvalidState,
				Message: "The user has not started the test yet.",
			}
		} else {
			return &errs.Error{
				Type: errs.Internal,
				Message: err.Error(),
			}
		}
	}
	// send an email of confirmation of test submission with the result id for future reference
	// and it also goes into the 'Completed' page
	// TODO:
	fmt.Println(result_id)

	// return
	return nil
}

func (s *StudentService) Completed(ctx *gin.Context, userID int64, tab string) (*dto.Completed, error) {

	switch tab {
	case "tests":
		cTests, err := s.queries.CompletedTestsStudent(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &dto.Completed{
			Data: cTests,
		}, nil
	case "interviews":
		cInts, err := s.queries.CompletedInterviewsStudent(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &dto.Completed{
			Data: cInts,
		}, nil
	default:
		return nil, nil
	}
}

func (s *StudentService) ProfileData(ctx *gin.Context, userID int64) (*dto.ProfileData, error) {
	overData, err := s.queries.ApplicationsStatusCounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	userData, err := s.queries.UsersTableData(ctx, userID)
	if err != nil {
		return nil, err
	}

	data, err := s.queries.ProfileData(ctx, userID)
	if err != nil {
		return nil, err
	}

	appsHistory, err := s.queries.ApplicationHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	intsHistory, err := s.queries.InterviewHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	testHistory, err := s.queries.TestHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	// let us try to send a 'Sankey' type graph
	sankeyChrt := gocharts.SankeyApplications(&overData)
	
	return &dto.ProfileData{
		OverData: &overData,
		UsersData: &userData,
		ProData: &data,
		AppsHistory: &appsHistory,
		IntsHistory: &intsHistory,
		TestsHistory: &testHistory,
		SankeyChrt: sankeyChrt,
	}, nil
}

func (s *StudentService) GetStudentFile(ctx *gin.Context, userID int64, fileType string) (string, *errs.Error) {
	// get paths to all available files
	filePaths, err := s.queries.GetAllFilePaths(ctx, userID)
	if err != nil {
		return "", &errs.Error{
			Type: errs.Internal,
			Message: "failed to fetch file paths for user ID",
		}
	}

	// check what type is requested
	filepath := filePaths.PictureUrl.String
	if fileType == "resume" {
		filepath = filePaths.ResumeUrl.String
	}
	if fileType == "result" {
		filepath = filePaths.ResultUrl
	}

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

func (s *StudentService) UpdateDetails(ctx *gin.Context, userID int64, details *dto.UpdateStudentDetails) (error) {

	err := s.queries.UpdateStudentDetails(ctx, sqlc.UpdateStudentDetailsParams{
		Course: details.Course,
		Department: details.Department,
		YearOfStudy: details.YearOfStudy,
		Cgpa: pgtype.Float8{Float64: details.CGPA, Valid: true},
		ContactNo: details.ContactNumber,
		Address: pgtype.Text{String: details.Address, Valid: true},
		Skills: pgtype.Text{String: details.Skills, Valid: true},
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *StudentService) UpdateFile(ctx *gin.Context, userID int64, file *multipart.FileHeader, fileType string) (*errs.Error) {
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
	case "Resume":
		err = s.queries.UpdateStudentResume(ctx, sqlc.UpdateStudentResumeParams{
			ResumeUrl: pgtype.Text{String: fileSavePath, Valid: true},
			UserID: userID,
		})
	case "Result":
		err = s.queries.UpdateStudentResult(ctx, sqlc.UpdateStudentResultParams{
			ResultUrl: fileSavePath,
			UserID: userID,
		})
	case "ProfilePic":
		err = s.queries.UpdateStudentProfilePic(ctx, sqlc.UpdateStudentProfilePicParams{
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