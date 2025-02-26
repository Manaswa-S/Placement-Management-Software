package ctxutils

import (
	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
)

// extractUserID extracts the user ID and other required parameters from the context with explicit type assertion.
// any returned error is directly included in the response as returned
func ExtractUserID(ctx *gin.Context) (int64, *errs.Error) {

	userid, exists := ctx.Get("ID")
	if !exists {
		return 0, &errs.Error{
			Type: errs.MissingRequiredField,
			Message: "Missing user ID in request.",
			ToRespondWith: true,
		}
	}

	userID, ok := userid.(int64)
	if !ok {
		return 0, &errs.Error{
			Type: errs.InvalidFormat,
			Message: "User ID of improper format.",
			ToRespondWith: true,
		}
	}

	return userID, nil 
}