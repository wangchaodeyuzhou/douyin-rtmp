@echo off
setlocal EnableDelayedExpansion

REM 抖音RTMP推流地址获取工具构建脚本 (Windows)

echo === 抖音RTMP推流地址获取工具构建脚本 ===
echo 版本: v1.0.0
echo 作者: Go版本
echo.

set PROJECT_NAME=douyin-rtmp
set VERSION=v1.0.0

REM 检查Go环境
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Go环境，请先安装Go 1.21+
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do (
    set GO_VERSION=%%i
    goto :found_version
)
:found_version
echo Go版本: %GO_VERSION%

REM 检查Npcap
echo 检查Npcap依赖...
if exist "C:\Windows\System32\Npcap" (
    echo Npcap: 已安装
) else (
    echo 警告: 未找到Npcap，请从 https://npcap.com 下载安装
)

REM 解析命令行参数
set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=build

if "%COMMAND%"=="clean" goto :clean
if "%COMMAND%"=="deps" goto :deps
if "%COMMAND%"=="test" goto :test
if "%COMMAND%"=="build" goto :build
if "%COMMAND%"=="build-all" goto :build_all
if "%COMMAND%"=="package" goto :package
if "%COMMAND%"=="help" goto :help

echo 未知命令: %COMMAND%
goto :help

:clean
echo 清理构建文件...
if exist build rmdir /s /q build
if exist %PROJECT_NAME%.exe del %PROJECT_NAME%.exe
echo 清理完成
goto :end

:deps
echo 下载Go依赖...
go mod download
go mod tidy
echo 依赖下载完成
goto :end

:test
echo 运行单元测试...
go test -v ./...
echo 测试完成
goto :end

:build
echo 下载依赖...
call :deps
echo 构建Windows版本...
if not exist build\windows_amd64 mkdir build\windows_amd64
set LDFLAGS=-s -w -X main.AppVersion=%VERSION%
go build -ldflags="%LDFLAGS%" -o build\windows_amd64\%PROJECT_NAME%.exe .

REM 复制资源文件
if exist assets xcopy /s /e assets build\windows_amd64\assets\

echo 构建完成: build\windows_amd64\%PROJECT_NAME%.exe
goto :end

:build_all
echo 构建所有平台版本...
call :deps

REM Windows amd64
echo 构建 windows/amd64...
if not exist build\windows_amd64 mkdir build\windows_amd64
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -X main.AppVersion=%VERSION%" -o build\windows_amd64\%PROJECT_NAME%.exe .

REM Linux amd64
echo 构建 linux/amd64...
if not exist build\linux_amd64 mkdir build\linux_amd64
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w -X main.AppVersion=%VERSION%" -o build\linux_amd64\%PROJECT_NAME% .

REM macOS amd64
echo 构建 darwin/amd64...
if not exist build\darwin_amd64 mkdir build\darwin_amd64
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w -X main.AppVersion=%VERSION%" -o build\darwin_amd64\%PROJECT_NAME% .

REM macOS arm64
echo 构建 darwin/arm64...
if not exist build\darwin_arm64 mkdir build\darwin_arm64
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w -X main.AppVersion=%VERSION%" -o build\darwin_arm64\%PROJECT_NAME% .

echo 所有平台构建完成
goto :end

:package
echo 创建发布包...
if not exist dist mkdir dist

REM 检查PowerShell是否可用（用于创建ZIP文件）
powershell -Command "Get-Command Compress-Archive" >nul 2>&1
if %errorlevel% equ 0 (
    REM 使用PowerShell创建ZIP文件
    for /d %%d in (build\*) do (
        echo 打包 %%~nd...
        powershell -Command "Compress-Archive -Path '%%d\*' -DestinationPath 'dist\%PROJECT_NAME%_%VERSION%_%%~nd.zip'"
    )
) else (
    echo 警告: PowerShell不可用，无法创建ZIP文件
    echo 请手动压缩 build 目录中的文件
)

echo 发布包创建完成，位于 dist 目录
goto :end

:help
echo 用法: %0 [命令]
echo.
echo 可用命令:
echo   clean         清理构建文件
echo   deps          下载Go依赖
echo   test          运行单元测试
echo   build         构建Windows版本
echo   build-all     构建所有平台版本
echo   package       创建发布包
echo   help          显示此帮助信息
echo.
echo 示例:
echo   %0 build          # 构建Windows版本
echo   %0 build-all      # 构建所有平台
echo   %0 package        # 创建发布包
goto :end

:end
echo.
pause