package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"go.mod/internal/apicalls"
	gocharts "go.mod/internal/go-charts"
	sqlc "go.mod/internal/sqlc/generate"
)

var (
	ctx context.Context
	queries *sqlc.Queries
	testID int64
	gapi *apicalls.Caller
)

var (
	totalPoints int64 = 0
)

func GenerateTestResult(queRies *sqlc.Queries, gAPI *apicalls.Caller, tesTID int64) (error) {
	context, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	queries = queRies
	gapi = gAPI
	testID = tesTID
	ctx = context
	fmt.Println("here")
	err := evaluate(queries, gapi, testID)
	if err != nil {
		// send an error email to admin
		fmt.Println(err)
	}

	return nil
}

func generateCumulativeResult() (error) { 

	data, err := queries.CumulativeResultData(ctx, testID)
	if err != nil {
		return err
	}
	_ = data[0]

	fmt.Println(totalPoints)
	
	xaxis := make([]int64, 11)
	yaxis := make([]int64, 11)
	xaxis[0] = 0
	yaxis[0] = 0


	curr := int64(totalPoints / 10) + 1
	for i := 1; i < 10; i++ {
		xaxis[i] = curr * int64(i)
	}
	xaxis[10] = totalPoints

	fmt.Println(xaxis)

	for _, f := range data {
		s := f.Score.Int64

		for i, x := range xaxis {
			if x >= s {
				yaxis[i] += 1
				break
			}
		}

	}

	fmt.Println(yaxis)

	xaxisStr := make([]string, 11)



	for i := range xaxis {
		xaxisStr[i] = fmt.Sprintf("%d / %d%s", xaxis[i], i * 10, "%")
	}

	yaxis[1] = 12
	yaxis[2] = 3
	yaxis[3] = 8
	yaxis[4] = 12
	yaxis[5] = 2
	yaxis[6] = 10
	yaxis[7] = 7
	yaxis[8] = 15
	yaxis[9] = 17
	yaxis[10] = 4


	resultPath, err := gocharts.CumulativeResultCharts(xaxisStr, yaxis)
	if err != nil {
		return err
	}

	fmt.Println(resultPath)
	
	return nil
}

func generateIndividualResult() error { 
	return nil 
}


func evaluate(queries *sqlc.Queries, gapi *apicalls.Caller, testID int64) (error) {

	
	
	// get the fileid or the formid of the test
	fileID, err := queries.TestFileId(ctx ,testID)
	if err != nil {
		return err
	}
	// clear the table to store answers, this is done to avoid unique constraint violation error
	// this can also be nested directly into the insert query or can be sorted with a on conflict clause
	// but it is not neccessary here, less complexity
	err = queries.ClearAnswersTable(ctx)
	if err != nil {
		return err
	}
	// this gets the complete form data including correct answers
	gForm, err := gapi.GetCompleteForm(fileID)
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

			err = queries.InsertAnswers(ctx, sqlc.InsertAnswersParams{
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
	totalPoints, err = queries.EvaluateTestResult(ctx)
	if err != nil {
		return err
	}

	generateCumulativeResult()


	return nil
}