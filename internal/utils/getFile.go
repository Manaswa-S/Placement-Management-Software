package utils

import (
	"errors"
	"os"
)

// not much as of now
// can be updated later on

// can be used to get files from offshore storage and return it as []byte
func GetFileFromPath(path string) (error) {
	_, err := os.Stat(path)
	if err != nil {
		return errors.New("file not found")
	}
	return nil
}