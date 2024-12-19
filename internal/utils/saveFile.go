package utils

import (
	"errors"
	"mime/multipart"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// nothing as of now
// can be updated later on if needed
func SaveFile(ctx *gin.Context, path string, file *multipart.FileHeader) (string, error) {

	savePath := filepath.Join(path, file.Filename)
	err := ctx.SaveUploadedFile(file, savePath)
	if err != nil {
		return "", errors.New("failed to save file in storage. try again")
	}
	return savePath, nil
}