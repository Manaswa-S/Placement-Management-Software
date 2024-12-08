package services

import (
	"context"
	"fmt"

	sqlc "go.mod/sqlc/generate"
)


type StudentService struct {
	Queries *sqlc.Queries
}

func (s *StudentService) StudentS(ctx context.Context) (string, error) {

		fmt.Println("after StudentS service")
		return "StudentS", nil
}