package services

import (
	"fmt"

	sqlc "go.mod/internal/sqlc/generate"
)

type StudentService struct {
	queries *sqlc.Queries
}
func NewStudentService(queriespool *sqlc.Queries) *StudentService {
	return &StudentService{queries: queriespool}
}

func (s *StudentService) StudentFunc() {
	fmt.Println("student func")
}