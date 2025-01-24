package tasks

import (
	"context"
	"fmt"
	"time"

	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/utils"
)

var errored = 0

func (a *AsyncService) TestResultsPoller(ctx context.Context) error {
	fmt.Println("Starting the test results poller...")

	timeout := config.TestResultPollerTimeout * time.Second
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for range ticker.C {
		testID, err := a.Queries.TestResultPoller(ctx)
		if err != nil {
			if err.Error() != errs.NoRowsMatch {
				fmt.Println(err)
				errored += 1
				if errored > 4 {
					return err
				}
			}
		} else {
			if len(testID) > 0 {
				err = utils.GenerateTestResult(a.Queries, a.GAPIService, testID[0])
				if err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}
