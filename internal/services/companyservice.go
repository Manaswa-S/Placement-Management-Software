package services

import (
	"context"
	"fmt"

	sqlc "go.mod/sqlc/generate"
)


type CompanyService struct {
	Queries *sqlc.Queries
}

func (s *CompanyService) CompanyS(ctx context.Context) (string, error) {

		fmt.Println("after CompanyS service")
		return "CompanyS", nil
}