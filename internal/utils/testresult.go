package utils

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/jackc/pgx/v5/pgtype"
	"go.mod/internal/apicalls"
	"go.mod/internal/config"
	"go.mod/internal/dto"
	gocharts "go.mod/internal/go-charts"
	sqlc "go.mod/internal/sqlc/generate"
)

type resultData struct {
	ctx context.Context
	queries *sqlc.Queries
	gapi *apicalls.Caller

	testID int64

	totalPoints int64
	testData sqlc.TestDataRow

	// err error
}


// GenerateTestResultDraft generates test's cumulative result draft.
// Returns the internal path to the result file or an error.
// The result file is an html page that requires internet connectivity to render, 	
// this is to maintain the interactivity of the charts and graphs
func GenerateTestResultDraft(sqlcQueries *sqlc.Queries, googleAPI *apicalls.Caller, testid int64) (string, error) {
	// have a separate context as this works async
	context, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// initialize struct for dependencies
	data := &resultData{
		ctx: context,
		queries: sqlcQueries,
		gapi: googleAPI,
		testID: testid,
	}

	// gets the fileid, calls the API's for the form data, extracts the correct answers from them
	// and evaluates the test responses, updates the test results for score, etc
	err := evaluate(data)
	if err != nil {
		// send an error email to admin
		fmt.Println(err)
	}

	// this function is responsible for generating all the charts for the result
	page, err := generateCumulativeCharts(data)
	if err != nil {
		// send an error email to admin
		fmt.Println(err)
	}

	// the file strucuture : ./test_result/{testid}/individual/...individual_results
	//                       ./test_result/{testid}/...cumulative_result
	// older versions of the result are over-written and only the latest is stored for the simplicity, versioning can be added later
	// a new file is created if none exists

	resultDir := fmt.Sprintf("%s%d/%s", os.Getenv("TestResultStorageDir"), testid, "individual")
	err = os.MkdirAll(resultDir, 0755)
	if err != nil {
		return "", err
	}

	cResultPath := fmt.Sprintf("%s%d/%d&%s%s", os.Getenv("TestResultStorageDir"), testid, testid, "testresult", ".html")
	file, err := os.Create(cResultPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// render the charts on the file as html
	err = page.Render(io.MultiWriter(file))
	if err != nil {
		return "", err
	}

	// TODO:
	// further parts of the result are added here

	// construct a local struct for the email data
	emailData := struct {
		CompanyName string
		TestID int64
		TestName string
		EndTime string
		Threshold int32
		TimeNow string
	} {
		data.testData.CompanyName,
		data.testData.TestID,
		data.testData.TestName,
		data.testData.EndTime,
		data.testData.Threshold,
		time.Now().Local().Format("03:04 PM 02-01-2006"),
	}
	// generate the email template
	template, err := DynamicHTML("./template/company/emails/resultdraft.html", emailData)
	if err != nil {
		return "", err
	}
	// send the email 
	err = SendEmailHTMLWithAttachmentFilePath(template, []string{data.testData.RepresentativeEmail}, cResultPath, fmt.Sprintf("%dresult%s", data.testData.TestID, ".html"))
	if err != nil {
		return "", err
	}
	// update the test result_url in the db, this is the cumulative result path and not individual results
	err = sqlcQueries.UpdateTestResultURLUnprotected(context, sqlc.UpdateTestResultURLUnprotectedParams{
		TestID: testid,
		ResultUrl: pgtype.Text{String: cResultPath, Valid: true},
	})
	if err != nil {
		return "", err
	}

	// return the result path with no errors
	return cResultPath, nil
}

func generateCumulativeCharts(data *resultData) (*components.Page, error) { 

	// get all required data from the db which is kept local
	// includes multiple db calls
	// this increases as we add more insights that require more data
	cumulativeData, err := data.queries.CumulativeResultData(data.ctx, data.testID)
	if err != nil {
		return nil, err
	}

	factor := float64(data.testData.Threshold) / float64(100)
	cutoffMarks := int64(factor * float64(data.totalPoints))
	passfailCount, err := data.queries.TestPassFailCount(data.ctx, sqlc.TestPassFailCountParams{
		TestID: data.testID,
		Score: pgtype.Int8{Int64: cutoffMarks, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// chart data calc starts
	
	// 1) a bar chart distribution of marks (X) vs no. of students in that range (Y) 
	// the total points of the test are divided equally into 10 parts and arranged in increasing order

	xaxis := make([]int64, 11)
	yaxis := make([]int64, 11)
	curr := (data.totalPoints / 10) + 1
	for i := range xaxis {
		xaxis[i] = curr * int64(i)
	}
	xaxis[10] = data.totalPoints
	for _, f := range cumulativeData {
		s := f.Score.Int64
		for i, x := range xaxis {
			if x >= s {
				yaxis[i]++
				break
			}
		}
	}
	xaxisStr := make([]string, 11)
	for i := range xaxis {
		xaxisStr[i] = fmt.Sprintf("%d / %d%%", xaxis[i], i * 10)
	}

	// 2) a pie chart that shows the pass / fail ratio

	// more coming soon !



	chartsData := &dto.CumulativeChartsData {
		Xaxis: xaxisStr,
		Yaxis: yaxis,
		PassCount: passfailCount.PassCount,
		FailCount: passfailCount.FailCount,
	}

	// we send all that calc data to the go-charts func to create charts out of them
	page, err := gocharts.ResultDraft(chartsData)
	if err != nil {
		return nil, err
	}

	// return the charts page with no errors
	return page, nil
}

func evaluate(data *resultData) (error) {
	var err error
	// get the fileid or the formid of the test
	data.testData, err = data.queries.TestData(data.ctx ,data.testID)
	if err != nil {
		return err
	}
	// clear the table to store answers, this is done to avoid unique constraint violation error
	// this can also be nested directly into the insert query or can be sorted with a on conflict clause
	// but it is not neccessary here, less complexity
	err = data.queries.ClearAnswersTable(data.ctx)
	if err != nil {
		return err
	}
	// this gets the complete form data including correct answers
	gForm, err := data.gapi.GetCompleteForm(data.testData.FileID)
	if err != nil {
		return err
	}
	// loop over the form, check if fields exist and extract values
	// insert the {questionId, answer, points} in the temp_answers table
	// this table is then used to evaluate the responses 
	for _, b := range gForm.Items {
		qItem := b.QuestionItem
		ans := []string{}
		if (qItem != nil && qItem.Question != nil && qItem.Question.Grading != nil &&
			qItem.Question.Grading.CorrectAnswers != nil && qItem.Question.Grading.CorrectAnswers.Answers != nil ) {
				
			for _, a := range qItem.Question.Grading.CorrectAnswers.Answers {
				ans = append(ans, a.Value)
			}

			err = data.queries.InsertAnswers(data.ctx, sqlc.InsertAnswersParams{
				QuestionID: b.ItemId,
				CorrectAnswer: ans,
				Points: pgtype.Int4{Int32: int32(qItem.Question.Grading.PointValue), Valid: true},
			})
			if err != nil {
				return err
			}
		}
	}
	// evaluate the responses accordingly
	// this also updates the testresults.score with the SUM(points)
	data.totalPoints, err = data.queries.EvaluateTestResult(data.ctx)
	if err != nil {
		return err
	}



	// the idea here is that it generates the results and a bunch of analytics and insights 
	// and renders it into a html page stored locally as temparoray files.
	// the path is returned as a string and used further
	// or can directly be referenced is hard-coded

	// several factors like thresholds, etc are taken into consideration while generating results
	// the result isnt made public until the company approves it
	
	return nil
}





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
				// TODO: raise critical error
				fmt.Println(err)
			}
		} (i)
	}
	for i := 0; i < config.NoOfTestResultFailedWorkers; i++ {
		wg.Add(1)
		go func (workerID int) {
			err := sendEmailFailed(workerID, failedQueue, &wg)
			if err != nil {
				// TODO: raise critical error
				fmt.Println(err)
			}
		} (i)
	}

	// enqueue tasks
	err := enqueueTestResults(taskQueue, data)
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}

func enqueueTestResults(taskQ chan *EmailTask, data *PublishData) error {

	// get the test metadata
	testData, err := data.queries.TestData(data.ctx, data.testID)
	if err != nil {
		return err
	}
	// TODO: this can actually be problematic,
	// we would want to extract qCount or other data from form instread of the test metadata which cannot be trusted
	data.qCount = testData.QCount

	// get all data neeeded to generate result
	stResult, err := data.queries.StudentTestResult(data.ctx, data.testID)
	if err != nil {
		return err
	}

	leng := len(stResult)
	for i := 0; i < leng; i++ {

		curr := stResult[i]
		// generate result for individual student
		resultPath, err := generateIndividualCharts(data, &curr)
		if err != nil {
			// TODO: add a retry logic for this too
			fmt.Println(err)
		}
		template, err := DynamicHTML("./template/student/emails/resultPublished.html", &IndividualEmailData{
			StudentName: curr.StudentName,
			TestName: testData.TestName,
			JobTitle: testData.Title,
			CompanyName: testData.CompanyName,
			StartTime: curr.StartTime,
		})
		if err != nil {
			// TODO: add a retry logic for this too
			fmt.Println(err)
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

	return nil
}

func sendEmailTask(workerID int, taskQ <-chan *EmailTask, failedQ chan *EmailTask, wg *sync.WaitGroup) error {
	defer wg.Done()

	fmt.Printf("worker %d started", workerID)

	for task := range taskQ {
		err := SendEmailHTMLWithAttachmentFilePath(task.BodyTemplate, task.RecipientEmail, task.AttachmentPath, "testresult.html")
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

	fmt.Printf("worker %d started", workerID)

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
