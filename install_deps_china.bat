@echo off
chcp 65001 >nul

title 抖音RTMP推流工具 - 国内镜像依赖安装

echo ==========================================
echo    使用国内镜像源安装项目依赖
echo ==========================================
echo.

:: 更新pip
echo [1/3] 更新 pip...
python -m pip install --upgrade pip -i https://pypi.tuna.tsinghua.edu.cn/simple/
if %errorLevel% neq 0 (
    echo [错误] pip 更新失败！
    pause
    exit /b 1
)
echo [✓] pip 更新成功！
echo.

:: 安装项目依赖
echo [2/3] 安装项目依赖...
pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple/
if %errorLevel% neq 0 (
    echo [错误] 依赖安装失败！
    echo.
    echo 尝试单独安装各个包:
    echo.
    pip install scapy -i https://pypi.tuna.tsinghua.edu.cn/simple/
    pip install "requests>=2.31.0" -i https://pypi.tuna.tsinghua.edu.cn/simple/
    pip install "psutil>=6.1.1" -i https://pypi.tuna.tsinghua.edu.cn/simple/
    pip install "PyYAML>=6.0" -i https://pypi.tuna.tsinghua.edu.cn/simple/
) else (
    echo [✓] 项目依赖安装成功！
)
echo.

:: 安装打包工具
echo [3/3] 安装打包工具...
set /p "choice=选择打包工具 [1]PyInstaller [2]Nuitka [3]两者都安装: "

if "%choice%"=="1" (
    pip install pyinstaller -i https://pypi.tuna.tsinghua.edu.cn/simple/
    echo [✓] PyInstaller 安装完成！
) else if "%choice%"=="2" (
    pip install ordered-set zstandard nuitka -i https://pypi.tuna.tsinghua.edu.cn/simple/
    echo [✓] Nuitka 安装完成！
) else (
    pip install pyinstaller -i https://pypi.tuna.tsinghua.edu.cn/simple/
    pip install ordered-set zstandard nuitka -i https://pypi.tuna.tsinghua.edu.cn/simple/
    echo [✓] 所有打包工具安装完成！
)

echo.
echo ==========================================
echo 安装完成！现在可以运行 build.bat 进行打包
echo ==========================================
pause