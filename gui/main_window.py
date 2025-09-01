import tkinter as tk
from tkinter import ttk, messagebox, scrolledtext
import sys
from scapy.arch.windows import get_windows_if_list
from core.npcap import NpcapManager
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

        # 检查Npcap
        if not self.check_npcap():
            self.check_and_install_npcap()
            sys.exit(1)

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
        return self.npcap.check()

    def check_and_install_npcap(self):
        """检查并安装Npcap"""
        return self.npcap.check_and_install()

    def uninstall_npcap(self):
        """卸载Npcap"""
        self.npcap.uninstall_npcap()

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
            # 使用 NpcapManager 的方法进行安装
            if not self.npcap.check():
                self.npcap.install_npcap()
            else:
                messagebox.showinfo("提示", "Npcap 已经安装")

        except Exception as e:
            error_msg = f"启动 Npcap 安装程序失败: {str(e)}"
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
