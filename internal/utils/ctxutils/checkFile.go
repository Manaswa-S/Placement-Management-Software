package ctxutils

import (
	"os"

	"github.com/gin-gonic/gin"
	errs "go.mod/internal/const"
)

// checkFile checks file validity/existence for the given filePath.
// It also does ctx.Set(error) and returns a structured *errs.Error object too for any errors
func CheckFile(ctx *gin.Context, filePath string) *errs.Error {

	if filePath == "" {
		ctx.Set("error", "filePath is empty string.")
		return &errs.Error{
			Type: errs.NotFound,
			Message: "filePath is empty.",
		}
	}

	_, err := os.Stat(filePath)
	if err != nil {
		ctx.Set("error", "cannot get file data for path : " + filePath)
		return &errs.Error{
			Type: errs.NotFound,
			Message: "Failed to get metadata/access file : " + filePath,
		}
	}

	return nil
}