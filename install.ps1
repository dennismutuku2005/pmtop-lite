$repo = "dennismutuku2005/pmtop-lite"
$installDir = Join-Path $HOME ".pmtop\bin"
$shortcutPath = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\pmtop.lnk"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Stop any running instances to avoid "File in Use" errors
$runningProcs = Get-Process -Name "pmtop", "pmtop-gui" -ErrorAction SilentlyContinue
if ($runningProcs) {
    Write-Host "Closing running instances of pmtop to prepare for update..." -ForegroundColor Yellow
    $runningProcs | Stop-Process -Force
    Start-Sleep -Seconds 1
}

Write-Host "Fetching latest pmtop Suite (CLI + GUI)..." -ForegroundColor Cyan
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"

$cliAsset = $release.assets | Where-Object { $_.name -like "pmtop-cli*windows_amd64.zip" } | Select-Object -First 1
$guiAsset = $release.assets | Where-Object { $_.name -like "pmtop-gui*windows_amd64.zip" } | Select-Object -First 1

if (!$cliAsset -or !$guiAsset) {
    Write-Host "Could not find all required Windows release packages." -ForegroundColor Red
    exit 1
}

Write-Host "Cleaning old installation..." -ForegroundColor Cyan
if (Test-Path $installDir) {
    Remove-Item (Join-Path $installDir "*") -Force -Recurse -ErrorAction SilentlyContinue
} else {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Install CLI
$tempCli = Join-Path $env:TEMP "pmtop_cli.zip"
Write-Host "Downloading CLI ($($cliAsset.name))..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $cliAsset.browser_download_url -OutFile $tempCli
Expand-Archive -Path $tempCli -DestinationPath $installDir -Force
Remove-Item $tempCli

# Install GUI
$tempGui = Join-Path $env:TEMP "pmtop_gui.zip"
Write-Host "Downloading GUI ($($guiAsset.name))..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $guiAsset.browser_download_url -OutFile $tempGui
Expand-Archive -Path $tempGui -DestinationPath $installDir -Force
Remove-Item $tempGui



# Create Start Menu Shortcut for the GUI
Write-Host "Registering pmtop as a Windows App..." -ForegroundColor Green
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($shortcutPath)
$Shortcut.TargetPath = Join-Path $installDir "pmtop-gui.exe"
$Shortcut.WorkingDirectory = $installDir
$Shortcut.Description = "pmtop — Modern Port & Process Monitor"
$Shortcut.Save()

# Update PATH if needed
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
}

Write-Host "Success! pmtop suite is ready." -ForegroundColor Green
Write-Host "- Search 'pmtop' in Start Menu to launch GUI" -ForegroundColor Cyan
Write-Host "- Type 'pmtop' in terminal for CLI" -ForegroundColor Cyan

