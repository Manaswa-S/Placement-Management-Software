package services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
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


func (s *OpenService) DiscussionsData(ctx *gin.Context, userID int64, page string) (*[]sqlc.DiscussionsDataRow, int32, int32, error) {

	pageNo, err := strconv.ParseInt(page, 10, 32)	
	if err != nil {
		return nil, 0, 0, err
	}
	if pageNo < 1 {
		return nil, 0, 0, errors.New("page number must be greater than 0")
	}

	limit := int32(config.DiscussionPageLimit)
	offset := int32(pageNo - 1) * limit

	data, err := s.queries.DiscussionsData(ctx, sqlc.DiscussionsDataParams{
		Offset: offset,
		Limit: limit,
		UserID: userID,
	})
	if err != nil {
		fmt.Println(err)
		return nil, 0, 0, err
	}

	return &data, limit, offset, nil
}

func (s *OpenService) NewDiscussion(ctx *gin.Context, userID int64, data *dto.NewDiscussion) *errs.Error {

	msgLen := len(data.Message)

	if msgLen < config.DiscussConfig.NewPostLowerLimit ||
		msgLen > config.DiscussConfig.NewPostUpperLimit {
		return &errs.Error{
			Type: errs.PreconditionFailed,
			Message: fmt.Sprintf("The Discussion message should have length between %d and %d characters.",
				config.DiscussConfig.NewPostLowerLimit, config.DiscussConfig.NewPostUpperLimit),
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

func (s *OpenService) NewReply(ctx *gin.Context, userID int64, data *dto.NewReply) *errs.Error {




	return nil
}

func (s *OpenService) GetReplies(ctx *gin.Context, postid string) (*[]sqlc.GetRepliesRow, *errs.Error) {

	postID, err := strconv.ParseInt(postid, 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Type: errs.InvalidFormat,
			Message: err.Error(),
			ToRespondWith: true,
		}
	}

	data, err := s.queries.GetReplies(ctx, pgtype.Int8{Int64: postID, Valid: true})
	if err != nil {
		return nil, &errs.Error{
			Type: errs.InvalidFormat,
			Message: err.Error(),
			ToRespondWith: true,
		}
	}

	return &data, nil
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