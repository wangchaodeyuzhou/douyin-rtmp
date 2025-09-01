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
        
        self.create_widgets()
        
        # 检查 FFmpeg 状态
        self.check_ffmpeg_status()
    
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
        ffmpeg_status_label = ttk.Label(
            status_frame, 
            textvariable=self.ffmpeg_status, 
            width=12,
            foreground="green"
        )
        ffmpeg_status_label.pack(side=tk.LEFT, padx=1)
        
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
            width=12
        )
        self.start_btn.pack(side=tk.LEFT, padx=5)
        
        self.stop_btn = ttk.Button(
            control_frame1,
            text="停止推流",
            command=self.stop_stream,
            width=12,
            state="disabled"
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
            self.stream_status.set("推流中")
            self.start_btn.config(state="disabled")
            self.stop_btn.config(state="normal")
            self.log_to_stream_console(f"✓ {message}")
        else:
            messagebox.showerror("推流失败", message)
            self.log_to_stream_console(f"✗ {message}")
    
    def stop_stream(self):
        """停止推流"""
        success, message = self.streamer.stop_stream()
        
        if success:
            self.stream_status.set("未推流")
            self.start_btn.config(state="normal")
            self.stop_btn.config(state="disabled")
            self.log_to_stream_console(f"✓ {message}")
        else:
            self.log_to_stream_console(f"✗ {message}")
    
    def check_ffmpeg_status(self):
        """检查 FFmpeg 状态"""
        available, info = self.streamer.get_ffmpeg_info()
        
        if available:
            self.ffmpeg_status.set("可用")
            # 更新标签颜色为绿色
            for widget in self.main_frame.winfo_children():
                if isinstance(widget, ttk.LabelFrame) and widget.cget("text") == "推流管理":
                    for child in widget.winfo_children():
                        if isinstance(child, ttk.Frame):
                            for subchild in child.winfo_children():
                                if isinstance(subchild, ttk.Label) and subchild.cget("foreground") == "green":
                                    subchild.config(foreground="green")
            self.log_to_stream_console(f"✓ FFmpeg 可用: {info}")
        else:
            self.ffmpeg_status.set("不可用")
            # 更新标签颜色为红色
            for widget in self.main_frame.winfo_children():
                if isinstance(widget, ttk.LabelFrame) and widget.cget("text") == "推流管理":
                    for child in widget.winfo_children():
                        if isinstance(child, ttk.Frame):
                            for subchild in child.winfo_children():
                                if isinstance(subchild, ttk.Label) and "FFmpeg" in str(subchild.cget("textvariable")):
                                    subchild.config(foreground="red")
            self.log_to_stream_console(f"✗ FFmpeg 不可用: {info}")
            
            if not available:
                messagebox.showwarning(
                    "FFmpeg 不可用",
                    "未找到 FFmpeg，推流功能需要 FFmpeg 支持。\n\n"
                    "请下载并安装 FFmpeg，然后将其添加到系统环境变量中。\n"
                    "下载地址: https://ffmpeg.org/download.html"
                )
    
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
    
    def on_close(self):
        """关闭时的清理工作"""
        if self.streamer.is_streaming:
            self.streamer.stop_stream()