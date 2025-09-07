#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FFmpeg 文件准备工具
用于下载和准备 FFmpeg 文件到 resources 文件夹
"""

import os
import sys
import requests
import zipfile
import shutil
from pathlib import Path

def download_file(url, local_path, description="文件"):
    """下载文件并显示进度"""
    print(f"正在下载 {description}...")
    
    try:
        response = requests.get(url, stream=True)
        response.raise_for_status()
        
        total_size = int(response.headers.get('content-length', 0))
        downloaded = 0
        
        with open(local_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                if chunk:
                    f.write(chunk)
                    downloaded += len(chunk)
                    if total_size > 0:
                        progress = (downloaded / total_size) * 100
                        print(f"\r下载进度: {progress:.1f}% ({downloaded // 1024 // 1024}MB / {total_size // 1024 // 1024}MB)", end="")
        
        print(f"\n{description} 下载完成: {local_path}")
        return True
        
    except Exception as e:
        print(f"\n下载失败: {str(e)}")
        return False

def extract_ffmpeg_exe(zip_path, output_path):
    """从 ZIP 包中提取 ffmpeg.exe"""
    try:
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            # 查找 ffmpeg.exe
            ffmpeg_file = None
            for file_info in zip_ref.filelist:
                if file_info.filename.endswith('ffmpeg.exe'):
                    ffmpeg_file = file_info
                    break
            
            if ffmpeg_file:
                # 提取文件
                with zip_ref.open(ffmpeg_file) as source, open(output_path, 'wb') as target:
                    target.write(source.read())
                print(f"ffmpeg.exe 已提取到: {output_path}")
                return True
            else:
                print("在 ZIP 包中未找到 ffmpeg.exe")
                return False
                
    except Exception as e:
        print(f"提取失败: {str(e)}")
        return False

def main():
    """主函数"""
    print("=" * 50)
    print("FFmpeg 文件准备工具")
    print("=" * 50)
    
    # 确定项目根目录
    script_dir = Path(__file__).parent
    resources_dir = script_dir / "resources"
    
    # 创建 resources 目录
    resources_dir.mkdir(exist_ok=True)
    
    print(f"项目目录: {script_dir}")
    print(f"资源目录: {resources_dir}")
    
    # 询问用户选择
    print("\n请选择准备方式:")
    print("1. 下载并提取 ffmpeg.exe (推荐)")
    print("2. 下载完整 ZIP 包")
    print("3. 退出")
    
    choice = input("\n请输入选择 (1-3): ").strip()
    
    if choice == "1":
        # 方式1: 下载并提取单个文件
        print("\n选择了方式1: 下载并提取 ffmpeg.exe")
        
        # 下载链接
        download_url = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
        temp_zip = resources_dir / "temp_ffmpeg.zip"
        ffmpeg_exe = resources_dir / "ffmpeg.exe"
        
        # 检查是否已存在
        if ffmpeg_exe.exists():
            overwrite = input(f"\nffmpeg.exe 已存在，是否覆盖？(y/n): ").strip().lower()
            if overwrite != 'y':
                print("操作已取消")
                return
        
        # 下载文件
        if download_file(download_url, temp_zip, "FFmpeg ZIP 包"):
            # 提取 ffmpeg.exe
            if extract_ffmpeg_exe(temp_zip, ffmpeg_exe):
                # 删除临时文件
                temp_zip.unlink()
                print(f"\n✅ 成功! ffmpeg.exe 已准备到: {ffmpeg_exe}")
                print(f"文件大小: {ffmpeg_exe.stat().st_size // 1024 // 1024} MB")
            else:
                temp_zip.unlink()
                print("❌ 提取失败")
        else:
            print("❌ 下载失败")
    
    elif choice == "2":
        # 方式2: 下载完整 ZIP 包
        print("\n选择了方式2: 下载完整 ZIP 包")
        
        download_url = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
        zip_file = resources_dir / "ffmpeg-release-essentials.zip"
        
        # 检查是否已存在
        if zip_file.exists():
            overwrite = input(f"\nZIP 包已存在，是否覆盖？(y/n): ").strip().lower()
            if overwrite != 'y':
                print("操作已取消")
                return
        
        # 下载文件
        if download_file(download_url, zip_file, "FFmpeg ZIP 包"):
            print(f"\n✅ 成功! ZIP 包已下载到: {zip_file}")
            print(f"文件大小: {zip_file.stat().st_size // 1024 // 1024} MB")
        else:
            print("❌ 下载失败")
    
    elif choice == "3":
        print("\n退出程序")
        return
    
    else:
        print("\n无效选择")
        return
    
    print("\n" + "=" * 50)
    print("准备完成！现在可以构建项目了。")
    print("=" * 50)

if __name__ == "__main__":
    main()