package services

import (
	"context"
	"fmt"

	sqlc "go.mod/sqlc/generate"
)


type SuperUserService struct {
	Queries *sqlc.Queries
}

func (s *SuperUserService) SuperUserS(ctx context.Context) (string, error) {

		fmt.Println("after SuperUserS service")
		return "SuperUserS", nil
}