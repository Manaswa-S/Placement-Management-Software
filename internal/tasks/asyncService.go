package tasks

import (
	"context"

	"go.mod/internal/apicalls"
	sqlc "go.mod/internal/sqlc/generate"
)

type AsyncService struct {
	Queries *sqlc.Queries
	GAPIService *apicalls.Caller
}

func NewAsyncService(queries *sqlc.Queries, gapiService *apicalls.Caller) *AsyncService {
	return &AsyncService{
		Queries: queries,
		GAPIService: gapiService,
	}
}

// StartAsyncs is the handler for the async services for tasks like auto-result generation, database cleanups and triggers, etc
func (a *AsyncService) StartAsyncs() error {
	// has its own context
	ctx := context.Background()

	// starts the test result poller as a go-routine
	go func() {
		err := a.TestResultsPoller(ctx)
		if err != nil {
			return
		}
	} ()



	return nil
}