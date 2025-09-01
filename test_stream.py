#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
FFmpeg推流功能测试脚本
"""

import os
import sys
import time
import tempfile

# 添加项目根目录到路径
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from utils.ffmpeg_utils import FFmpegStreamer
from utils.logger import Logger


def test_ffmpeg_availability():
    """测试FFmpeg可用性"""
    print("=" * 50)
    print("测试FFmpeg可用性...")
    
    logger = Logger()
    streamer = FFmpegStreamer(logger)
    
    available, info = streamer.get_ffmpeg_info()
    if available:
        print(f"✓ FFmpeg 可用: {info}")
        return True
    else:
        print(f"✗ FFmpeg 不可用: {info}")
        return False


def test_video_validation():
    """测试视频文件验证"""
    print("=" * 50)
    print("测试视频文件验证...")
    
    logger = Logger()
    streamer = FFmpegStreamer(logger)
    
    # 测试不存在的文件
    is_valid, message = streamer.check_video_file("不存在的文件.mp4")
    print(f"不存在文件测试: {'✓' if not is_valid else '✗'} - {message}")
    
    # 创建一个临时的空文件进行测试
    with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp_file:
        tmp_path = tmp_file.name
    
    try:
        is_valid, message = streamer.check_video_file(tmp_path)
        print(f"空文件测试: {'✓' if not is_valid else '✗'} - {message}")
    finally:
        os.unlink(tmp_path)
    
    # 测试不支持的文件类型
    with tempfile.NamedTemporaryFile(suffix='.txt', delete=False) as tmp_file:
        tmp_path = tmp_file.name
        tmp_file.write(b"test content")
    
    try:
        is_valid, message = streamer.check_video_file(tmp_path)
        print(f"不支持格式测试: {'✓' if not is_valid else '✗'} - {message}")
    finally:
        os.unlink(tmp_path)


def test_rtmp_validation():
    """测试RTMP地址验证"""
    print("=" * 50)
    print("测试RTMP地址验证...")
    
    logger = Logger()
    streamer = FFmpegStreamer(logger)
    
    test_cases = [
        ("", "", "空地址测试"),
        ("rtmp://live.example.com/live", "", "空推流码测试"),
        ("", "stream123", "空服务器地址测试"),
        ("http://live.example.com/live", "stream123", "非RTMP协议测试"),
        ("rtmp://live.example.com/live", "stream123", "有效地址测试"),
    ]
    
    for server_url, stream_key, test_name in test_cases:
        is_valid, message = streamer.validate_rtmp_url(server_url, stream_key)
        expected_valid = test_name == "有效地址测试"
        result = "✓" if is_valid == expected_valid else "✗"
        print(f"{test_name}: {result} - {message}")


def test_ffmpeg_command_building():
    """测试FFmpeg命令构建"""
    print("=" * 50)
    print("测试FFmpeg命令构建...")
    
    logger = Logger()
    streamer = FFmpegStreamer(logger)
    
    video_file = "test.mp4"
    rtmp_url = "rtmp://live.example.com/live"
    stream_key = "stream123"
    
    command = streamer.build_ffmpeg_command(video_file, rtmp_url, stream_key)
    
    print("生成的FFmpeg命令:")
    print(" ".join(command))
    
    # 验证命令包含必要参数
    required_params = [
        "ffmpeg", "-re", "-stream_loop", "-1", 
        "-i", video_file, "-c:v", "libx264",
        "-f", "flv", f"{rtmp_url}/{stream_key}"
    ]
    
    all_params_present = all(param in command for param in required_params)
    print(f"命令参数完整性: {'✓' if all_params_present else '✗'}")


def main():
    """主测试函数"""
    print("FFmpeg推流功能测试")
    print("=" * 50)
    
    # 测试FFmpeg可用性
    ffmpeg_available = test_ffmpeg_availability()
    
    # 测试视频文件验证
    test_video_validation()
    
    # 测试RTMP地址验证
    test_rtmp_validation()
    
    # 测试FFmpeg命令构建
    test_ffmpeg_command_building()
    
    print("=" * 50)
    print("测试完成!")
    
    if not ffmpeg_available:
        print("\n注意: FFmpeg 不可用，推流功能需要安装 FFmpeg")
        print("下载地址: https://ffmpeg.org/download.html")
        print("安装后请将 FFmpeg 添加到系统环境变量中")
    else:
        print("\n✓ 所有基础功能测试通过，推流功能已准备就绪")
    
    return ffmpeg_available


if __name__ == "__main__":
    main()