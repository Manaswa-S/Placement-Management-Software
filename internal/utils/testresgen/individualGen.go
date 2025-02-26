package testresgen

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/go-echarts/go-echarts/v2/opts"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	"go.mod/internal/dto"
	gocharts "go.mod/internal/go-charts"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
	"go.mod/internal/utils/ctxutils"
)



type PublishData struct {
	ctx context.Context
	queries *sqlc.Queries
	gapi *apicalls.Caller

	testID int64
	qCount int64
}
type EmailTask struct {
	RecipientEmail 	[]string
	BodyTemplate	bytes.Buffer
	AttachmentPath 	string
}
type IndividualEmailData struct {
	StudentName string
	TestName string
	JobTitle string
	CompanyName string
	StartTime string

}

func PublishTestResults(sqlcQueries *sqlc.Queries, googleAPI *apicalls.Caller, testid int64) (error) {
	// the context is cancelled before the workers can finish
	// use wg.Wait() if context is needed 
	context, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// manage params
	data := &PublishData{
		ctx: context,
		queries: sqlcQueries,
		gapi: googleAPI,
		testID: testid,
	}

	// create neccessary waits and channels
	var wg sync.WaitGroup
	taskQueue := make(chan *EmailTask, config.TestResultTaskQueueBufferCapacity) // main channel containing tasks
	failedQueue := make(chan *EmailTask, config.TestResultFailedQueueBufferCapacity) // channel containing failed tasks for retry

	// spawn workers
	for i := 0; i < config.NoOfTestResultTaskWorkers; i++ {
		wg.Add(1)
		go func (workerID int) {
			err := sendEmailTask(workerID, taskQueue, failedQueue, &wg)
			if err != nil {
				ctxutils.NewError(&dto.ErrorData{
					Critical: fmt.Sprintf("Failed to spawn worker for test result publishing worker ID : %d : %v", 
						workerID, err.Error()),
				})
				return
			}
		} (i)
	}
	for i := 0; i < config.NoOfTestResultFailedWorkers; i++ {
		wg.Add(1)
		go func (workerID int) {
			err := sendEmailFailed(workerID, failedQueue, &wg)
			if err != nil {
				ctxutils.NewError(&dto.ErrorData{
					Critical: fmt.Sprintf("Failed to spawn failed-worker for test result publishing worker ID : %d : %v", 
						workerID, err.Error()),
				})
				return
			}
		} (i)
	}

	// enqueue tasks
	failedGen, err := enqueueTestResults(taskQueue, data)
	if err != nil {
		ctxutils.NewError(&dto.ErrorData{
			Critical: fmt.Sprintf("Failed to enqueue test result tasks : %v", 
				err.Error()),
		})
		return err
	}

	wg.Wait()

	ctxutils.NewError(&dto.ErrorData{
		Info: fmt.Sprintf("Test Results published successfully.\n" + 
		"Test ID : %d \n" + 
		"Failed Generations : %d \n" + 
		"Failed Result IDs : %d \n" + 
		"", testid, len(failedGen), failedGen),
	})

	return nil
}

func enqueueTestResults(taskQ chan *EmailTask, data *PublishData) ([]int64, error) {
	
	failedGen := make([]int64, 0)

	// get the test metadata
	testData, err := data.queries.TestData(data.ctx, data.testID)
	if err != nil {
		return failedGen, errors.New("Failed to get test metadata : " + err.Error())
	}
	// TODO: this can actually be problematic,
	// we would want to extract qCount or other data from form instread of the test metadata which cannot be trusted
	data.qCount = testData.QCount

	// get all data neeeded to generate result
	stResult, err := data.queries.StudentTestResult(data.ctx, data.testID)
	if err != nil {
		return failedGen, errors.New("Failed to get data for result generation : " + err.Error())
	}

	leng := len(stResult)
	for i := 0; i < leng; i++ {

		curr := stResult[i]
		// generate result for individual student
		resultPath, err := generateIndividualCharts(data, &curr)
		if err != nil {
			// TODO: well the actual error string is still not being logged
			failedGen = append(failedGen, curr.ResultID)
			continue
		}

		template, err := utils.DynamicHTML("./template/student/emails/resultPublished.html", &IndividualEmailData{
			StudentName: curr.StudentName,
			TestName: testData.TestName,
			JobTitle: testData.Title,
			CompanyName: testData.CompanyName,
			StartTime: curr.StartTime,
		})
		if err != nil {
			failedGen = append(failedGen, curr.ResultID)
			continue
		}
		// enque the email with attachment path and template in the task channel
		mail := &EmailTask{
			RecipientEmail: []string{curr.StudentEmail},
			BodyTemplate: template,
			AttachmentPath: resultPath,
		}
		taskQ <- mail
	}

	defer close(taskQ)

	return failedGen, nil
}

func sendEmailTask(workerID int, taskQ <-chan *EmailTask, failedQ chan *EmailTask, wg *sync.WaitGroup) error {
	defer wg.Done()

	for task := range taskQ {
		err := utils.SendEmailHTMLWithAttachmentFilePath(task.BodyTemplate, task.RecipientEmail, task.AttachmentPath, "testresult.html")
		if err != nil {
			failedQ <- task
		} 
	}

	// close the failed tasks channel
	defer close(failedQ)

	return nil
}

func sendEmailFailed(workerID int, failedQ <-chan *EmailTask, wg *sync.WaitGroup) error {
	defer wg.Done()

	for task := range failedQ {
		// TODO: need a failure strategy
		fmt.Println(task.RecipientEmail)
	}

	return nil
}

func generateIndividualCharts(data *PublishData, curr *sqlc.StudentTestResultRow) (string, error) { 

	// curr.Score
	// curr.TotalTimeTaken

	accuracy := ( float32(curr.CorrectResponse) / float32(curr.QuestionsAttempted)) * 100
	
	// get the complete page with charts on it
	page, err := gocharts.IndividualResult(&dto.IndividualChartsData{
		FunnelDimensions: []string{"Total", "Attempted", "Correct"},
		FunnelValues: []int64{data.qCount, curr.QuestionsAttempted, curr.CorrectResponse},

		RadarNames: []*opts.Indicator{
			{Name: "Accuracy", Max: 100, Color: "red"},
			{Name: "Attempted", Max: float32(data.qCount), Color: "blue"},
			{Name: "Correct", Max: float32(data.qCount), Color: "green"},
		},
		RadarValues: []float32{accuracy, float32(curr.QuestionsAttempted), float32(curr.CorrectResponse)},
	})
	if err != nil {
		return "", err
	}

	
	// add other stuff here
	// add graphs and charts


	// construct file name and path
	strUUID := hex.EncodeToString(curr.UserUuid.Bytes[:])
	resultName := fmt.Sprintf("%s&%d&%s%s", strUUID, data.testID, "testresult", ".html")
	resultPath := fmt.Sprintf("%s%d/%s/%s", os.Getenv("TestResultStorageDir"), data.testID, "individual", resultName)
	// create the file
	file, err := os.Create(resultPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// render/write the complete page to the empty file
	err = page.Render(io.MultiWriter(file))
	if err != nil {
		return "", err
	}

	// return result path and no error
	return resultPath, nil 
}
