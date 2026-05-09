@echo off
chcp 65001 >nul
echo ========================================
echo   知识库系统 - 启动脚本
echo ========================================
echo.

:: 检查 Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Go，请先安装 Go 1.21+
    pause
    exit /b 1
)
echo [OK] Go 已安装

:: 检查 Node.js
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Node.js，请先安装 Node.js 18+
    pause
    exit /b 1
)
echo [OK] Node.js 已安装

echo.
echo ========================================
echo   启动后端服务
echo ========================================
cd /d "%~dp0backend"
echo 正在安装 Go 依赖...
go mod tidy
echo 启动后端服务 http://localhost:8080 ...
start "知识库后端" go run main.go

timeout /t 3 /nobreak >nul

echo.
echo ========================================
echo   启动前端服务
echo ========================================
cd /d "%~dp0frontend"
if not exist "node_modules" (
    echo 正在安装前端依赖，首次启动可能需要几分钟...
    call npm install
)
echo 启动前端服务 http://localhost:3000 ...
start "知识库前端" npm run dev

echo.
echo ========================================
echo   启动完成！
echo   后端：http://localhost:8080
echo   前端：http://localhost:3000
echo ========================================
echo.
echo 按 Ctrl+C 可停止服务
pause