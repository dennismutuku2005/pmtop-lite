$repo = "dennismutuku2005/pmtop-lite"
$installDir = Join-Path $HOME ".pmtop\bin"
$shortcutPath = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\pmtop.lnk"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Write-Host "Fetching latest pmtop Suite (CLI + GUI)..." -ForegroundColor Cyan
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -like "*windows_amd64.zip" } | Select-Object -First 1

if (!$asset) {
    Write-Host "Could not find a Windows release package." -ForegroundColor Red
    exit 1
}

$tempZip = Join-Path $env:TEMP "pmtop_suite.zip"
Write-Host "Downloading $($asset.name)..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip

Write-Host "Installing to $installDir..." -ForegroundColor Cyan
Expand-Archive -Path $tempZip -DestinationPath $installDir -Force
Remove-Item $tempZip

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

