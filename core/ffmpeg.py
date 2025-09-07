import os
import sys
import webbrowser
import subprocess
import zipfile
import requests
from tkinter import messagebox
import tempfile
import shutil
import time


class FFmpegManager:
    def __init__(self, logger):
        self.logger = logger
        self._last_check_result = None  # 缓存上次检查结果，避免重复日志
        """获取资源文件的绝对路径，兼容开发、Nuitka打包和PyInstaller打包后的环境"""
        if hasattr(sys, "_MEIPASS"):
            # 在PyInstaller打包后的环境中
            application_path = sys._MEIPASS
        else:
            # 在Nuitka打包后的环境或开发环境中
            application_path = os.path.dirname(os.path.dirname(__file__))

        self.installer_path = os.path.join(
            application_path, "resources", "ffmpeg-release-essentials.zip"
        )
        
        # FFmpeg 可执行文件路径（类似 Npcap 方式）
        self.ffmpeg_exe_path = os.path.join(
            application_path, "resources", "ffmpeg.exe"
        )
        
        # FFmpeg 完整工具包路径
        self.ffmpeg_tools_path = os.path.join(
            application_path, "resources", "ffmpeg-tools"
        )
        
        # FFmpeg 多个下载源
        self.ffmpeg_download_sources = [
            {
                "name": "官方构建版本",
                "url": "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
                "size": "约80MB"
            },
            {
                "name": "备用源1 (GitHub)", 
                "url": "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
                "size": "约120MB"
            },
            {
                "name": "备用源2 (精简版)",
                "url": "https://github.com/GyanD/codexffmpeg/releases/download/6.1/ffmpeg-6.1-essentials_build.zip",
                "size": "约65MB"
            }
        ]
        self.ffmpeg_web_url = "https://ffmpeg.org/download.html"

        # 添加路径检查日志（仅在初始化时输出一次）
        if os.path.exists(self.ffmpeg_exe_path):
            self.logger.info("找到 FFmpeg 可执行文件")
        elif os.path.exists(self.ffmpeg_tools_path):
            self.logger.info("找到 FFmpeg 完整工具包")
        elif os.path.exists(self.installer_path):
            self.logger.info("找到 FFmpeg 安装包")
        else:
            self.logger.info(
                f"未找到本地 FFmpeg 文件，将使用在线下载"
            )

    def check(self):
        """检查FFmpeg是否已安装"""
        try:
            result = subprocess.run(
                ["ffmpeg", "-version"],
                capture_output=True,
                text=True,
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            is_available = result.returncode == 0
            
            # 只在状态发生变化时才输出日志
            if self._last_check_result != is_available:
                if is_available:
                    self.logger.info("检测到FFmpeg已安装")
                else:
                    self.logger.error("FFmpeg安装异常")
                self._last_check_result = is_available
                
            return is_available
        except FileNotFoundError:
            # 只在状态变化时输出日志
            if self._last_check_result != False:
                self.logger.error("未检测到FFmpeg")
                self._last_check_result = False
            return False
        except Exception as e:
            # 只在状态变化时输出日志
            if self._last_check_result != False:
                self.logger.error(f"检查FFmpeg时出错: {str(e)}")
                self._last_check_result = False
            return False

    def check_and_install(self, silent=False):
        """检查并安装FFmpeg"""
        if self.check():
            return True

        if silent:
            # 静默模式，不显示弹窗，直接尝试安装
            try:
                # 优先检查本地文件（按优先级排序）
                if self._install_from_local_files():
                    self.logger.info("FFmpeg 自动安装成功")
                    return True
                else:
                    self.logger.info("本地安装失败，尝试在线下载")
                    return self._download_and_install(silent=True)
            except Exception as e:
                error_msg = f"自动安装FFmpeg失败: {str(e)}"
                self.logger.error(error_msg)
                return False
        else:
            # 交互模式，显示确认对话框
            response = messagebox.askquestion(
                "缺少必要组件",
                "检测到未安装FFmpeg，该组件是推流功能必需的。\n是否立即安装？\n(需要管理员权限)",
                icon="warning",
            )

            if response == "yes":
                try:
                    # 优先检查本地文件（按优先级排序）
                    if self._install_from_local_files():
                        return True
                    else:
                        self.logger.info("本地安装失败，尝试在线下载")
                        return self._download_and_install()
                except Exception as e:
                    error_msg = f"安装FFmpeg失败: {str(e)}"
                    self.logger.error(error_msg)
                    messagebox.showerror("错误", error_msg)
                    self.show_ffmpeg_warning()
            else:
                self.show_ffmpeg_warning()
        return False

    def _show_download_source_dialog(self):
        """显示下载源选择对话框"""
        import tkinter as tk
        from tkinter import ttk
        
        result = [None]  # 使用列表保存结果以便在闭包中修改
        
        def on_select(index):
            result[0] = index
            dialog.destroy()
            
        def on_cancel():
            dialog.destroy()
        
        # 创建对话框
        dialog = tk.Toplevel()
        dialog.title("选择下载源")
        dialog.geometry("400x300")
        dialog.resizable(False, False)
        
        # 居中显示
        dialog.update_idletasks()
        x = (dialog.winfo_screenwidth() // 2) - (400 // 2)
        y = (dialog.winfo_screenheight() // 2) - (300 // 2)
        dialog.geometry(f"400x300+{x}+{y}")
        
        # 设置为模态对话框
        dialog.transient()
        dialog.grab_set()
        
        # 主框架
        main_frame = ttk.Frame(dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)
        
        # 标题
        title_label = ttk.Label(main_frame, text="选择 FFmpeg 下载源", font=("宋体", 12, "bold"))
        title_label.pack(pady=(0, 15))
        
        # 说明文字
        desc_label = ttk.Label(main_frame, text="请选择一个下载源：", font=("宋体", 10))
        desc_label.pack(pady=(0, 10))
        
        # 下载源选项
        for i, source in enumerate(self.ffmpeg_download_sources):
            frame = ttk.Frame(main_frame)
            frame.pack(fill=tk.X, pady=5)
            
            btn = ttk.Button(
                frame, 
                text=f"{source['name']} ({source['size']})",
                command=lambda idx=i: on_select(idx),
                width=30
            )
            btn.pack()
        
        # 取消按钮
        cancel_btn = ttk.Button(main_frame, text="取消", command=on_cancel, width=15)
        cancel_btn.pack(pady=(20, 0))
        
        # 等待用户选择
        dialog.wait_window()
        
        return result[0]
        
    def _install_from_local_files(self):
        """从本地文件安装 FFmpeg（按优先级排序）"""
        # 方案 1：直接使用 ffmpeg.exe 可执行文件
        if os.path.exists(self.ffmpeg_exe_path):
            self.logger.info("找到本地 FFmpeg 可执行文件，开始安装...")
            if self._install_from_single_exe():
                return True
        
        # 方案 2：使用完整工具包文件夹
        if os.path.exists(self.ffmpeg_tools_path):
            self.logger.info("找到 FFmpeg 完整工具包，开始安装...")
            if self._install_from_tools_folder():
                return True
        
        # 方案 3：使用 ZIP 安装包
        if os.path.exists(self.installer_path):
            self.logger.info(f"找到 FFmpeg ZIP 安装包，开始安装: {self.installer_path}")
            if self._install_ffmpeg_from_zip(self.installer_path):
                return True
        
        self.logger.info("未找到任何本地 FFmpeg 文件")
        return False
    
    def _install_from_single_exe(self):
        """从单个 ffmpeg.exe 文件安装"""
        try:
            # 创建安装目录
            install_dir = self._get_install_directory()
            if not install_dir:
                return False
            
            bin_dir = os.path.join(install_dir, "bin")
            os.makedirs(bin_dir, exist_ok=True)
            
            # 复制 ffmpeg.exe 到安装目录
            target_path = os.path.join(bin_dir, "ffmpeg.exe")
            shutil.copy2(self.ffmpeg_exe_path, target_path)
            
            self.logger.info(f"FFmpeg 已复制到: {target_path}")
            
            # 添加到环境变量
            if self._add_to_path(bin_dir):
                self.logger.info(f"FFmpeg 安装成功: {target_path}")
                return True
            else:
                self.logger.error("添加到 PATH 失败")
                return False
                
        except Exception as e:
            self.logger.error(f"从单个文件安装 FFmpeg 失败: {str(e)}")
            return False
    
    def _install_from_tools_folder(self):
        """从工具文件夹安装"""
        try:
            # 创建安装目录
            install_dir = self._get_install_directory()
            if not install_dir:
                return False
            
            # 复制整个工具文件夹
            if os.path.exists(install_dir):
                shutil.rmtree(install_dir)
            
            shutil.copytree(self.ffmpeg_tools_path, install_dir)
            self.logger.info(f"FFmpeg 工具包已复制到: {install_dir}")
            
            # 查找 ffmpeg.exe
            ffmpeg_exe = None
            for root, dirs, files in os.walk(install_dir):
                if "ffmpeg.exe" in files:
                    ffmpeg_exe = os.path.join(root, "ffmpeg.exe")
                    break
            
            if not ffmpeg_exe:
                self.logger.error("在工具包中未找到 ffmpeg.exe")
                return False
            
            # 添加到环境变量
            ffmpeg_bin_dir = os.path.dirname(ffmpeg_exe)
            if self._add_to_path(ffmpeg_bin_dir):
                self.logger.info(f"FFmpeg 安装成功: {ffmpeg_exe}")
                return True
            else:
                self.logger.error("添加到 PATH 失败")
                return False
                
        except Exception as e:
            self.logger.error(f"从工具文件夹安装 FFmpeg 失败: {str(e)}")
            return False
    
    def _get_install_directory(self):
        """获取安装目录"""
        try:
            # 尝试系统目录
            install_dir = r"C:\Program Files\FFmpeg"
            os.makedirs(install_dir, exist_ok=True)
            return install_dir
        except PermissionError:
            # 没有权限，使用用户目录
            install_dir = os.path.join(os.path.expanduser("~"), "FFmpeg")
            os.makedirs(install_dir, exist_ok=True)
            self.logger.info(f"使用用户目录安装: {install_dir}")
            return install_dir
        except Exception as e:
            self.logger.error(f"创建安装目录失败: {str(e)}")
            return None
    
    def _download_with_retry(self, url, file_path, max_retries=3, chunk_size=8192):
        """带重试机制的下载方法"""
        for attempt in range(max_retries):
            try:
                self.logger.info(f"尝试下载 (第 {attempt + 1}/{max_retries} 次)...")
                
                # 配置请求参数
                session = requests.Session()
                session.headers.update({
                    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
                })
                
                # 发起请求
                with session.get(url, stream=True, timeout=(30, 60)) as response:
                    response.raise_for_status()
                    total_size = int(response.headers.get('content-length', 0))
                    downloaded = 0
                    last_progress = 0
                    
                    with open(file_path, 'wb') as f:
                        for chunk in response.iter_content(chunk_size=chunk_size):
                            if chunk:
                                f.write(chunk)
                                downloaded += len(chunk)
                                
                                # 更新进度显示
                                if total_size > 0:
                                    progress = int((downloaded / total_size) * 100)
                                    if progress >= last_progress + 5:  # 每5%显示一次
                                        self.logger.info(f"下载进度: {progress}% ({downloaded // 1024 // 1024}MB / {total_size // 1024 // 1024}MB)")
                                        last_progress = progress
                
                self.logger.info("下载完成")
                return True
                
            except requests.exceptions.ConnectionError as e:
                self.logger.error(f"第 {attempt + 1} 次下载失败: 网络连接错误 - {str(e)}")
                if attempt < max_retries - 1:
                    wait_time = (attempt + 1) * 2  # 递增等待时间
                    self.logger.info(f"等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                else:
                    self.logger.error("所有重试都失败了")
                    return False
                    
            except requests.exceptions.Timeout as e:
                self.logger.error(f"第 {attempt + 1} 次下载失败: 连接超时 - {str(e)}")
                if attempt < max_retries - 1:
                    wait_time = (attempt + 1) * 2
                    self.logger.info(f"等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                else:
                    return False
                    
            except Exception as e:
                self.logger.error(f"第 {attempt + 1} 次下载失败: {str(e)}")
                if attempt < max_retries - 1:
                    wait_time = (attempt + 1) * 2
                    self.logger.info(f"等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                else:
                    return False
        
        return False
        
    def _download_and_install(self, silent=False):
        """下载并安装FFmpeg"""
        try:
            if silent:
                # 静默模式，使用第一个可用的下载源
                for source_choice, download_info in enumerate(self.ffmpeg_download_sources):
                    download_url = download_info["url"]
                    self.logger.info(f"尝试从 {download_info['name']} 下载 FFmpeg...")
                    
                    # 创建临时目录
                    temp_dir = tempfile.mkdtemp()
                    zip_path = os.path.join(temp_dir, "ffmpeg.zip")
                    
                    try:
                        # 使用改进的下载方法
                        if self._download_with_retry(download_url, zip_path):
                            self.logger.info("下载完成，开始安装...")
                            
                            # 安装下载的文件
                            if self._install_ffmpeg_from_zip(zip_path):
                                return True
                            else:
                                self.logger.error(f"从 {download_info['name']} 下载成功但安装失败")
                        else:
                            self.logger.error(f"从 {download_info['name']} 下载失败")
                    finally:
                        # 清理临时文件
                        try:
                            shutil.rmtree(temp_dir)
                        except:
                            pass
                
                # 所有源都失败
                self.logger.error("所有下载源都尝试失败")
                return False
            else:
                # 交互模式，让用户选择下载源
                source_choice = self._show_download_source_dialog()
                if source_choice is None:
                    self.show_ffmpeg_warning()
                    return False
                    
                download_info = self.ffmpeg_download_sources[source_choice]
                download_url = download_info["url"]
                
                # 询问用户是否同意下载
                response = messagebox.askyesno(
                    "在线下载",
                    f"是否从 {download_info['name']} 下载 FFmpeg？\n"
                    f"文件大小：{download_info['size']}\n"
                    f"请确保网络连接正常。"
                )
                
                if not response:
                    self.show_ffmpeg_warning()
                    return False

                self.logger.info("开始下载FFmpeg...")
                
                # 创建临时目录
                temp_dir = tempfile.mkdtemp()
                zip_path = os.path.join(temp_dir, "ffmpeg.zip")
                
                try:
                    # 使用改进的下载方法
                    if self._download_with_retry(download_url, zip_path):
                        self.logger.info("下载完成，开始安装...")
                        
                        # 安装下载的文件
                        if self._install_ffmpeg_from_zip(zip_path):
                            return True
                        else:
                            messagebox.showerror("错误", "FFmpeg安装失败")
                            return False
                    else:
                        # 当前源下载失败，提示用户尝试其他源
                        retry_choice = messagebox.askyesno(
                            "下载失败",
                            f"从 {download_info['name']} 下载失败。\n\n"
                            f"是否尝试其他下载源？"
                        )
                        if retry_choice:
                            return self._try_alternative_sources(source_choice)
                        else:
                            self.show_ffmpeg_warning()
                            return False
                        
                finally:
                    # 清理临时文件
                    try:
                        shutil.rmtree(temp_dir)
                    except:
                        pass
                        
        except Exception as e:
            self.logger.error(f"下载FFmpeg失败: {str(e)}")
            if not silent:
                messagebox.showerror("错误", f"下载FFmpeg失败: {str(e)}")
                self.show_ffmpeg_warning()
            return False

    def _try_alternative_sources(self, failed_source_index):
        """尝试其他下载源"""
        for i, source in enumerate(self.ffmpeg_download_sources):
            if i == failed_source_index:  # 跳过已失败的源
                continue
                
            self.logger.info(f"尝试从 {source['name']} 下载...")
            
            # 创建临时目录
            temp_dir = tempfile.mkdtemp()
            zip_path = os.path.join(temp_dir, "ffmpeg.zip")
            
            try:
                if self._download_with_retry(source["url"], zip_path):
                    self.logger.info("下载完成，开始安装...")
                    
                    if self._install_ffmpeg_from_zip(zip_path):
                        return True
                    else:
                        self.logger.error(f"从 {source['name']} 下载成功但安装失败")
                        
            except Exception as e:
                self.logger.error(f"从 {source['name']} 下载失败: {str(e)}")
            finally:
                # 清理临时文件
                try:
                    shutil.rmtree(temp_dir)
                except:
                    pass
        
        # 所有源都失败
        self.logger.error("所有下载源都尝试失败了")
        messagebox.showerror(
            "下载失败",
            "所有下载源都尝试失败了。\n\n"
            "请检查网络连接或稍后再试。"
        )
        self.show_ffmpeg_warning()
        return False
        
    def _install_ffmpeg_from_zip(self, zip_path):
        """从ZIP文件安装FFmpeg"""
        try:
            # 创建安装目录
            install_dir = r"C:\Program Files\FFmpeg"
            
            # 检查是否有权限创建目录
            try:
                os.makedirs(install_dir, exist_ok=True)
            except PermissionError:
                # 如果没有权限，安装到用户目录
                install_dir = os.path.join(os.path.expanduser("~"), "FFmpeg")
                os.makedirs(install_dir, exist_ok=True)
                self.logger.info(f"使用用户目录安装: {install_dir}")
            
            # 解压文件
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                zip_ref.extractall(install_dir)
            
            # 查找ffmpeg.exe的实际路径
            ffmpeg_exe = None
            for root, dirs, files in os.walk(install_dir):
                if "ffmpeg.exe" in files:
                    ffmpeg_exe = os.path.join(root, "ffmpeg.exe")
                    break
            
            if not ffmpeg_exe:
                self.logger.error("解压后未找到ffmpeg.exe")
                return False
            
            # 获取FFmpeg的bin目录
            ffmpeg_bin_dir = os.path.dirname(ffmpeg_exe)
            
            # 添加到环境变量PATH
            if self._add_to_path(ffmpeg_bin_dir):
                self.logger.info(f"FFmpeg安装成功: {ffmpeg_exe}")
                self.logger.info(f"已添加到PATH: {ffmpeg_bin_dir}")
                return True
            else:
                self.logger.error("添加到PATH失败")
                return False
                
        except Exception as e:
            self.logger.error(f"安装FFmpeg失败: {str(e)}")
            return False

    def _add_to_path(self, path):
        """将路径添加到系统环境变量PATH"""
        try:
            import winreg
            
            # 尝试添加到系统PATH（需要管理员权限）
            try:
                with winreg.OpenKey(winreg.HKEY_LOCAL_MACHINE,
                                   r"SYSTEM\CurrentControlSet\Control\Session Manager\Environment",
                                   0, winreg.KEY_ALL_ACCESS) as key:
                    current_path, _ = winreg.QueryValueEx(key, "PATH")
                    if path not in current_path:
                        new_path = current_path + ";" + path
                        winreg.SetValueEx(key, "PATH", 0, winreg.REG_EXPAND_SZ, new_path)
                        self.logger.info("已添加到系统PATH")
                        return True
            except (PermissionError, OSError):
                # 如果没有管理员权限，添加到用户PATH
                with winreg.OpenKey(winreg.HKEY_CURRENT_USER,
                                   r"Environment",
                                   0, winreg.KEY_ALL_ACCESS) as key:
                    try:
                        current_path, _ = winreg.QueryValueEx(key, "PATH")
                    except FileNotFoundError:
                        current_path = ""
                    
                    if path not in current_path:
                        new_path = current_path + ";" + path if current_path else path
                        winreg.SetValueEx(key, "PATH", 0, winreg.REG_EXPAND_SZ, new_path)
                        self.logger.info("已添加到用户PATH")
                        return True
            
            return True
            
        except Exception as e:
            self.logger.error(f"添加到PATH失败: {str(e)}")
            return False

    def _get_friendly_error_message(self, error):
        """获取用户友好的错误信息"""
        error_str = str(error).lower()
        
        if "connection aborted" in error_str or "10053" in error_str:
            return (
                "网络连接被中断，可能的原因：\n\n"
                "1. 网络不稳定或速度过慢\n"
                "2. 防火墙或杀毒软件阻止了连接\n"
                "3. 代理设置问题\n\n"
                "建议解决方案：\n"
                "- 检查网络连接\n"
                "- 暂时关闭防火墙或杀毒软件\n"
                "- 尝试使用移动热点或其他网络"
            )
        elif "timeout" in error_str:
            return (
                "网络连接超时，可能的原因：\n\n"
                "1. 网络速度过慢\n"
                "2. 服务器响应过慢\n"
                "3. DNS 解析问题\n\n"
                "建议尝试其他下载源或稍后再试。"
            )
        elif "connection error" in error_str or "network" in error_str:
            return (
                "网络连接错误，请检查：\n\n"
                "1. 网络连接是否正常\n"
                "2. 防火墙设置\n"
                "3. 代理服务器配置\n\n"
                "建议尝试手动下载或使用其他网络。"
            )
        else:
            return f"下载过程中发生错误：\n{str(error)}\n\n请尝试其他下载源或稍后再试。"
    
    def show_ffmpeg_warning(self):
        """显示FFmpeg未安装警告"""
        # 创建更完善的提示对话框
        import tkinter as tk
        from tkinter import ttk
        
        dialog = tk.Toplevel()
        dialog.title("获取 FFmpeg")
        dialog.geometry("450x350")
        dialog.resizable(False, False)
        
        # 居中显示
        dialog.update_idletasks()
        x = (dialog.winfo_screenwidth() // 2) - (450 // 2)
        y = (dialog.winfo_screenheight() // 2) - (350 // 2)
        dialog.geometry(f"450x350+{x}+{y}")
        
        # 设置为模态对话框
        dialog.transient()
        dialog.grab_set()
        
        # 主框架
        main_frame = ttk.Frame(dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)
        
        # 标题
        title_label = ttk.Label(main_frame, text="获取 FFmpeg", font=("宋体", 14, "bold"))
        title_label.pack(pady=(0, 15))
        
        # 说明文字
        desc_text = (
            "FFmpeg 是推流功能必需的组件。\n\n"
            "您可以选择以下方式获取："
        )
        desc_label = ttk.Label(main_frame, text=desc_text, font=("宋体", 10), justify=tk.LEFT)
        desc_label.pack(pady=(0, 20))
        
        # 选项按钮
        def on_manual_download():
            dialog.destroy()
            webbrowser.open(self.ffmpeg_web_url)
            
        def on_retry_auto():
            dialog.destroy()
            # 重新尝试自动安装
            self.check_and_install()
            
        def on_cancel():
            dialog.destroy()
        
        # 按钮组
        btn_frame = ttk.Frame(main_frame)
        btn_frame.pack(fill=tk.X, pady=10)
        
        manual_btn = ttk.Button(
            btn_frame, 
            text="手动下载安装",
            command=on_manual_download,
            width=20
        )
        manual_btn.pack(pady=5, fill=tk.X)
        
        retry_btn = ttk.Button(
            btn_frame,
            text="重新尝试自动安装", 
            command=on_retry_auto,
            width=20
        )
        retry_btn.pack(pady=5, fill=tk.X)
        
        cancel_btn = ttk.Button(
            btn_frame,
            text="取消",
            command=on_cancel,
            width=20
        )
        cancel_btn.pack(pady=(15, 5), fill=tk.X)
        
        # 提示信息
        tip_text = (
            "提示：如果自动下载失败，可以手动下载\n"
            "FFmpeg 并解压到 C:\\Program Files\\FFmpeg\\"
        )
        tip_label = ttk.Label(main_frame, text=tip_text, font=("宋体", 8), foreground="gray")
        tip_label.pack(pady=(10, 0))
        
        # 等待用户选择
        dialog.wait_window()

    def uninstall_ffmpeg(self):
        """卸载FFmpeg"""
        try:
            # 确认是否卸载
            if not messagebox.askyesno(
                "确认", "确定要卸载FFmpeg吗？\n卸载后需要重新安装才能使用推流功能。"
            ):
                return

            # 查找FFmpeg安装路径
            install_paths = [
                r"C:\Program Files\FFmpeg",
                os.path.join(os.path.expanduser("~"), "FFmpeg"),
            ]

            removed_paths = []
            for path in install_paths:
                if os.path.exists(path):
                    try:
                        shutil.rmtree(path)
                        removed_paths.append(path)
                        self.logger.info(f"已删除FFmpeg目录: {path}")
                    except Exception as e:
                        self.logger.error(f"删除目录失败 {path}: {str(e)}")

            # 从环境变量中移除
            self._remove_from_path()

            if removed_paths:
                messagebox.showinfo("成功", f"FFmpeg已卸载\n删除的目录：\n" + "\n".join(removed_paths))
            else:
                messagebox.showinfo("提示", "未找到FFmpeg安装目录")

        except Exception as e:
            error_msg = f"卸载FFmpeg时出错: {str(e)}"
            self.logger.error(error_msg)
            messagebox.showerror("错误", error_msg)

    def _remove_from_path(self):
        """从环境变量PATH中移除FFmpeg路径"""
        try:
            import winreg
            
            # 从系统PATH移除
            try:
                with winreg.OpenKey(winreg.HKEY_LOCAL_MACHINE,
                                   r"SYSTEM\CurrentControlSet\Control\Session Manager\Environment",
                                   0, winreg.KEY_ALL_ACCESS) as key:
                    current_path, _ = winreg.QueryValueEx(key, "PATH")
                    path_list = [p for p in current_path.split(";") if "ffmpeg" not in p.lower()]
                    new_path = ";".join(path_list)
                    winreg.SetValueEx(key, "PATH", 0, winreg.REG_EXPAND_SZ, new_path)
                    self.logger.info("已从系统PATH移除FFmpeg")
            except (PermissionError, OSError):
                pass
            
            # 从用户PATH移除
            try:
                with winreg.OpenKey(winreg.HKEY_CURRENT_USER,
                                   r"Environment",
                                   0, winreg.KEY_ALL_ACCESS) as key:
                    current_path, _ = winreg.QueryValueEx(key, "PATH")
                    path_list = [p for p in current_path.split(";") if "ffmpeg" not in p.lower()]
                    new_path = ";".join(path_list)
                    winreg.SetValueEx(key, "PATH", 0, winreg.REG_EXPAND_SZ, new_path)
                    self.logger.info("已从用户PATH移除FFmpeg")
            except (FileNotFoundError, OSError):
                pass
                
        except Exception as e:
            self.logger.error(f"从PATH移除失败: {str(e)}")

    def install_ffmpeg(self, silent=False):
        """手动安装 FFmpeg"""
        try:
            if self.check():
                return True
                
            self.logger.info("开始安装 FFmpeg...")
            
            # 优先尝试本地文件
            if self._install_from_local_files():
                return True
            
            # 本地文件安装失败，尝试在线下载
            return self.check_and_install(silent=silent)
        except Exception as e:
            self.logger.error(f"安装 FFmpeg 失败: {str(e)}")
            if not silent:
                messagebox.showerror("错误", f"启动 FFmpeg 安装程序失败: {str(e)}")
            return False

    def get_ffmpeg_info(self):
        """获取FFmpeg版本信息"""
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