package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"douyin-rtmp/internal/gui"
)

const (
	AppName    = "抖音直播推流地址获取工具"
	AppVersion = "v1.0.0"
	AppAuthor  = "Go版本"
)

var (
	// 命令行参数
	versionFlag = flag.Bool("version", false, "显示版本信息")
	helpFlag    = flag.Bool("help", false, "显示帮助信息")
	debugFlag   = flag.Bool("debug", false, "启用调试模式")
	consoleFlag = flag.Bool("console", false, "启用控制台模式（无GUI）")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 处理版本信息
	if *versionFlag {
		printVersion()
		return
	}

	// 处理帮助信息
	if *helpFlag {
		printHelp()
		return
	}

	// 检查运行环境
	if err := checkEnvironment(); err != nil {
		log.Fatalf("环境检查失败: %v", err)
	}

	// 检查权限（仅在需要时）
	if runtime.GOOS == "windows" {
		if !isRunningAsAdmin() {
			fmt.Println("警告: 建议以管理员权限运行程序以确保网络捕获功能正常工作")
		}
	}

	// 根据参数选择运行模式
	if *consoleFlag {
		runConsoleMode()
	} else {
		runGUIMode()
	}
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("%s %s\n", AppName, AppVersion)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("作者: %s\n", AppAuthor)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Printf("%s %s - 使用说明\n\n", AppName, AppVersion)

	fmt.Println("用法:")
	fmt.Printf("  %s [选项]\n\n", os.Args[0])

	fmt.Println("选项:")
	fmt.Println("  -version     显示版本信息")
	fmt.Println("  -help        显示此帮助信息")
	fmt.Println("  -debug       启用调试模式")
	fmt.Println("  -console     启用控制台模式（无GUI）")

	fmt.Println("\n描述:")
	fmt.Println("  本工具用于捕获抖音直播推流地址，支持图形界面和命令行模式。")
	fmt.Println("  默认启动图形界面模式，使用 -console 参数可切换到命令行模式。")

	fmt.Println("\n示例:")
	fmt.Printf("  %s                    # 启动图形界面\n", os.Args[0])
	fmt.Printf("  %s -console           # 启动命令行模式\n", os.Args[0])
	fmt.Printf("  %s -debug             # 启用调试模式\n", os.Args[0])
	fmt.Printf("  %s -version           # 显示版本信息\n", os.Args[0])

	fmt.Println("\n注意事项:")
	fmt.Println("  - Windows系统建议以管理员权限运行")
	fmt.Println("  - Linux/macOS系统需要使用 sudo 运行")
	fmt.Println("  - 确保已安装相应的网络捕获库（Windows需要Npcap，Linux需要libpcap）")
}

// checkEnvironment 检查运行环境
func checkEnvironment() error {
	// 检查操作系统支持
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		// 支持的操作系统
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	// 检查Go版本（可选）
	// 这里可以添加更多环境检查逻辑

	return nil
}

// isRunningAsAdmin 检查是否以管理员权限运行（Windows）
func isRunningAsAdmin() bool {
	if runtime.GOOS != "windows" {
		return true // 非Windows系统假定有权限
	}

	// Windows权限检查的简化实现
	// 实际项目中可能需要更复杂的检查
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

// runGUIMode 运行图形界面模式
func runGUIMode() {
	fmt.Printf("启动 %s %s (图形界面模式)\n", AppName, AppVersion)

	// 创建并运行GUI应用
	app := gui.NewApp()

	// 如果启用调试模式，设置相应的日志级别
	if *debugFlag {
		fmt.Println("调试模式已启用")
		// 可以在这里设置更详细的日志输出
	}

	// 运行应用
	app.Run()
}

// runConsoleMode 运行控制台模式
func runConsoleMode() {
	fmt.Printf("启动 %s %s (控制台模式)\n", AppName, AppVersion)
	fmt.Println("控制台模式功能开发中...")

	// 这里可以实现控制台版本的功能
	// 包括命令行交互、配置文件操作等

	consoleApp := NewConsoleApp()
	consoleApp.Run()
}

// ConsoleApp 控制台应用结构
type ConsoleApp struct {
	// 控制台应用的字段
}

// NewConsoleApp 创建控制台应用
func NewConsoleApp() *ConsoleApp {
	return &ConsoleApp{}
}

// Run 运行控制台应用
func (ca *ConsoleApp) Run() {
	fmt.Println("=== 抖音直播推流地址获取工具 (控制台版) ===")
	fmt.Println()

	// 显示菜单
	ca.showMenu()

	// 主循环
	for {
		fmt.Print("\n请选择操作 (输入数字): ")

		var choice int
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			fmt.Println("输入无效，请输入数字")
			continue
		}

		switch choice {
		case 1:
			ca.listInterfaces()
		case 2:
			ca.startCapture()
		case 3:
			ca.showStatus()
		case 4:
			ca.showHelp()
		case 0:
			fmt.Println("退出程序...")
			return
		default:
			fmt.Println("无效选择，请重新输入")
		}
	}
}

// showMenu 显示菜单
func (ca *ConsoleApp) showMenu() {
	fmt.Println("可用操作:")
	fmt.Println("  1. 列出网络接口")
	fmt.Println("  2. 开始捕获")
	fmt.Println("  3. 显示状态")
	fmt.Println("  4. 显示帮助")
	fmt.Println("  0. 退出程序")
}

// listInterfaces 列出网络接口
func (ca *ConsoleApp) listInterfaces() {
	fmt.Println("\n=== 网络接口列表 ===")
	fmt.Println("功能开发中...")
	// 这里应该调用网络接口管理器来获取接口列表
}

// startCapture 开始捕获
func (ca *ConsoleApp) startCapture() {
	fmt.Println("\n=== 开始捕获 ===")
	fmt.Println("功能开发中...")
	// 这里应该实现捕获逻辑
}

// showStatus 显示状态
func (ca *ConsoleApp) showStatus() {
	fmt.Println("\n=== 当前状态 ===")
	fmt.Printf("应用版本: %s\n", AppVersion)
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPU数量: %d\n", runtime.NumCPU())
	// 可以添加更多状态信息
}

// showHelp 显示帮助
func (ca *ConsoleApp) showHelp() {
	fmt.Println("\n=== 使用帮助 ===")
	fmt.Println("1. 首先列出可用的网络接口")
	fmt.Println("2. 选择合适的网络接口开始捕获")
	fmt.Println("3. 打开抖音直播伴侣开始直播")
	fmt.Println("4. 工具会自动捕获并显示推流地址")
	fmt.Println("\n注意: 请确保以管理员权限运行程序")
}

// init 初始化函数
func init() {
	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 根据操作系统设置不同的默认值
	switch runtime.GOOS {
	case "windows":
		// Windows特定初始化
	case "linux":
		// Linux特定初始化
	case "darwin":
		// macOS特定初始化
	}
}
