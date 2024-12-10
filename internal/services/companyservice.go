package services

import (
	"fmt"

	sqlc "go.mod/internal/sqlc/generate"
)


type CompanyService struct {
	queries *sqlc.Queries
}

func NewCompanyService(queriespool *sqlc.Queries) *CompanyService {
	return &CompanyService{queries: queriespool}
}

func (c *CompanyService) CompanyFunc() {
	fmt.Println("company func")
}