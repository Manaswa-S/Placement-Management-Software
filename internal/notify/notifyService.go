package notify

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	errs "go.mod/internal/const"
	"go.mod/internal/dto"
	sqlc "go.mod/internal/sqlc/generate"
)

type Notify struct {
	RedisClient *redis.Client
	Queries *sqlc.Queries
} 

func NewNotifyService(redisClient *redis.Client, queries *sqlc.Queries) *Notify {
	return &Notify{
		RedisClient: redisClient,
		Queries: queries,
	}
}

func (n *Notify) NewNotification(ctx *gin.Context, userID int64, toSend *dto.NotificationData) (*errs.Error) {

	if toSend == nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Message toSend cannot be empty.",
		}
	}

	err := n.Queries.InsertNotifications(ctx, sqlc.InsertNotificationsParams{
		UserID: userID,
		Title: pgtype.Text{String: toSend.Title, Valid: true},
		Description: pgtype.Text{String: toSend.Description, Valid: true},
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		return &errs.Error{
			Type: errs.Internal,
			Message: "Failed to insert notification into db : " + err.Error(),
		}
	}

	return nil
}

func (n *Notify) GetNotifications(ctx *gin.Context, userID int64, start string, end string) (*[]sqlc.Notification, *errs.Error) {

	pageStart, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return nil , &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse page start : " + err.Error(),
		}
	}
	_, err = strconv.ParseInt(end, 10, 64)
	if err != nil {
		return nil , &errs.Error{
			Type: errs.Internal,
			Message: "Failed to parse page end : " + err.Error(),
		}
	}

	allNotifs, err := n.Queries.GetNotifications(ctx, sqlc.GetNotificationsParams{
		UserID: userID,
		Limit: 15,
		Offset: int32(pageStart),
	})
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get notifications : " + err.Error(),
		}
	}


	return &allNotifs, nil
}
