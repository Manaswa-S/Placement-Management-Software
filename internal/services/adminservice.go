package services

import (
	"context"
	"fmt"

	sqlc "go.mod/sqlc/generate"
)


type AdminService struct {
	Queries *sqlc.Queries
}

func (s *AdminService) AdminS(ctx context.Context) (string, error) {

		fmt.Println("after AdminS service")
		return "AdminS", nil
}