# ============================================
# 一键构建 + Docker 打包脚本 (Windows)
# 使用：.\build-and-deploy.ps1
# ============================================

param(
    [string]$Version = (Get-Date -Format "yyyyMMdd-HHmmss"),
    [string]$ImageName = "knowledge-base"
)

$ErrorActionPreference = "Stop"
$OutputDir = ".\dist"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  知识库系统 - 构建与打包" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "版本: $Version"
Write-Host "输出: $OutputDir"
Write-Host ""

# 1. 清理旧文件
Write-Host "[INFO] 清理旧构建文件..." -ForegroundColor Green
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null
New-Item -ItemType Directory -Path "$OutputDir\static" | Out-Null

# 2. 编译后端
Write-Host "[INFO] 编译后端..." -ForegroundColor Green
Push-Location backend
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:GOPROXY = "https://goproxy.cn,direct"
go build -ldflags="-s -w" -o "..\$OutputDir\knowledge-base" main.go
Pop-Location
Write-Host "[INFO] 后端编译完成: $OutputDir\knowledge-base" -ForegroundColor Green

# 3. 构建前端
Write-Host "[INFO] 构建前端..." -ForegroundColor Green
Push-Location frontend
npm install --registry=https://registry.npmmirror.com
npm run build
Copy-Item -Recurse -Force "dist\*" "..\$OutputDir\static\"
Pop-Location
Write-Host "[INFO] 前端构建完成: $OutputDir\static" -ForegroundColor Green

# 4. 复制 Dockerfile
Copy-Item "Dockerfile.deploy" "$OutputDir\Dockerfile"

# 5. 构建 Docker 镜像
Write-Host "[INFO] 构建 Docker 镜像..." -ForegroundColor Green
Push-Location $OutputDir
docker build -t "${ImageName}:${Version}" -t "${ImageName}:latest" .
Pop-Location
Write-Host "[INFO] Docker 镜像构建完成: ${ImageName}:${Version}" -ForegroundColor Green

# 6. 显示镜像信息
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  构建完成" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
docker images $ImageName
Write-Host ""
Write-Host "运行命令：" -ForegroundColor Yellow
Write-Host "  docker run -d -p 8080:8080 `
    -e DB_HOST=数据库地址 `
    -e DB_PASSWORD=密码 `
    -e JWT_SECRET=密钥 `
    ${ImageName}:${Version}"
Write-Host ""
Write-Host "或使用 docker-compose：" -ForegroundColor Yellow
Write-Host "  docker-compose -f docker-compose.deploy.yml up -d"
