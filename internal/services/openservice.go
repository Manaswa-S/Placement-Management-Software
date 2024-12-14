package services

import (
	sqlc "go.mod/internal/sqlc/generate"
)


type OpenService struct {
	queries *sqlc.Queries
}

func NewOpenService(queriespool *sqlc.Queries) *OpenService {
	return &OpenService{queries: queriespool}
}

