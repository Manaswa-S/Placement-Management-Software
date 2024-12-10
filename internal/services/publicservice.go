package services

import (
	"context"

	sqlc "go.mod/sqlc/generate"
)


type PublicService struct {
	Queries *sqlc.Queries
}

func (s *PublicService) SignupPost(ctx context.Context, signupData ) (string, error) {

	userData, err := s.Queries.SignupUser(ctx, )
		
}