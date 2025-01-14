package services

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mod/internal/apicalls"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
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

func (a *AdminService) AdminFunc(ctx *gin.Context) (any, error) {
	changeList, err := a.queries.GetResponses(ctx, 172)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(changeList)
	var data map[string]dto.TestResponse
	err = json.Unmarshal(changeList, &data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(data)

	return data["4efc75e1"], nil
}
