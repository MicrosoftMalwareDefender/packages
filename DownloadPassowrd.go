package pack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url, fileName, destPath string) error {
	// Perform the HTTP request
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the response status is successful (200 OK)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status: %s", response.Status)
	}

	// Read the content of the response body
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	// Save the content to a file in the specified directory
	filePath := filepath.Join(destPath, fileName)
	err = ioutil.WriteFile(filePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}
