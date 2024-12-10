package services

import (
	"fmt"

	sqlc "go.mod/internal/sqlc/generate"
)

type AdminService struct {
	queries *sqlc.Queries
}
func NewAdminService(queriespool *sqlc.Queries) *AdminService {
	return &AdminService{queries: queriespool}
}

func (a *AdminService) AdminFunc() {
	fmt.Println("admin func")
}
