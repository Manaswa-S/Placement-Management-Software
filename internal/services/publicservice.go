package services

import (
	"context"
	"fmt"

	sqlc "go.mod/sqlc/generate"
)


type PublicService struct {
	Queries *sqlc.Queries
}

func (s *PublicService) PublicS(ctx context.Context) (string, error) {

		fmt.Println("after PublicS service")
		return "PublicS", nil
}