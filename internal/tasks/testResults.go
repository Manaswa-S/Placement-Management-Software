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
var errQuota = 4

// TestResultsPoller polls the database with a fixed timeout.
// A row is selected only if its end_time has expired and result_url is NULL, which is updated once a result is generated.
// So the test poller generates only for the first time after the test has expired.
// Has an error quota that suppresses errors for some time depending upon the poller interval.
func (a *AsyncService) TestResultsPoller(ctx context.Context) error {

	timeout := config.TestResultPollerTimeout * time.Second

	fmt.Printf("Starting the test results poller : Timeout: %d\n", timeout)

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for range ticker.C {
		testID, err := a.Queries.TestResultPoller(ctx)
		if err != nil && err.Error() != errs.NoRowsMatch {
			fmt.Println(err)
			errored += 1
			if errored > errQuota {
				// TODO: raise a critical error
				return err
			}
		} else {
			// calls the generate test result draft util
			_, err := utils.GenerateResultDraft(a.Queries, a.GAPIService, testID)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}
