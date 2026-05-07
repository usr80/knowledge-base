# Build backend only (Linux amd64)
# Output: build/backend/knowledge-base

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"
$BackendBuild = Join-Path $BuildDir "backend"

Write-Host "=========================================="
Write-Host "  Build Backend (Linux amd64)"
Write-Host "=========================================="

# Create build directory
if (-not (Test-Path $BackendBuild)) {
    New-Item -ItemType Directory -Path $BackendBuild -Force | Out-Null
}

# Build
Push-Location (Join-Path $ProjectRoot "backend")
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:GOPROXY = "https://goproxy.cn,direct"

Write-Host "Compiling..." -ForegroundColor Green
go build -ldflags="-s -w" -o "$BackendBuild\knowledge-base" main.go
Pop-Location

# Copy config template
Copy-Item (Join-Path $ProjectRoot "backend\.env.example") "$BackendBuild\.env.example" -ErrorAction SilentlyContinue

# Show result
$outputFile = Join-Path $BackendBuild "knowledge-base"
if (Test-Path $outputFile) {
    $size = (Get-Item $outputFile).Length
    Write-Host ""
    Write-Host "Done: build/backend/knowledge-base ($size bytes)" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}