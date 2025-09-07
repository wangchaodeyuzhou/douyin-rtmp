import tkinter as tk
from tkinter import ttk, messagebox, scrolledtext
import sys
from scapy.arch.windows import get_windows_if_list
from core.npcap import NpcapManager
from core.ffmpeg import FFmpegManager
from utils.logger import Logger
from utils.network import NetworkInterface
from gui.widgets import (
    create_log_panel,
    create_help_dialog,
    create_disclaimer_dialog,
    create_about_dialog,
)
from utils.config import VERSION
import threading
from utils.version import check_for_updates
from gui.obs import OBSPanel
from gui.control import ControlPanel
from gui.stream_panel import StreamPanel
from utils.resource import resource_path
from utils.config import get_config, set_config
from gui.auth import AuthStatusWidget
from utils.auth import AuthManager


class StreamCaptureGUI:
    def __init__(self, root):
        self.root = root
        self.root.title(f"推流工具 {VERSION}")
        self.root.geometry("1000x600")

        # 使窗口居中显示
        self.center_window()

        # 设置窗口图标
        try:
            icon_path = resource_path("assets/logo.ico")
            self.root.iconbitmap(icon_path)
        except tk.TclError:
            print("无法加载图标文件")

        # 初始化基础组件
        self.logger = Logger()
        self.auth_manager = AuthManager()
        
        # 状态更新相关标志
        self._components_initialized = False
        self._status_update_scheduled = False

        # 创建主框架
        self.main_frame = ttk.Frame(self.root, padding="10")
        self.main_frame.grid(row=0, column=0, sticky="news")

        # 基本UI设置
        self.setup_basic_ui()

        # 延迟执行可能导致闪现的初始化操作
        self.root.after(100, self.delayed_init)

    def delayed_init(self):
        """延迟执行的初始化操作"""
        # 初始化其他组件
        self.network = NetworkInterface(self.logger)
        self.npcap = NpcapManager(self.logger)
        self.ffmpeg = FFmpegManager(self.logger)
        
        # 标记组件已初始化
        self._components_initialized = True
        
        # 组件初始化完成后立即更新状态
        self.update_component_status()

        # 检查Npcap
        if not self.check_npcap():
            self.check_and_install_npcap()
            sys.exit(1)
            
        # 检查FFmpeg
        if not self.check_ffmpeg():
            # 静默检查和安装模式，不显示弹窗
            self.ffmpeg.check_and_install(silent=True)
            # 安装后再次更新状态
            self.update_component_status()

        # 检查更新
        self.root.after(1000, self.async_check_updates)

    def setup_basic_ui(self):
        # 设置网格权重
        self.root.grid_rowconfigure(0, weight=1)
        self.root.grid_columnconfigure(0, weight=1)
        self.main_frame.grid_columnconfigure(1, weight=1)

        # 状态变量
        self.server_address = tk.StringVar()
        self.stream_code = tk.StringVar()

        # 创建界面组件
        self.create_widgets()

    def create_widgets(self):
        # 创建菜单栏
        menubar = tk.Menu(self.root)
        self.root.config(menu=menubar)

        # 工具菜单
        tools_menu = tk.Menu(menubar, tearoff=0)
        menubar.add_cascade(label="工具", menu=tools_menu)
        tools_menu.add_command(label="安装 Npcap", command=self.install_npcap)
        tools_menu.add_command(label="卸载 Npcap", command=self.uninstall_npcap)
        tools_menu.add_separator()
        tools_menu.add_command(label="安装 FFmpeg", command=self.install_ffmpeg)
        tools_menu.add_command(label="卸载 FFmpeg", command=self.uninstall_ffmpeg)
        tools_menu.add_separator()
        tools_menu.add_command(label="密钥管理", command=self.show_auth_management)
        tools_menu.add_command(label="重新验证", command=self.reauthenticate)


        # 主布局使用网格
        self.main_frame.columnconfigure(1, weight=1)

        # 创建控制面板
        self.control_panel = ControlPanel(self)

        # 创建日志面板并保存引用
        self.log_notebook = create_log_panel(self)  # 保存notebook的引用以供后续使用

        # 添加底栏
        self.create_status_bar()

        # 添加OBS管理面板
        self.obs_panel = OBSPanel(self, self.main_frame, self.logger)

        # 添加推流面板
        self.stream_panel = StreamPanel(self, self.main_frame, self.logger)

        # 绑定窗口关闭事件
        self.root.protocol("WM_DELETE_WINDOW", self.on_closing)

    def create_status_bar(self):
        """创建底栏"""
        status_frame = ttk.Frame(self.main_frame)
        status_frame.grid(row=2, column=0, columnspan=4, sticky="we", pady=5)

        # 左侧版本信息
        ttk.Label(status_frame, text=f"版本: {VERSION}").pack(side=tk.LEFT, padx=5)
        
        # 系统组件状态
        components_frame = ttk.Frame(status_frame)
        components_frame.pack(side=tk.LEFT, padx=20)
        
        # Npcap 状态
        self.npcap_status_var = tk.StringVar(value="检查中...")
        ttk.Label(components_frame, text="Npcap:").pack(side=tk.LEFT, padx=2)
        self.npcap_status_label = ttk.Label(components_frame, textvariable=self.npcap_status_var, width=8)
        self.npcap_status_label.pack(side=tk.LEFT, padx=(0, 10))
        
        # FFmpeg 状态
        self.ffmpeg_status_var = tk.StringVar(value="检查中...")
        ttk.Label(components_frame, text="FFmpeg:").pack(side=tk.LEFT, padx=2)
        self.ffmpeg_status_label = ttk.Label(components_frame, textvariable=self.ffmpeg_status_var, width=8)
        self.ffmpeg_status_label.pack(side=tk.LEFT, padx=(0, 10))
        
        # 中间区域 - 认证状态
        auth_frame = ttk.Frame(status_frame)
        auth_frame.pack(side=tk.LEFT, padx=20)
        
        self.auth_status_widget = AuthStatusWidget(auth_frame, self.auth_manager)
        self.auth_status_widget.pack()

        # 右侧按钮组
        buttons_frame = ttk.Frame(status_frame)
        buttons_frame.pack(side=tk.RIGHT)

        # 自动检查更新复选框
        self.check_update_var = tk.BooleanVar(
            value=(
                get_config("check_update")
                if get_config("check_update") is not None
                else True
            )
        )
        check_update_cb = ttk.Checkbutton(
            buttons_frame,
            text="启动时检查更新",
            variable=self.check_update_var,
            command=self.on_check_update_changed,
        )
        check_update_cb.pack(side=tk.LEFT, padx=5)
        
        # 初始化状态显示
        self.root.after(100, self.update_component_status)


    def update_component_status(self):
        """更新系统组件状态显示"""
        # 检查组件是否已初始化，避免竞争条件
        if not self._components_initialized:
            # 如果组件还没有初始化，且没有已调度的更新，则稍后重试
            if not self._status_update_scheduled:
                self._status_update_scheduled = True
                self.root.after(200, self._retry_update_component_status)
            return
            
        # 更新 Npcap 状态
        try:
            if self.check_npcap():
                self.npcap_status_var.set("已安装")
                self.npcap_status_label.config(foreground="green")
            else:
                self.npcap_status_var.set("未安装")
                self.npcap_status_label.config(foreground="red")
        except Exception as e:
            self.logger.error(f"检查 Npcap 状态时出错: {str(e)}")
            self.npcap_status_var.set("检查失败")
            self.npcap_status_label.config(foreground="red")
        
        # 更新 FFmpeg 状态
        try:
            if self.check_ffmpeg():
                self.ffmpeg_status_var.set("已安装")
                self.ffmpeg_status_label.config(foreground="green")
            else:
                self.ffmpeg_status_var.set("未安装")
                self.ffmpeg_status_label.config(foreground="red")
        except Exception as e:
            self.logger.error(f"检查 FFmpeg 状态时出错: {str(e)}")
            self.ffmpeg_status_var.set("检查失败")
            self.ffmpeg_status_label.config(foreground="red")
            
        # 通知推流面板更新 FFmpeg 状态
        self.update_stream_panel_ffmpeg_status()
        
    def _retry_update_component_status(self):
        """重试更新组件状态"""
        self._status_update_scheduled = False  # 重置标志
        self.update_component_status()  # 再次尝试更新
        
    def update_stream_panel_ffmpeg_status(self):
        """通知推流面板更新 FFmpeg 状态"""
        if hasattr(self, 'stream_panel') and self.stream_panel:
            try:
                self.stream_panel.update_ffmpeg_status()
            except Exception as e:
                self.logger.error(f"更新推流面板 FFmpeg 状态时出错: {str(e)}")
    
    def show_donation(self):
        """显示打赏对话框"""
        from gui.widgets import create_donation_dialog

        create_donation_dialog(self.root, self.logger, resource_path)

    def log_to_console(self, message):
        """输出日志到控制台"""
        self.logger.info(message)

    def clear_console(self):
        """清除控制台内容"""
        self.logger.clear_console()

    def check_npcap(self):
        """检查Npcap是否已安装"""
        if not hasattr(self, 'npcap') or self.npcap is None:
            return False
        return self.npcap.check()

    def check_and_install_npcap(self):
        """检查并安装Npcap"""
        if not hasattr(self, 'npcap') or self.npcap is None:
            return False
        return self.npcap.check_and_install()

    def uninstall_npcap(self):
        """卸载Npcap"""
        if not hasattr(self, 'npcap') or self.npcap is None:
            messagebox.showerror("错误", "Npcap 组件未初始化")
            return
        self.npcap.uninstall_npcap()
        
    def check_ffmpeg(self):
        """检查FFmpeg是否已安装"""
        if not hasattr(self, 'ffmpeg') or self.ffmpeg is None:
            return False
        return self.ffmpeg.check()

    def check_and_install_ffmpeg(self):
        """检查并安装FFmpeg"""
        if not hasattr(self, 'ffmpeg') or self.ffmpeg is None:
            return False
        return self.ffmpeg.check_and_install()

    def uninstall_ffmpeg(self):
        """卸载FFmpeg"""
        if not hasattr(self, 'ffmpeg') or self.ffmpeg is None:
            messagebox.showerror("错误", "FFmpeg 组件未初始化")
            return
        
        # 调用 FFmpeg 管理器的卸载方法
        self.ffmpeg.uninstall_ffmpeg()
        
        # 卸载后更新状态（无论成功失败）
        self.update_component_status()

    def async_check_updates(self):
        """异步检查更新"""
        # 检查是否启用自动更新，默认为 True
        if get_config("check_update") is not None:
            check_update = get_config("check_update")
        else:
            check_update = True
            
        if not check_update:
            return

        thread = threading.Thread(target=check_for_updates)
        thread.daemon = True
        thread.start()

    def install_npcap(self):
        """手动安装 Npcap"""
        try:
            # 检查组件是否初始化
            if not hasattr(self, 'npcap') or self.npcap is None:
                messagebox.showerror("错误", "Npcap 组件未初始化，请稍后再试")
                return
                
            # 使用 NpcapManager 的方法进行安装
            if not self.npcap.check():
                self.npcap.install_npcap()
            else:
                messagebox.showinfo("提示", "Npcap 已经安装")

        except Exception as e:
            error_msg = f"启动 Npcap 安装程序失败: {str(e)}"
            self.logger.error(error_msg)
            messagebox.showerror("错误", error_msg)
            
    def install_ffmpeg(self):
        """手动安装 FFmpeg"""
        try:
            # 检查组件是否初始化
            if not hasattr(self, 'ffmpeg') or self.ffmpeg is None:
                messagebox.showerror("错误", "FFmpeg 组件未初始化，请稍后再试")
                return
                
            # 使用 FFmpegManager 的方法进行安装
            if not self.ffmpeg.check():
                if self.ffmpeg.install_ffmpeg():
                    # 安装成功，更新状态
                    self.update_component_status()
                    messagebox.showinfo("成功", "FFmpeg 安装成功！")
                else:
                    # 安装失败，但仍需更新状态
                    self.update_component_status()
                    messagebox.showerror("失败", "FFmpeg 安装失败，请检查网络连接或手动安装")
            else:
                messagebox.showinfo("提示", "FFmpeg 已经安装")

        except Exception as e:
            error_msg = f"启动 FFmpeg 安装程序失败: {str(e)}"
            self.logger.error(error_msg)
            messagebox.showerror("错误", error_msg)

    def show_help(self):
        """显示使用说明弹窗"""
        create_help_dialog(self.root)

    def show_disclaimer(self):
        """显示免责声明弹窗"""
        create_disclaimer_dialog(self.root)

    def clear_logs(self):
        """清除所有日志"""
        self.logger.clear_console()
        self.logger.clear_packet_console()

    def clear_packet_console(self):
        """清除数据包控制台内容"""
        self.logger.clear_packet_console()
        self.logger.info("数据包日志已清除")  # 在主控制台显示清除提示

    def center_window(self):
        """使窗口在屏幕中心显示"""
        # 获取屏幕宽度和高度
        screen_width = self.root.winfo_screenwidth()
        screen_height = self.root.winfo_screenheight()

        # 获取窗口宽度和高度
        window_width = 1000
        window_height = 600

        # 计算窗口居中的x和y坐标
        x = (screen_width - window_width) // 2
        y = (screen_height - window_height) // 2

        # 设置窗口位置
        self.root.geometry(f"{window_width}x{window_height}+{x}+{y}")

    def on_check_update_changed(self):
        """处理自动检查更新复选框状态变化"""
        set_config("check_update", self.check_update_var.get())
    
    def show_auth_management(self):
        """显示认证管理对话框"""
        from gui.auth import show_auth_dialog
        
        # 检查当前认证状态
        is_authenticated, status_msg, auth_info = self.auth_manager.check_auth_status()
        
        if is_authenticated:
            # 已认证，显示详细信息
            details = self.auth_manager.get_auth_details()
            if details:
                activate_date = details.get('activate_date', '未知')
                expire_date = details.get('expire_date', '未知')
                remaining_days = details.get('remaining_days', 0)
                
                info_msg = f"""当前认证状态：已授权

激活时间：{activate_date}
到期时间：{expire_date}
剩余天数：{remaining_days} 天"""
                
                messagebox.showinfo("认证状态", info_msg)
            else:
                messagebox.showinfo("认证状态", "已认证，但无法获取详细信息")
        else:
            # 未认证，显示认证对话框
            messagebox.showwarning("认证状态", f"未认证：{status_msg}")
            self.reauthenticate()
    
    def reauthenticate(self):
        """重新认证"""
        from gui.auth import show_auth_dialog
        
        result = messagebox.askyesno("重新验证", "是否要重新进行密钥验证？")
        if result:
            auth_result, key = show_auth_dialog(self.root, "重新验证")
            if auth_result:
                messagebox.showinfo("成功", "重新验证成功！")
                # 更新认证状态显示
                if hasattr(self, 'auth_status_widget'):
                    self.auth_status_widget.update_status()
            else:
                messagebox.showinfo("取消", "重新验证已取消")

    def check_updates_manually(self):
        """手动检查更新"""
        self.logger.info("正在检查更新...")

        def check_update_with_feedback():
            has_update, clicked_yes = check_for_updates()
            if not has_update:
                self.root.after(
                    0,
                    lambda: messagebox.showinfo(
                        "检查更新", "您当前使用的已经是最新版本！"
                    ),
                )
                self.logger.info("当前已是最新版本")

            if not clicked_yes:
                self.logger.info("发现了最新版本，但您取消了更新")

        thread = threading.Thread(target=check_update_with_feedback)
        thread.daemon = True
        thread.start()

    def on_closing(self):
        """窗口关闭时的清理工作"""
        # 停止推流
        if hasattr(self, 'stream_panel'):
            self.stream_panel.on_close()
        
        # 停止捕获
        if hasattr(self, 'control_panel') and self.control_panel.is_capturing:
            self.control_panel.toggle_capture()
        
        # 关闭窗口
        self.root.destroy()
