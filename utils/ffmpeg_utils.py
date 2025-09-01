import os
import subprocess
import threading
import time
from datetime import datetime
from tkinter import messagebox
import psutil


class FFmpegStreamer:
    def __init__(self, logger):
        self.logger = logger
        self.process = None
        self.is_streaming = False
        self.stream_thread = None
        self.video_file = None
        self.rtmp_url = None
        self.stream_key = None
        self.log_callback = None
        self._stop_event = threading.Event()

    def set_log_callback(self, callback):
        """设置日志回调函数"""
        self.log_callback = callback

    def check_ffmpeg(self):
        """检查 FFmpeg 是否可用"""
        try:
            result = subprocess.run(
                ["ffmpeg", "-version"],
                capture_output=True,
                text=True,
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            return result.returncode == 0
        except FileNotFoundError:
            return False

    def check_video_file(self, file_path):
        """检查视频文件是否存在且可用"""
        if not os.path.exists(file_path):
            return False, "视频文件不存在"

        if not os.path.isfile(file_path):
            return False, "路径不是有效的文件"

        # 检查文件扩展名
        valid_extensions = ['.mp4', '.avi', '.mkv', '.mov', '.flv', '.wmv', '.ts', '.m4v']
        file_ext = os.path.splitext(file_path)[1].lower()
        if file_ext not in valid_extensions:
            return False, f"不支持的视频格式: {file_ext}"

        # 检查文件大小
        try:
            file_size = os.path.getsize(file_path)
            if file_size == 0:
                return False, "视频文件为空"
        except Exception as e:
            return False, f"无法读取文件信息: {str(e)}"

        return True, "视频文件检查通过"

    def validate_rtmp_url(self, server_url, stream_key):
        """验证 RTMP 推流地址和推流码"""
        if not server_url or not server_url.strip():
            return False, "推流服务器地址为空"

        if not stream_key or not stream_key.strip():
            return False, "推流码为空"

        # 检查 RTMP 地址格式
        if not server_url.startswith('rtmp://'):
            return False, "推流地址必须以 rtmp:// 开头"

        return True, "推流地址验证通过"

    def build_ffmpeg_command(self, video_file, rtmp_url, stream_key):
        """构建 FFmpeg 命令"""
        full_rtmp_url = f"{rtmp_url}/{stream_key}"
        
        command = [
            "ffmpeg",
            "-re",  # 以原始帧率读取输入
            "-stream_loop", "-1",  # 无限循环
            "-i", video_file,  # 输入文件
            "-c:v", "libx264",  # 视频编码器
            "-preset", "fast",  # 编码预设
            "-maxrate", "3000k",  # 最大比特率
            "-bufsize", "6000k",  # 缓冲区大小
            "-pix_fmt", "yuv420p",  # 像素格式
            "-g", "50",  # GOP 大小
            "-c:a", "aac",  # 音频编码器
            "-b:a", "128k",  # 音频比特率
            "-ac", "2",  # 音频声道数
            "-ar", "44100",  # 音频采样率
            "-f", "flv",  # 输出格式
            full_rtmp_url
        ]
        
        return command

    def start_stream(self, video_file, server_url, stream_key):
        """开始推流"""
        if self.is_streaming:
            return False, "推流已在进行中"

        # 检查 FFmpeg
        if not self.check_ffmpeg():
            return False, "未找到 FFmpeg，请确保已安装并配置环境变量"

        # 检查视频文件
        is_valid, message = self.check_video_file(video_file)
        if not is_valid:
            return False, f"视频文件检查失败: {message}"

        # 验证推流地址
        is_valid, message = self.validate_rtmp_url(server_url, stream_key)
        if not is_valid:
            return False, f"推流地址验证失败: {message}"

        self.video_file = video_file
        self.rtmp_url = server_url
        self.stream_key = stream_key
        self._stop_event.clear()

        # 启动推流线程
        self.stream_thread = threading.Thread(target=self._stream_process)
        self.stream_thread.daemon = True
        self.stream_thread.start()

        return True, "推流已启动"

    def stop_stream(self):
        """停止推流"""
        if not self.is_streaming:
            return False, "当前没有推流任务"

        self._stop_event.set()
        
        # 终止 FFmpeg 进程
        if self.process:
            try:
                self.process.terminate()
                # 等待进程结束
                try:
                    self.process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    # 如果进程没有正常结束，强制杀死
                    self.process.kill()
                    self.process.wait()
            except Exception as e:
                self.logger.error(f"停止推流进程时出错: {str(e)}")

        self.is_streaming = False
        self.process = None

        return True, "推流已停止"

    def _stream_process(self):
        """推流处理线程"""
        try:
            self.is_streaming = True
            command = self.build_ffmpeg_command(self.video_file, self.rtmp_url, self.stream_key)
            
            self._log(f"开始推流: {os.path.basename(self.video_file) if self.video_file else '未知文件'}")
            self._log(f"推流地址: {self.rtmp_url}")
            self._log(f"FFmpeg 命令: {' '.join(command)}")

            # 启动 FFmpeg 进程
            self.process = subprocess.Popen(
                command,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1,
                creationflags=subprocess.CREATE_NO_WINDOW
            )

            # 读取输出日志
            while not self._stop_event.is_set() and self.process.poll() is None:
                try:
                    if self.process.stdout:
                        output = self.process.stdout.readline()
                        if output:
                            output = output.strip()
                            if output:
                                self._log(f"FFmpeg: {output}")
                except Exception as e:
                    self._log(f"读取 FFmpeg 输出时出错: {str(e)}")
                    break

            # 检查进程退出状态
            if self.process.poll() is not None:
                exit_code = self.process.returncode
                if exit_code == 0:
                    self._log("推流正常结束")
                else:
                    self._log(f"推流异常结束，退出代码: {exit_code}")

        except Exception as e:
            self._log(f"推流过程中发生错误: {str(e)}")
        finally:
            self.is_streaming = False
            self.process = None
            self._log("推流线程已结束")

    def _log(self, message):
        """记录日志"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        log_message = f"[{timestamp}] {message}"
        
        if self.logger:
            self.logger.info(log_message)
        
        if self.log_callback:
            try:
                self.log_callback(log_message)
            except Exception as e:
                if self.logger:
                    self.logger.error(f"执行日志回调时出错: {str(e)}")

    def get_stream_status(self):
        """获取推流状态"""
        return {
            "is_streaming": self.is_streaming,
            "video_file": self.video_file,
            "rtmp_url": self.rtmp_url,
            "stream_key": self.stream_key,
            "process_alive": self.process and self.process.poll() is None
        }

    def get_ffmpeg_info(self):
        """获取 FFmpeg 信息"""
        try:
            result = subprocess.run(
                ["ffmpeg", "-version"],
                capture_output=True,
                text=True,
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            if result.returncode == 0:
                # 提取版本信息
                lines = result.stdout.split('\n')
                version_line = lines[0] if lines else "未知版本"
                return True, version_line
            else:
                return False, "FFmpeg 不可用"
        except FileNotFoundError:
            return False, "未找到 FFmpeg"
        except Exception as e:
            return False, f"检查 FFmpeg 时出错: {str(e)}"