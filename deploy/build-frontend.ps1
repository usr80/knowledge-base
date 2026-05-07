# Build frontend only (Vite/Vuetify)
# Output: build/frontend/dist/

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"
$FrontendBuild = Join-Path $BuildDir "frontend"

Write-Host "=========================================="
Write-Host "  Build Frontend (Vite)"
Write-Host "=========================================="

# Create build directory
if (-not (Test-Path $FrontendBuild)) {
    New-Item -ItemType Directory -Path $FrontendBuild -Force | Out-Null
}

# Build
Push-Location (Join-Path $ProjectRoot "frontend")

Write-Host "Installing dependencies..." -ForegroundColor Green
npm install --silent

Write-Host "Building..." -ForegroundColor Green
npm run build
Pop-Location

# Show result
$distDir = Join-Path $FrontendBuild "dist"
if (Test-Path $distDir) {
    $files = Get-ChildItem -Recurse $distDir -File
    $totalSize = ($files | Measure-Object -Property Length -Sum).Sum
    Write-Host ""
    Write-Host "Done: build/frontend/dist/ ($totalSize bytes, $($files.Count) files)" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}