package services

import (
	"github.com/gin-gonic/gin"
	sqlc "go.mod/internal/sqlc/generate"
)


type PublicService struct {
	queries *sqlc.Queries
}

func NewPublicService(queriespool *sqlc.Queries) *PublicService {
	return &PublicService{queries: queriespool}
}

func (s *PublicService) SignupPost(ctx *gin.Context, params sqlc.SignupUserParams) (sqlc.User, error) {
	
	return s.queries.SignupUser(ctx, params)	
}