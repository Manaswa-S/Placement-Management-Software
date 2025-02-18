package services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mod/internal/config"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
)


type OpenService struct {
	queries *sqlc.Queries
}

func NewOpenService(queriespool *sqlc.Queries) *OpenService {
	return &OpenService{queries: queriespool}
}


func (s *OpenService) DiscussionsData(ctx *gin.Context, page string) (*[]sqlc.DiscussionsDataRow, int64, int64, error) {

	pageNo, err := strconv.ParseInt(page, 10, 64)	
	if err != nil {
		return nil, 0, 0, err
	}
	if pageNo < 1 {
		return nil, 0, 0, errors.New("page number must be greater than 0")
	}

	limit := int64(config.DiscussionPageLimit)
	
	offset := (pageNo - 1) * limit

	data, err := s.queries.DiscussionsData(ctx, sqlc.DiscussionsDataParams{
		Offset: int32(offset),
		Limit: int32(limit),
	})
	if err != nil {
		fmt.Println(err)
		return nil, 0, 0, err
	}

	return &data, limit, offset, nil
}

func (s *OpenService) NewDiscussion(ctx *gin.Context, userID int64, data *dto.NewDiscussion) *errs.Error {

	if len(data.Message) <= 75 {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "The new Discussion Message should atleast be 75 characters long.",
			ToRespondWith: true,
		}
	}

	err := s.queries.InsertDiscussion(ctx, sqlc.InsertDiscussionParams{
		UserID: userID,
		Content: data.Message,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to insert new discussion message : " + err.Error(),
		}
	}

	return nil
}

func (s *OpenService) EditDiscussion(ctx *gin.Context, userID int64, data *dto.EditDiscussion) *errs.Error {

	if len(data.Message) <= 75 {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: "The editted Discussion Message should atleast be 75 characters long.",
			ToRespondWith: true,
		}
	}

	err := s.queries.UpdateDiscussion(ctx, sqlc.UpdateDiscussionParams{
		UserID: userID,
		PostID: data.PostID,
		Content: data.Message,
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to edit discussion message : " + err.Error(),
		}
	}

	return nil
}