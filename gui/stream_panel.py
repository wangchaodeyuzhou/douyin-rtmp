import tkinter as tk
import tkinter as tk
from tkinter import ttk, filedialog, scrolledtext, messagebox
import os
from utils.ffmpeg_utils import FFmpegStreamer


class StreamPanel:
    def __init__(self, parent, main_frame, logger):
        self.parent = parent
        self.main_frame = main_frame
        self.logger = logger
        
        # 初始化 FFmpeg 推流器
        self.streamer = FFmpegStreamer(logger)
        self.streamer.set_log_callback(self.on_stream_log)
        
        # 状态变量
        self.video_file_path = tk.StringVar()
        self.stream_status = tk.StringVar(value="未推流")
        self.ffmpeg_status = tk.StringVar(value="检查中...")
        
        # 状态监控相关
        self._status_check_job = None  # 状态检查定时任务ID
        
        self.create_widgets()
        
        # 检查 FFmpeg 状态
        self.check_ffmpeg_status()
        
        # 启动状态监控
        self.start_status_monitoring()
        
        # 初始化按钮状态（确保与实际推流状态一致）
        self.main_frame.after(100, self._initialize_button_states)
    
    def create_widgets(self):
        """创建推流面板界面"""
        # 创建推流管理面板
        stream_frame = ttk.LabelFrame(self.main_frame, text="推流管理", padding="5")
        stream_frame.grid(row=0, column=4, sticky="news", pady=5, padx=5)
        stream_frame.columnconfigure(1, weight=1)
        
        # FFmpeg 状态显示
        status_frame = ttk.Frame(stream_frame)
        status_frame.grid(row=0, column=0, columnspan=2, pady=5, sticky="we")
        
        ttk.Label(status_frame, text="FFmpeg:", width=8).pack(side=tk.LEFT, padx=2)
        self.ffmpeg_status_label = ttk.Label(
            status_frame, 
            textvariable=self.ffmpeg_status, 
            width=12,
            foreground="green"
        )
        self.ffmpeg_status_label.pack(side=tk.LEFT, padx=1)
        
        ttk.Label(status_frame, text="推流状态:", width=8).pack(side=tk.LEFT, padx=10)
        stream_status_label = ttk.Label(
            status_frame, 
            textvariable=self.stream_status, 
            width=8
        )
        stream_status_label.pack(side=tk.LEFT, padx=1)
        
        # 视频文件选择
        ttk.Label(stream_frame, text="视频文件:", width=8).grid(
            row=1, column=0, sticky=tk.W, pady=5, padx=5
        )
        
        file_frame = ttk.Frame(stream_frame)
        file_frame.grid(row=1, column=1, sticky="we", pady=5, padx=5)
        file_frame.columnconfigure(0, weight=1)
        
        self.file_entry = ttk.Entry(
            file_frame, 
            textvariable=self.video_file_path,
            state="readonly"
        )
        self.file_entry.grid(row=0, column=0, sticky="we", padx=(0, 5))
        
        ttk.Button(
            file_frame,
            text="选择",
            command=self.select_video_file,
            width=8
        ).grid(row=0, column=1)
        
        # 推流控制按钮组1
        control_frame1 = ttk.Frame(stream_frame)
        control_frame1.grid(row=2, column=0, columnspan=2, pady=(5, 2))
        
        self.start_btn = ttk.Button(
            control_frame1,
            text="开始推流",
            command=self.start_stream,
            width=12,
            state="disabled"  # 初始状态设为禁用，等待状态检查后再更新
        )
        self.start_btn.pack(side=tk.LEFT, padx=5)
        
        self.stop_btn = ttk.Button(
            control_frame1,
            text="停止推流",
            command=self.stop_stream,
            width=12,
            state="disabled"  # 初始状态设为禁用，等待状态检查后再更新
        )
        self.stop_btn.pack(side=tk.LEFT, padx=5)
        
        # 推流控制按钮组2
        control_frame2 = ttk.Frame(stream_frame)
        control_frame2.grid(row=3, column=0, columnspan=2, pady=(5, 2))
        
        ttk.Button(
            control_frame2,
            text="检查FFmpeg",
            command=self.check_ffmpeg_status,
            width=12
        ).pack(side=tk.LEFT, padx=5)
        
        ttk.Button(
            control_frame2,
            text="清除日志",
            command=self.clear_stream_log,
            width=12
        ).pack(side=tk.LEFT, padx=5)
        
        # 推流日志显示区域
        log_frame = ttk.LabelFrame(stream_frame, text="推流日志", padding="3")
        log_frame.grid(row=4, column=0, columnspan=2, sticky="news", pady=5)
        log_frame.columnconfigure(0, weight=1)
        log_frame.rowconfigure(0, weight=1)
        
        self.stream_log = scrolledtext.ScrolledText(
            log_frame,
            wrap=tk.WORD,
            height=8,
            width=30,
            font=("Consolas", 9)
        )
        self.stream_log.grid(row=0, column=0, sticky="news")
        
        # 设置推流面板的行权重
        stream_frame.rowconfigure(4, weight=1)
    
    def select_video_file(self):
        """选择视频文件"""
        file_path = filedialog.askopenfilename(
            title="选择视频文件",
            filetypes=[
                ("视频文件", "*.mp4 *.avi *.mkv *.mov *.flv *.wmv *.ts *.m4v"),
                ("MP4 文件", "*.mp4"),
                ("AVI 文件", "*.avi"),
                ("MKV 文件", "*.mkv"),
                ("MOV 文件", "*.mov"),
                ("所有文件", "*.*")
            ]
        )
        
        if file_path:
            self.video_file_path.set(file_path)
            self.logger.info(f"选择视频文件: {os.path.basename(file_path)}")
            
            # 验证视频文件
            is_valid, message = self.streamer.check_video_file(file_path)
            if is_valid:
                self.log_to_stream_console(f"✓ {message}")
            else:
                self.log_to_stream_console(f"✗ {message}")
                messagebox.showwarning("文件检查", f"视频文件检查失败:\n{message}")
    
    def start_stream(self):
        """开始推流"""
        # 检查是否已选择视频文件
        if not self.video_file_path.get():
            messagebox.showwarning("提示", "请先选择视频文件")
            return
        
        # 获取推流地址和推流码
        server_url = self.parent.server_address.get().strip()
        stream_key = self.parent.stream_code.get().strip()
        
        if not server_url or not stream_key:
            messagebox.showwarning("提示", "请先获取推流地址和推流码")
            return
        
        # 开始推流
        success, message = self.streamer.start_stream(
            self.video_file_path.get(),
            server_url,
            stream_key
        )
        
        if success:
            self.log_to_stream_console(f"✓ {message}")
            # 状态会通过监控机制自动更新
        else:
            messagebox.showerror("推流失败", message)
            self.log_to_stream_console(f"✗ {message}")
    
    def stop_stream(self):
        """停止推流"""
        success, message = self.streamer.stop_stream()
        
        if success:
            self.log_to_stream_console(f"✓ {message}")
            # 状态会通过监控机制自动更新
        else:
            self.log_to_stream_console(f"✗ {message}")
    
    def check_ffmpeg_status(self):
        """检查 FFmpeg 状态"""
        available, info = self.streamer.get_ffmpeg_info()
        
        if available:
            self.ffmpeg_status.set("可用")
            self.ffmpeg_status_label.config(foreground="green")
            self.log_to_stream_console(f"✓ FFmpeg 可用: {info}")
        else:
            self.ffmpeg_status.set("不可用")
            self.ffmpeg_status_label.config(foreground="red")
            self.log_to_stream_console(f"✗ FFmpeg 不可用: {info}")
            
            if not available:
                # 尝试安装 FFmpeg（静默安装，不显示弹窗）
                self.log_to_stream_console("正在尝试自动安装 FFmpeg...")
                if hasattr(self.parent, 'ffmpeg') and self.parent.ffmpeg:
                    # 静默安装，不显示弹窗
                    if self.parent.ffmpeg.install_ffmpeg(silent=True):
                        # 安装成功，重新检查
                        self.check_ffmpeg_status()
                        # 同步更新主窗口状态
                        if hasattr(self.parent, 'update_component_status'):
                            self.parent.update_component_status()
                    else:
                        self.log_to_stream_console("自动安装失败，请手动安装")
                else:
                    self.log_to_stream_console("请通过主菜单安装 FFmpeg")
    
    def on_stream_log(self, message):
        """处理推流日志回调"""
        def update_log():
            self.stream_log.insert(tk.END, f"{message}\n")
            self.stream_log.see(tk.END)
        
        # 确保在主线程中更新 GUI
        try:
            if self.main_frame.winfo_exists():
                self.main_frame.after_idle(update_log)
        except tk.TclError:
            pass  # 窗口已关闭
    
    def log_to_stream_console(self, message):
        """向推流控制台输出日志"""
        from datetime import datetime
        timestamp = datetime.now().strftime("%H:%M:%S")
        log_message = f"[{timestamp}] {message}"
        
        self.stream_log.insert(tk.END, f"{log_message}\n")
        self.stream_log.see(tk.END)
    
    def clear_stream_log(self):
        """清除推流日志"""
        self.stream_log.delete(1.0, tk.END)
        self.log_to_stream_console("日志已清除")
    
    def get_stream_status(self):
        """获取推流状态"""
        return self.streamer.get_stream_status()
    
    def update_ffmpeg_status(self):
        """公开的 FFmpeg 状态更新方法，供外部调用"""
        self.check_ffmpeg_status()
    
    def start_status_monitoring(self):
        """启动状态监控"""
        self._monitor_stream_status()
    
    def _initialize_button_states(self):
        """初始化按钮状态，确保与实际推流状态一致"""
        try:
            # 获取当前推流状态
            status = self.streamer.get_stream_status()
            is_streaming = status.get('is_streaming', False)
            
            # 根据实际状态初始化按钮
            self._update_button_states(is_streaming)
            
            self.log_to_stream_console(f"初始化按钮状态: {'推流中' if is_streaming else '未推流'}")
            
        except Exception as e:
            self.logger.error(f"初始化按钮状态时出错: {str(e)}")
            # 错误情况下默认设置为未推流状态
            self._update_button_states(False)
    
    def _monitor_stream_status(self):
        """监控推流状态并更新UI"""
        try:
            # 获取当前推流状态
            status = self.streamer.get_stream_status()
            is_streaming = status.get('is_streaming', False)
            
            # 更新UI状态
            self._update_button_states(is_streaming)
            
            # 继续监控（每1秒检查一次）
            self._status_check_job = self.main_frame.after(1000, self._monitor_stream_status)
            
        except tk.TclError:
            # 窗口已关闭，停止监控
            pass
        except Exception as e:
            # 发生错误，记录日志但继续监控
            self.logger.error(f"监控推流状态时出错: {str(e)}")
            self._status_check_job = self.main_frame.after(1000, self._monitor_stream_status)
    
    def _update_button_states(self, is_streaming):
        """更新按钮状态"""
        try:
            if is_streaming:
                # 正在推流
                self.stream_status.set("推流中")
                self.start_btn.config(state="disabled")
                self.stop_btn.config(state="normal")
            else:
                # 未推流或已停止
                self.stream_status.set("未推流")
                self.start_btn.config(state="normal")
                self.stop_btn.config(state="disabled")
                
        except tk.TclError:
            # 窗口已关闭，忽略错误
            pass
        except Exception as e:
            # 其他错误，记录日志
            self.logger.error(f"更新按钮状态时出错: {str(e)}")
    
    def on_close(self):
        """关闭时的清理工作"""
        # 停止状态监控
        if self._status_check_job:
            try:
                self.main_frame.after_cancel(self._status_check_job)
            except tk.TclError:
                pass
        
        # 停止推流
        if self.streamer.is_streaming:
            self.streamer.stop_stream()