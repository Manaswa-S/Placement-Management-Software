package utils

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

// nothing as of now
// can be updated later on if needed
func SaveFile(ctx *gin.Context, path string, file *multipart.FileHeader) (string, error) {

	err := ctx.SaveUploadedFile(file, path)
	if err != nil {
		return "", err
	}
	return path, nil
}