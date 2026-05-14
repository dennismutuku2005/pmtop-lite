$repo = "dennismutuku2005/pmtop-lite"
$installDir = Join-Path $HOME ".pmtop\bin"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Write-Host "Fetching latest version of pmtop..." -ForegroundColor Cyan
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -like "*windows_amd64.zip" } | Select-Object -First 1

if (!$asset) {
    Write-Host "Could not find a Windows release asset." -ForegroundColor Red
    exit 1
}

$tempZip = Join-Path $env:TEMP "pmtop.zip"
Write-Host "Downloading $($asset.name)..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip

Write-Host "Installing to $installDir..." -ForegroundColor Cyan
Expand-Archive -Path $tempZip -DestinationPath $installDir -Force
Remove-Item $tempZip

# Update PATH if needed
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
}

Write-Host "Success! pmtop $($release.tag_name) is installed." -ForegroundColor Green
Write-Host "Restart your terminal or run 'refreshenv' to start using it." -ForegroundColor Gray
