chcp 65001 >nul

@echo off
setlocal enabledelayedexpansion

:: 设置脚本标题
title 抖音RTMP推流工具 - 打包脚本

:: 显示打包信息
echo ==========================================
echo        抖音RTMP推流工具 打包脚本
echo ==========================================
echo.

:: 检查并显示 Python 环境信息
echo 正在检查 Python 环境...
echo ----------------------------------------

python --version >nul 2>&1
if %errorLevel% neq 0 (
    echo [错误] Python 未安装或未添加到环境变量！
    echo 请安装 Python 3.12 并确保其已添加到系统环境变量中。
    echo.
    echo 建议下载地址: https://www.python.org/downloads/
    pause
    exit /b 1
) else (
    echo [Python 版本]
    python --version
    echo [Python 路径]
    where python
)

:: 检查并显示 pip 信息
pip --version >nul 2>&1
if %errorLevel% neq 0 (
    echo [错误] pip 未安装或未正确配置！
    echo 请确保 pip 已正确安装。
    pause
    exit /b 1
) else (
    echo [pip 版本]
    pip --version
)

:: 检查必要文件
echo.
echo 正在检查项目文件...
if not exist "main.py" (
    echo [错误] 未找到 main.py 文件！
    echo 请确保在项目根目录下运行此脚本。
    pause
    exit /b 1
)

if not exist "requirements.txt" (
    echo [错误] 未找到 requirements.txt 文件！
    pause
    exit /b 1
)

if not exist "assets\logo.ico" (
    echo [警告] 未找到 logo.ico 文件，将使用默认图标。
)

echo ----------------------------------------
echo Python 环境检查通过！
echo ----------------------------------------

:: 检查命令行参数
set "skip_confirm="
set "use_nuitka="
set "clean_build="
for %%a in (%*) do (
    if "%%a"=="-y" set "skip_confirm=1"
    if "%%a"=="-yes" set "skip_confirm=1"
    if "%%a"=="nuitka" set "use_nuitka=1"
    if "%%a"=="--clean" set "clean_build=1"
    if "%%a"=="-c" set "clean_build=1"
)

:: 清理之前的构建文件
if defined clean_build (
    echo.
    echo 正在清理之前的构建文件...
    if exist "dist" (
        rmdir /s /q "dist"
        echo 已删除 dist 目录
    )
    if exist "build" (
        rmdir /s /q "build"
        echo 已删除 build 目录
    )
    if exist "nuitka_out" (
        rmdir /s /q "nuitka_out"
        echo 已删除 nuitka_out 目录
    )
    if exist "*.spec" (
        del /q "*.spec"
        echo 已删除 .spec 文件
    )
    echo 清理完成！
    echo.
)

:: 显示打包选项
echo.
echo 打包选项:
if defined use_nuitka (
    echo - 打包工具: Nuitka （高性能编译）
) else (
    echo - 打包工具: PyInstaller （快速打包）
)
echo - 目标平台: Windows x64
echo - 管理员权限: 是
echo - 包含资源: assets, resources
echo.

:: 添加确认步骤
if not defined skip_confirm (
    set /p "confirm=是否继续安装依赖并执行打包操作？(Y/N): "
    if /i not "!confirm!"=="Y" (
        echo 操作已取消。
        pause
        exit /b 0
    )
)

:: 记录开始时间
echo.
echo ==========================================
echo 开始打包 - %date% %time%
echo ==========================================

:: 检查打包方式
if defined use_nuitka (
    echo 使用 Nuitka 进行高性能编译打包...
    echo 注意: Nuitka 打包时间较长，请耐心等待...
    echo.
    
    :: 创建 nuitka 输出目录
    if not exist "nuitka_out" mkdir nuitka_out
    
    :: 安装必要的包
    echo [1/3] 正在安装依赖包...
    echo 执行命令: pip install -r requirements.txt
    pip install -r requirements.txt
    set "pip_exit_code=%errorLevel%"
    if !pip_exit_code! neq 0 (
        echo.
        echo [错误] 依赖包安装失败！错误代码: !pip_exit_code!
        echo.
        echo 可能的解决方案:
        echo 1. 检查网络连接是否正常
        echo 2. 尝试更新 pip: python -m pip install --upgrade pip
        echo 3. 尝试使用国内镜像: pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple/
        echo 4. 检查 Python 版本是否为 3.12
        echo 5. 确保以管理员权限运行命令提示符
        echo.
        echo 详细安装信息请查看上方输出...
        pause
        exit /b 1
    )
    echo [✓] 依赖包安装成功！
    
    echo [2/3] 正在安装 Nuitka 及相关工具...
    echo 执行命令: pip install ordered-set zstandard nuitka
    pip install ordered-set zstandard nuitka
    set "pip_exit_code=%errorLevel%"
    if !pip_exit_code! neq 0 (
        echo.
        echo [错误] Nuitka 安装失败！错误代码: !pip_exit_code!
        echo.
        echo 可能的解决方案:
        echo 1. 尝试单独安装: pip install nuitka
        echo 2. 检查 MinGW 是否已安装
        echo 3. 尝试使用国内镜像源
        echo.
        pause
        exit /b 1
    )
    echo [✓] Nuitka 安装成功！
    
    echo [3/3] 正在执行 Nuitka 编译打包...
    python -m nuitka --mingw64 --standalone --onefile --follow-imports ^
    --show-memory --show-progress --assume-yes-for-downloads ^
    --windows-uac-admin --windows-console-mode=disable ^
    --windows-icon-from-ico=assets/logo.ico --lto=yes ^
    --enable-plugin=tk-inter ^
    --include-data-files=assets/=assets/=**/*.* ^
    --include-data-files=resources/=resources/=**/*.* ^
    --include-data-files=api.yaml=api.yaml ^
    --output-dir=nuitka_out ^
    --output-filename=douyin-rtmp-nuitka.exe ^
    --company-name="DouyinRTMP" ^
    --product-name="抖音RTMP推流工具" ^
    --file-version="1.0.0.0" ^
    --product-version="1.0.0" ^
    --file-description="抖音直播推流地址获取工具" ^
    --copyright="Copyright (C) 2024" ^
    main.py
    
    :: 检查是否生成了可执行文件而不是错误代码
    if not exist "nuitka_out\douyin-rtmp-nuitka.exe" (
        echo [错误] Nuitka 打包失败！未找到生成的可执行文件。
        pause
        exit /b 1
    )
    
    set "output_file=nuitka_out\douyin-rtmp-nuitka.exe"
    
) else (
    echo 使用 PyInstaller 进行快速打包...
    echo.
    
    :: 安装必要的包
    echo [1/3] 正在安装依赖包...
    echo 执行命令: pip install -r requirements.txt
    pip install -r requirements.txt 2>&1 | findstr /C:"错误" /C:"ERROR" /C:"Failed" /C:"Could not" >nul
    set "pip_has_error=%errorLevel%"
    if !pip_has_error! equ 0 (
        echo.
        echo [错误] 依赖包安装过程中出现错误！
        echo.
        echo 可能的解决方案:
        echo 1. 检查网络连接是否正常
        echo 2. 尝试更新 pip: python -m pip install --upgrade pip
        echo 3. 尝试使用国内镜像: pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple/
        echo 4. 检查 Python 版本是否为 3.12
        echo 5. 确保以管理员权限运行命令提示符
        echo 6. 检查 requirements.txt 文件内容是否正确
        echo.
        echo 详细安装信息请查看上方输出...
        pause
        exit /b 1
    )
    echo [✓] 依赖包安装成功！
    
    echo [2/3] 正在安装 PyInstaller...
    echo 执行命令: pip install pyinstaller
    pip install pyinstaller 2>&1 | findstr /C:"错误" /C:"ERROR" /C:"Failed" /C:"Could not" >nul
    set "pip_has_error=%errorLevel%"
    if !pip_has_error! equ 0 (
        echo.
        echo [错误] PyInstaller 安装过程中出现错误！
        echo.
        echo 可能的解决方案:
        echo 1. 尝试更新 pip: python -m pip install --upgrade pip
        echo 2. 尝试使用国内镜像: pip install pyinstaller -i https://pypi.tuna.tsinghua.edu.cn/simple/
        echo 3. 检查磁盘空间是否充足
        echo 4. 确保以管理员权限运行
        echo.
        pause
        exit /b 1
    )
    echo [✓] PyInstaller 安装成功！
    
    echo [3/3] 正在执行 PyInstaller 打包...
    python -m PyInstaller --onefile --uac-admin --noconsole ^
    --icon=assets/logo.ico ^
    --add-data="resources;resources" ^
    --add-data="assets;assets" ^
    --add-data="api.yaml;." ^
    --name=douyin-rtmp ^
    --version-file=version_info.txt ^
    --distpath=dist ^
    --workpath=build ^
    --specpath=. ^
    --clean ^
    main.py
    
    :: 检查是否生成了可执行文件而不是错误代码
    if not exist "dist\douyin-rtmp.exe" (
        echo [错误] PyInstaller 打包失败！未找到生成的可执行文件。
        pause
        exit /b 1
    )
    
    set "output_file=dist\douyin-rtmp.exe"
)

echo.
echo ==========================================
echo 打包完成 - %date% %time%
echo ==========================================

:: 验证打包结果
if exist "!output_file!" (
    echo.
    echo [成功] 可执行文件已生成: !output_file!
    
    :: 显示文件信息
    for %%i in ("!output_file!") do (
        echo 文件大小: %%~zi 字节
        echo 修改时间: %%~ti
    )
    
    :: 询问是否立即运行
    echo.
    set /p "run_now=是否立即运行生成的可执行文件进行测试？(Y/N): "
    if /i "!run_now!"=="Y" (
        echo 正在启动程序...
        start "" "!output_file!"
    )
    
) else (
    echo [错误] 打包失败，未找到生成的可执行文件！
    pause
    exit /b 1
)

echo.
echo 提示:
echo - 首次运行可执行文件时，杀毒软件可能会报毒，请添加到白名单
echo - 程序需要管理员权限才能正常运行网络抓包功能
echo - 确保目标机器已安装或程序能自动安装 Npcap 驱动
echo.
echo 感谢使用抖音RTMP推流工具！
pause