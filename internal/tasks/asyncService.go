package tasks

import (
	"context"
	"fmt"

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

func (a *AsyncService) StartAsyncs() error {

	ctx := context.Background()

	go func() {
		err := a.TestResultsPoller(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}
	} ()



	return nil
}