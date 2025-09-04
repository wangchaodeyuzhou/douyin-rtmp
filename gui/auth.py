import tkinter as tk
from tkinter import ttk, messagebox
from utils.auth import AuthManager

class AuthDialog:
    """密钥认证对话框"""
    
    def __init__(self, parent, title="密钥验证", auth_manager=None):
        self.parent = parent
        self.auth_manager = auth_manager or AuthManager()
        self.result = False  # 认证结果
        self.key_value = ""  # 输入的密钥
        
        # 创建对话框窗口
        self.dialog = tk.Toplevel(parent)
        self.dialog.title(title)
        self.dialog.geometry("500x550")  # 增加窗口高度以确保所有组件都能显示
        self.dialog.resizable(False, False)
        
        # 设置为模态对话框
        self.dialog.transient(parent)
        self.dialog.grab_set()
        
        # 置顶显示
        self.dialog.lift()
        self.dialog.attributes('-topmost', True)
        
        # 居中显示
        self._center_dialog()
        
        # 创建界面
        self._create_widgets()
        
        # 聚焦到密钥输入框
        self.dialog.after(100, lambda: self.key_entry.focus_set())
        
        # 绑定回车键
        self.dialog.bind('<Return>', lambda e: self._on_verify())
        self.dialog.bind('<Escape>', lambda e: self._on_cancel())
        
        # 处理窗口关闭事件
        self.dialog.protocol("WM_DELETE_WINDOW", self._on_cancel)
        
        # 取消置顶状态
        self.dialog.after(1000, lambda: self.dialog.attributes('-topmost', False))
    
    def _center_dialog(self):
        """使对话框在屏幕中心显示"""
        self.dialog.update_idletasks()
        
        # 获取屏幕尺寸
        screen_width = self.dialog.winfo_screenwidth()
        screen_height = self.dialog.winfo_screenheight()
        
        # 获取对话框尺寸
        dialog_width = 500
        dialog_height = 550
        
        # 计算居中位置
        x = (screen_width - dialog_width) // 2
        y = (screen_height - dialog_height) // 2
        
        self.dialog.geometry(f"{dialog_width}x{dialog_height}+{x}+{y}")
    
    def _create_widgets(self):
        """创建界面组件"""
        # 主框架
        main_frame = ttk.Frame(self.dialog, padding="15")
        main_frame.pack(fill=tk.BOTH, expand=True)
        
        # 标题
        title_label = ttk.Label(main_frame, text="密钥验证", font=("宋体", 16, "bold"))
        title_label.pack(pady=(0, 15))
        
        # 说明文字
        desc_text = """请输入您的密钥以验证使用权限。
密钥格式要求：
• 至少8个字符
• 包含字母和数字

重要提示：每个密钥只能使用一次！
使用后将无法重复使用。
"""
        desc_label = ttk.Label(main_frame, text=desc_text, justify=tk.LEFT, font=("宋体", 9))
        desc_label.pack(pady=(0, 10))
        
        # 密钥使用情况显示
        usage_frame = ttk.LabelFrame(main_frame, text="密钥使用情况", padding="8")
        usage_frame.pack(fill=tk.X, pady=(0, 10))
        
        # 获取使用情况
        usage_info = self.auth_manager.get_used_keys_info()
        usage_text = f"已使用密钥数量：{usage_info['total_used']}"
        if self.auth_manager.is_first_time_setup():
            usage_text += "\n状态：首次使用，欢迎！"
        
        usage_label = ttk.Label(usage_frame, text=usage_text, font=("宋体", 9), foreground="blue")
        usage_label.pack(anchor=tk.W)
        
        # 密钥输入区域
        input_frame = ttk.LabelFrame(main_frame, text="密钥输入", padding="10")
        input_frame.pack(fill=tk.X, pady=(0, 15))
        
        # 密钥标签和输入框
        key_label = ttk.Label(input_frame, text="请输入密钥:", font=("宋体", 10))
        key_label.pack(anchor=tk.W, pady=(0, 5))
        
        self.key_entry = ttk.Entry(input_frame, font=("Consolas", 12), show="*", width=40)
        self.key_entry.pack(fill=tk.X, pady=(0, 8))
        
        # 显示/隐藏密钥复选框
        self.show_key_var = tk.BooleanVar()
        show_key_cb = ttk.Checkbutton(
            input_frame, 
            text="显示密钥", 
            variable=self.show_key_var,
            command=self._toggle_key_visibility
        )
        show_key_cb.pack(anchor=tk.W)
        
        # 状态信息显示区域
        self.status_frame = ttk.LabelFrame(main_frame, text="状态信息", padding="8")
        self.status_frame.pack(fill=tk.X, pady=(0, 15))
        
        self.status_text = tk.Text(
            self.status_frame, 
            height=3, 
            wrap=tk.WORD, 
            state=tk.DISABLED,
            font=("宋体", 9)
        )
        self.status_text.pack(fill=tk.X)
        
        # 检查当前认证状态
        self._update_status_info()
        
        # 按钮区域
        button_frame = ttk.Frame(main_frame)
        button_frame.pack(fill=tk.X, pady=(10, 0))
        
        # 使用pack布局而不grid，确保按钮能正常显示
        button_container = ttk.Frame(button_frame)
        button_container.pack(anchor=tk.CENTER)
        
        self.verify_btn = ttk.Button(
            button_container, 
            text="验证", 
            command=self._on_verify,
            width=15
        )
        self.verify_btn.pack(side=tk.LEFT, padx=(0, 10))
        
        cancel_btn = ttk.Button(
            button_container, 
            text="取消", 
            command=self._on_cancel,
            width=15
        )
        cancel_btn.pack(side=tk.LEFT)
    
    def _toggle_key_visibility(self):
        """切换密钥显示/隐藏"""
        if self.show_key_var.get():
            self.key_entry.config(show="")
        else:
            self.key_entry.config(show="*")
    
    def _update_status_info(self):
        """更新状态信息显示"""
        is_valid, status_msg, auth_info = self.auth_manager.check_auth_status()
        
        self.status_text.config(state=tk.NORMAL)
        self.status_text.delete(1.0, tk.END)
        self.status_text.insert(1.0, status_msg)
        
        # 设置颜色
        if is_valid:
            self.status_text.config(fg="green")
        else:
            self.status_text.config(fg="red")
        
        self.status_text.config(state=tk.DISABLED)
    
    def _on_verify(self):
        """验证按钮点击事件"""
        key = self.key_entry.get().strip()
        
        if not key:
            messagebox.showwarning("警告", "请输入密钥")
            return
        
        # 检查密钥是否已被使用
        if self.auth_manager._is_key_used(key):
            messagebox.showerror("验证失败", "该密钥已被使用，无法重复使用\n\n请使用其他未使用的密钥。")
            return
        
        # 验证密钥
        is_valid, error_msg = self.auth_manager.authenticate_key(key)
        
        if is_valid:
            # 保存认证信息（这里会自动标记密钥为已使用）
            if self.auth_manager.save_auth_info(key):
                self.result = True
                self.key_value = key
                messagebox.showinfo("成功", f"密钥验证成功！\n\n注意：此密钥已被标记为已使用，\n无法再次使用。")
                self._update_status_info()  # 更新状态显示
                self.dialog.destroy()
            else:
                messagebox.showerror("错误", "保存认证信息失败")
        else:
            messagebox.showerror("验证失败", error_msg)
    
    def _on_cancel(self):
        """取消按钮点击事件"""
        self.result = False
        self.dialog.destroy()
    
    def show_modal(self):
        """显示模态对话框并返回结果"""
        self.dialog.wait_window()
        return self.result, self.key_value


class AuthStatusWidget:
    """认证状态显示组件"""
    
    def __init__(self, parent_frame, auth_manager=None):
        self.auth_manager = auth_manager or AuthManager()
        self.parent_frame = parent_frame
        
        # 创建状态显示区域
        self.status_frame = ttk.LabelFrame(parent_frame, text="许可证状态", padding="10")
        
        # 状态标签
        self.status_label = ttk.Label(self.status_frame, text="检查中...", font=("宋体", 10))
        self.status_label.pack(anchor=tk.W)
        
        # 详情标签
        self.detail_label = ttk.Label(self.status_frame, text="", font=("宋体", 9), foreground="gray")
        self.detail_label.pack(anchor=tk.W, pady=(5, 0))
        
        # 更新状态
        self.update_status()
    
    def pack(self, **kwargs):
        """打包显示组件"""
        self.status_frame.pack(**kwargs)
    
    def grid(self, **kwargs):
        """网格显示组件"""
        self.status_frame.grid(**kwargs)
    
    def update_status(self):
        """更新认证状态显示"""
        is_valid, status_msg, auth_info = self.auth_manager.check_auth_status()
        
        if is_valid:
            self.status_label.config(text="✓ 已授权", foreground="green")
            if auth_info:
                details = self.auth_manager.get_auth_details()
                if details:
                    remaining_days = details.get("remaining_days", 0)
                    expire_date = details.get("expire_date", "未知")
                    detail_text = f"剩余 {remaining_days} 天 (到期: {expire_date})"
                    self.detail_label.config(text=detail_text)
        else:
            self.status_label.config(text="✗ 未授权", foreground="red")
            self.detail_label.config(text="需要输入有效密钥")


def show_auth_dialog(parent, title="密钥验证"):
    """显示认证对话框的便捷函数"""
    dialog = AuthDialog(parent, title)
    return dialog.show_modal()
