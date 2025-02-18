package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// not much as of now
// can be updated later on



// can be used to get files from offshore storage and return it as []byte
func GetFileFromPath(url string, pathToSave string) ([]byte, error) {
	// Send GET request to fetch the file
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the file: %v", err)
	}
	defer response.Body.Close()

	// Check if the request was successful
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file, status code: %d", response.StatusCode)
	}

	// Create a new file on the local system
	outFile, err := os.Create(pathToSave)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()
	
	// Copy the content of the response body into the local file
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Println("File downloaded successfully!")

	fileByte, err := os.ReadFile("./temp")
	if err != nil {
		fmt.Println(err)
	}

	return fileByte, nil
}