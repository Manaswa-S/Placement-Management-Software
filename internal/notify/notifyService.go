package notify

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/config"
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

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>


// NewNotification inserts a new notification in the db, uses dto.NotificationData
func (n *Notify) NewNotification(ctx *gin.Context, userID int64, toSend *dto.NotificationData) (*errs.Error) {

	if toSend == nil {
		return &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "The message for notification cannot be empty or nil.",
		}
	}

	if toSend.Title == "" || toSend.Description == "" {
		return &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "The title or description for notification cannot be empty or nil.",
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
// GetNotifications gets the notifications for 'userID' with offset of 'start' and an internal limit
func (n *Notify) GetNotifications(ctx *gin.Context, userID int64, page string) (*[]sqlc.GetNotificationsRow, *errs.Error) {

	pageNum, err := strconv.ParseInt(page, 10, 32)
	if err != nil {
		return nil , &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Invalid url parameter 'page'. Should be an integer.",
			ToRespondWith: true,
		}
	}

	limit := config.NotifsConfig.GetNotifsLimit
	offset := int32(pageNum - 1) * limit

	allNotifs, err := n.Queries.GetNotifications(ctx, sqlc.GetNotificationsParams{
		UserID: userID,
		Limit: limit,
		Offset: offset,
	})
	if err != nil {
		return nil, &errs.Error{
			Type: errs.Internal,
			Message: "Failed to get notifications : " + err.Error(),
		}
	}

	if len(allNotifs) == 0 {
		return &[]sqlc.GetNotificationsRow{}, nil
	}

	return &allNotifs, nil
}
