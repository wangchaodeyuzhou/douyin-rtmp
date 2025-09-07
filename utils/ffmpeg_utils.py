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
        
        # 重连相关参数
        self.max_retry_count = 3  # 最大重试次数
        self.retry_delay = 5  # 重试延迟秒数
        self.current_retry = 0  # 当前重试次数
        self.use_encoding = False  # 是否使用编码模式（初始使用直接复制）
        
        # 初始化 FFmpeg 管理器
        try:
            from core.ffmpeg import FFmpegManager
            self.ffmpeg_manager = FFmpegManager(logger)
        except ImportError:
            self.ffmpeg_manager = None

    def set_log_callback(self, callback):
        """设置日志回调函数"""
        self.log_callback = callback

    def check_ffmpeg(self):
        """检查 FFmpeg 是否可用"""
        if self.ffmpeg_manager:
            return self.ffmpeg_manager.check()
        else:
            # 备用检查方法
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
        """构建 FFmpeg 命令，优化网络容错能力"""
        full_rtmp_url = f"{rtmp_url}/{stream_key}"
        
        command = [
            "ffmpeg",
            "-re",  # 以原始帧率读取输入
            "-stream_loop", "-1",  # 无限循环
            "-i", video_file,  # 输入文件
            
            # 网络容错参数（优化順序，放在输出参数前）
            "-reconnect", "1",  # 启用自动重连
            "-reconnect_at_eof", "1",  # 在文件结束时重连
            "-reconnect_streamed", "1",  # 流媒体重连
            "-reconnect_delay_max", "2",  # 最大重连延迟 (2秒)
            
            # 输出格式参数
            "-f", "flv",  # 输出格式
            "-flvflags", "no_duration_filesize",  # FLV优化选项
            "-avoid_negative_ts", "make_zero",  # 避免负时间戳
            
            # 编解码参数（优先使用直接复制，提高性能）
            "-c", "copy",  # 直接复制流，不重新编码
            
            full_rtmp_url
        ]
        
        return command
        
    def build_ffmpeg_command_with_encoding(self, video_file, rtmp_url, stream_key):
        """构建带编码的 FFmpeg 命令（备用方案）"""
        full_rtmp_url = f"{rtmp_url}/{stream_key}"
        
        command = [
            "ffmpeg",
            "-re",  # 以原始帧率读取输入
            "-stream_loop", "-1",  # 无限循环
            "-i", video_file,  # 输入文件
            
            # 网络容错参数
            "-reconnect", "1",  # 启用自动重连
            "-reconnect_at_eof", "1",  # 在文件结束时重连
            "-reconnect_streamed", "1",  # 流媒体重连
            "-reconnect_delay_max", "2",  # 最大重连延迟 (2秒)
            
            # 输出格式参数
            "-f", "flv",  # 输出格式
            "-flvflags", "no_duration_filesize",  # FLV优化选项
            "-avoid_negative_ts", "make_zero",  # 避免负时间戳
            
            # 视频编码参数 (低码率提高稳定性)
            "-c:v", "libx264",  # 视频编码器
            "-preset", "veryfast",  # 更快的编码预设，降低延迟
            "-tune", "zerolatency",  # 零延迟调优
            "-maxrate", "1500k",  # 降低最大比特率到 1.5M
            "-bufsize", "1500k",  # 降低缓冲区大小
            "-pix_fmt", "yuv420p",  # 像素格式
            "-g", "30",  # 降低GOP大小，提高容错性
            "-keyint_min", "30",  # 最小关键帧间隔
            "-sc_threshold", "0",  # 禁用场景切换检测
            
            # 音频编码参数
            "-c:a", "aac",  # 音频编码器
            "-b:a", "96k",  # 降低音频比特率
            "-ac", "2",  # 音频声道数
            "-ar", "44100",  # 音频采样率
            
            full_rtmp_url
        ]
        
        return command

    def start_stream(self, video_file, server_url, stream_key):
        """开始推流"""
        if self.is_streaming:
            return False, "推流已在进行中"

        # 检查 FFmpeg
        if not self.check_ffmpeg():
            # 静默尝试安装 FFmpeg
            if self.ffmpeg_manager:
                self.logger.info("正在尝试自动安装 FFmpeg...")
                if self.ffmpeg_manager.install_ffmpeg(silent=True):
                    # 安装成功，重新检查
                    if not self.check_ffmpeg():
                        return False, "FFmpeg 安装后仍无法使用，请重启程序后再试"
                else:
                    return False, "未找到 FFmpeg，请确保已安装并配置环境变量"
            else:
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
        
        # 重置编码模式，初始使用直接复制
        self.use_encoding = False

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
        """推流处理线程，包含重连机制"""
        self.current_retry = 0
        
        while self.current_retry <= self.max_retry_count and not self._stop_event.is_set():
            try:
                self.is_streaming = True
                
                # 根据模式选择命令
                if self.use_encoding:
                    command = self.build_ffmpeg_command_with_encoding(self.video_file, self.rtmp_url, self.stream_key)
                    mode_desc = "编码模式"
                else:
                    command = self.build_ffmpeg_command(self.video_file, self.rtmp_url, self.stream_key)
                    mode_desc = "直接复制模式"
                
                if self.current_retry == 0:
                    self._log(f"开始推流: {os.path.basename(self.video_file) if self.video_file else '未知文件'} ({mode_desc})")
                    self._log(f"推流地址: {self.rtmp_url}")
                    self._log(f"FFmpeg 命令: {' '.join(command)}")
                else:
                    self._log(f"重试推流 (第 {self.current_retry}/{self.max_retry_count} 次) - {mode_desc}")

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
                successful_frames = 0  # 成功处理的帧数
                last_frame_time = time.time()
                
                while not self._stop_event.is_set() and self.process.poll() is None:
                    try:
                        if self.process.stdout:
                            output = self.process.stdout.readline()
                            if output:
                                output = output.strip()
                                if output:
                                    self._log(f"FFmpeg: {output}")
                                    
                                    # 检测是否有成功的帧输出
                                    if "frame=" in output and "fps=" in output:
                                        successful_frames += 1
                                        last_frame_time = time.time()
                                        
                                        # 如果成功推流了一段时间，重置重试计数器
                                        if successful_frames > 100:  # 成功处理100帧后重置
                                            self.current_retry = 0
                                            successful_frames = 0
                                            
                    except Exception as e:
                        self._log(f"读取 FFmpeg 输出时出错: {str(e)}")
                        break

                # 检查进程退出状态
                if self.process.poll() is not None:
                    exit_code = self.process.returncode
                    if exit_code == 0:
                        self._log("推流正常结束")
                        break  # 正常结束，不重试
                    else:
                        self._log(f"推流异常结束，退出代码: {exit_code}")
                        
                        # 检查是否是网络错误或编码错误，如果是则尝试重连
                        if self._should_retry(exit_code) and self.current_retry < self.max_retry_count:
                            self.current_retry += 1
                            
                            # 如果是直接复制模式失败，先尝试切换到编码模式
                            if not self.use_encoding and self._is_encoding_error(exit_code):
                                self.use_encoding = True
                                self._log("直接复制模式失败，尝试切换到编码模式...")
                                # 不增加重试计数，立即重试
                                self.current_retry -= 1
                            
                            retry_msg = f"检测到错误，{self.retry_delay}秒后将进行第 {self.current_retry} 次重试..."
                            if self.use_encoding:
                                retry_msg += " (使用编码模式)"
                            self._log(retry_msg)
                            
                            # 等待一段时间后重试
                            if not self._stop_event.wait(self.retry_delay):
                                continue  # 继续重试
                            else:
                                break  # 被停止
                        else:
                            self._log("达到最大重试次数或不支持的错误，停止重试")
                            break
                else:
                    # 进程被手动停止
                    break
                    
            except Exception as e:
                self._log(f"推流过程中发生错误: {str(e)}")
                if self.current_retry < self.max_retry_count:
                    self.current_retry += 1
                    self._log(f"{self.retry_delay}秒后将进行第 {self.current_retry} 次重试...")
                    if not self._stop_event.wait(self.retry_delay):
                        continue
                    else:
                        break
                else:
                    break
            finally:
                # 清理进程
                if self.process:
                    try:
                        if self.process.poll() is None:
                            self.process.terminate()
                            try:
                                self.process.wait(timeout=3)
                            except subprocess.TimeoutExpired:
                                self.process.kill()
                                self.process.wait()
                    except Exception as e:
                        self._log(f"清理 FFmpeg 进程时出错: {str(e)}")
                    finally:
                        self.process = None
        
        # 最终清理
        self.is_streaming = False
        self.process = None
        if self.current_retry > self.max_retry_count:
            self._log(f"已达到最大重试次数 ({self.max_retry_count})，推流终止")
        else:
            self._log("推流线程已结束")
    
    def _should_retry(self, exit_code):
        """判断是否应该重试基于退出代码"""
        # Windows 错误代码 4294957243 对应 -10053 (WSAECONNABORTED)
        # 其他常见网络错误代码
        network_error_codes = [
            4294957243,  # -10053 WSAECONNABORTED
            4294957244,  # -10052 WSAENETRESET  
            4294957242,  # -10054 WSAECONNRESET
            4294957240,  # -10056 WSAEISCONN
            1,           # 通用错误
            255          # 网络超时
        ]
        
        # 编解码相关错误
        encoding_error_codes = [
            1,           # 通用错误（可能是编解码问题）
            2,           # 格式不支持
            127,         # 命令找不到
        ]
        
        return exit_code in network_error_codes or exit_code in encoding_error_codes
    
    def _is_encoding_error(self, exit_code):
        """判断是否是编解码相关错误"""
        encoding_error_codes = [
            1,    # 通用错误（可能是直接复制模式不兼容）
            2,    # 格式不支持
            69,   # 输出错误
        ]
        return exit_code in encoding_error_codes

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
        if self.ffmpeg_manager:
            return self.ffmpeg_manager.get_ffmpeg_info()
        else:
            # 备用方法
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