package services

import (
	"go.mod/internal/apicalls"
	sqlc "go.mod/internal/sqlc/generate"
	"google.golang.org/api/drive/v3"
)

type AdminService struct {
	queries *sqlc.Queries
	GAPIService *apicalls.Caller
}
func NewAdminService(queriespool *sqlc.Queries, gapiService *apicalls.Caller) *AdminService {
	return &AdminService{
		queries: queriespool,
		GAPIService: gapiService,
	}
}

func (a *AdminService) AdminFunc() (*drive.ChangeList, error) {
	changeList, err := a.GAPIService.DriveChanges()
	if err != nil {
		return nil, err
	}

	return changeList, nil
}
