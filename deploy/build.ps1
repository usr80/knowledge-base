# Build script for Windows PowerShell
# Output: build/backend/ + build/frontend/

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"

Write-Host "=========================================="
Write-Host "  Knowledge Base - Build"
Write-Host "=========================================="
Write-Host "Output: $BuildDir"
Write-Host ""

# 1. Clean and create build directory
Write-Host "[1/4] Cleaning..." -ForegroundColor Green
if (Test-Path $BuildDir) {
    Remove-Item -Recurse -Force $BuildDir
}
New-Item -ItemType Directory -Path "$BuildDir\backend" -Force | Out-Null
New-Item -ItemType Directory -Path "$BuildDir\frontend" -Force | Out-Null

# 2. Build backend (Linux amd64)
Write-Host "[2/4] Building backend..." -ForegroundColor Green
Push-Location (Join-Path $ProjectRoot "backend")
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:GOPROXY = "https://goproxy.cn,direct"
go build -ldflags="-s -w" -o "$BuildDir\backend\knowledge-base" main.go
Pop-Location
Write-Host "  -> build/backend/knowledge-base" -ForegroundColor Gray

# 3. Build frontend
Write-Host "[3/4] Building frontend..." -ForegroundColor Green
Push-Location (Join-Path $ProjectRoot "frontend")
npm run build
Pop-Location
Write-Host "  -> build/frontend/dist/" -ForegroundColor Gray

# 4. Copy config template
Write-Host "[4/4] Copying config..." -ForegroundColor Green
Copy-Item (Join-Path $ProjectRoot "backend\.env.example") "$BuildDir\backend\.env.example"

Write-Host ""
Write-Host "=========================================="
Write-Host "  Build Complete"
Write-Host "=========================================="
Write-Host ""
Write-Host "Output:" -ForegroundColor Yellow
Get-ChildItem -Recurse $BuildDir -File | ForEach-Object {
    $rel = $_.FullName.Substring($BuildDir.Length + 1)
    Write-Host "  build/$rel ($($_.Length) bytes)" -ForegroundColor Gray
}
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Edit build/backend/.env"
Write-Host "  2. docker-compose up -d"
