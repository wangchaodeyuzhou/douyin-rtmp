import tkinter as tk
from tkinter import messagebox
from gui.main_window import StreamCaptureGUI
from gui.auth import show_auth_dialog
from utils.system import is_admin, run_as_admin
from utils.auth import AuthManager

def main():
    # 检查是否以管理员权限运行
    run_as_admin()
    print("检查管理员权限...")
    if not is_admin():
        # 创建临时窗口用于显示错误信息
        temp_root = tk.Tk()
        temp_root.withdraw()
        messagebox.showerror("错误", "请以管理员权限运行此程序！")
        temp_root.destroy()
        return
    print("管理员权限检查通过")

    # 检查认证状态（在创建主窗口之前）
    print("检查认证状态...")
    auth_manager = AuthManager()
    is_authenticated, status_msg, auth_info = auth_manager.check_auth_status()
    print(f"认证状态: {is_authenticated}, 信息: {status_msg}")
    
    if not is_authenticated:
        # 需要进行认证，创建临时窗口用于认证对话框
        print("检测到未认证，将显示认证对话框...")
        
        # 创建临时窗口用于认证
        auth_root = tk.Tk()
        auth_root.withdraw()  # 隐藏临时窗口
        
        # 显示提示信息
        messagebox.showinfo("提示", "首次使用需要进行密钥验证\n\n")
        
        # 显示认证对话框
        print("正在显示认证对话框...")
        auth_result, key = show_auth_dialog(auth_root, "密钥验证")
        print(f"认证结果: {auth_result}")
        
        if not auth_result:
            print("用户取消认证，程序退出")
            messagebox.showinfo("退出", "未通过验证，程序将退出")
            auth_root.destroy()
            return
        
        print("认证成功！")
        messagebox.showinfo("成功", "密钥验证成功，欢迎使用！")
        
        # 销毁临时认证窗口
        auth_root.destroy()

    else:
        print("用户已认证，直接进入程序")

    # 认证通过后，创建主窗口
    print("创建主窗口...")
    root = tk.Tk()
    
    # 启动主程序
    print("启动主程序界面...")
    StreamCaptureGUI(root)
    root.mainloop()

if __name__ == "__main__":
    main()