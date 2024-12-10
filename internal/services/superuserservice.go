package services

import (
	"fmt"

	sqlc "go.mod/internal/sqlc/generate"
)

type SuperService struct {
	queries *sqlc.Queries
}
func NewSuperService(queriespool *sqlc.Queries) *SuperService {
	return &SuperService{queries: queriespool}
}

func (su *SuperService) SuperFunc() {
	fmt.Println("super func")
}