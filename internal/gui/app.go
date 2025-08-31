package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"douyin-rtmp/internal/capture"
	"douyin-rtmp/internal/config"
	"douyin-rtmp/internal/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// App GUI应用程序主结构
type App struct {
	fyneApp        fyne.App
	window         fyne.Window
	logger         *logger.Logger
	configManager  *config.ConfigManager
	captureManager capture.CaptureManager

	// UI组件
	controlPanel   *ControlPanel
	logPanel       *LogPanel
	obsPanel       *OBSPanel
	adPanel        *AdPanel

	// 状态变量
	serverAddress  string
	streamCode     string

	// 应用状态
	isCapturing      bool
	captureStartTime time.Time

	// 配置变量
	checkUpdate      bool
}

// NewApp 创建新的GUI应用程序
func NewApp() *App {
	fyneApp := app.NewWithID("com.douyin-rtmp.app")

	// 创建日志管理器
	log := logger.NewLogger()

	// 创建配置管理器
	configMgr := config.GetDefaultConfigManager()

	// 创建捕获管理器
	captureMgr := capture.NewCaptureManager(log)

	// 创建主窗口
	window := fyneApp.NewWindow("抖音直播推流地址获取工具 v1.0.0")

	app := &App{
		fyneApp:        fyneApp,
		window:         window,
		logger:         log,
		configManager:  configMgr,
		captureManager: captureMgr,
		isCapturing:    false,
		checkUpdate:    true, // 默认启用自动检查更新
	}

	// 初始化UI组件
	app.controlPanel = NewControlPanel(app)
	app.logPanel = NewLogPanel(app)
	app.obsPanel = NewOBSPanel(app)
	app.adPanel = NewAdPanel(app)

	// 设置窗口
	app.setupWindow()

	// 设置日志回调
	app.setupLogCallbacks()

	// 设置捕获回调
	app.setupCaptureCallbacks()

	return app
}

// setupWindow 设置窗口
func (a *App) setupWindow() {
	// 从配置加载窗口设置
	width, height, x, y := a.configManager.GetWindowConfig()

	if width > 0 && height > 0 {
		a.window.Resize(fyne.NewSize(width, height))
	} else {
		a.window.Resize(fyne.NewSize(800, 600))
	}

	// 设置窗口位置（如果配置了的话）
	if x >= 0 && y >= 0 {
		a.window.SetFixedSize(false) // 允许调整大小
	} else {
		a.window.CenterOnScreen()
	}

	// 设置窗口图标
	a.window.SetIcon(theme.ComputerIcon())

	// 设置主要内容
	content := a.createMainContent()
	a.window.SetContent(content)

	// 设置窗口关闭回调
	a.window.SetCloseIntercept(a.onWindowClosing)
}

// createMainContent 创建主要内容
func (a *App) createMainContent() fyne.CanvasObject {
	// 设置菜单栏（与Python版本保持一致）
	mainMenu := a.createMenuBar()
	a.window.SetMainMenu(mainMenu)

	// 创建主要区域
	mainArea := a.createMainArea()

	// 创建状态栏
	statusBar := a.createStatusBar()

	// 组合布局
	return container.NewBorder(
		nil,       // top
		statusBar, // bottom
		nil,       // left
		nil,       // right
		mainArea,  // center
	)
}

// createMainArea 创建主要区域
func (a *App) createMainArea() fyne.CanvasObject {
	// 创建控制面板内容
	controlPanelContent := a.controlPanel.CreateContent()

	// 创建日志区域内容
	logContent := a.logPanel.CreateContent()

	// 创建OBS管理面板内容
	obsContent := a.obsPanel.CreateContent()

	// 创建广告面板内容
	adContent := a.adPanel.CreateContent()

	// 组合控制面板和OBS面板在右侧
	rightPanel := container.NewVBox(
		controlPanelContent,
		obsContent,
		adContent,
	)

	// 主内容区域，左侧是日志，右侧是控制面板
	mainContent := container.NewHSplit(
		logContent,
		rightPanel,
	)
	mainContent.Offset = 0.6 // 左侧日志占60%

	return mainContent
}

// setupLogCallbacks 设置日志回调
func (a *App) setupLogCallbacks() {
	a.logger.AddHandler(func(entry *logger.LogEntry) {
		// 更新日志显示
		a.logPanel.UpdateLogs(entry)
	})
}

// setupCaptureCallbacks 设置捕获回调
func (a *App) setupCaptureCallbacks() {
	a.captureManager.AddResultCallback(func(result *capture.CaptureResult) {
		if result.Success && result.RTMPInfo != nil {
			// 更新结果显示
			a.serverAddress = result.RTMPInfo.ServerAddress
			a.streamCode = result.RTMPInfo.StreamCode
			a.controlPanel.UpdateCaptureResult(result.RTMPInfo)

			// 如果获取到完整信息，显示通知
			if result.RTMPInfo.FullURL != "" {
				a.fyneApp.SendNotification(&fyne.Notification{
					Title:   "推流地址获取成功！",
					Content: "已成功获取完整的推流地址",
				})
			}
		}
	})
}

// Run 运行应用程序
func (a *App) Run() {
	a.logger.Info("启动抖音RTMP推流地址获取工具")

	// 检查管理员权限
	a.checkAdminPermission()

	// 延迟初始化
	go func() {
		time.Sleep(100 * time.Millisecond)
		a.delayedInit()
	}()

	// 运行应用
	a.window.ShowAndRun()
}

// delayedInit 延迟初始化
func (a *App) delayedInit() {
	// 初始化网络接口
	a.controlPanel.LoadInterfaces()

	// 检查更新（如果启用）
	if a.checkUpdate {
		time.Sleep(1000 * time.Millisecond)
		a.asyncCheckUpdates()
	}

	// 获取广告内容
	a.adPanel.AsyncFetchAdContent()
}

// checkAdminPermission 检查管理员权限
func (a *App) checkAdminPermission() {
	// Windows权限检查的简化实现
	// 实际项目中可能需要更复杂的检查
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		a.logger.Warn("建议以管理员权限运行程序以确保网络捕获功能正常工作")
	}
}

// asyncCheckUpdates 异步检查更新
func (a *App) asyncCheckUpdates() {
	go func() {
		// 这里实现版本检查逻辑
		a.logger.Info("正在检查更新...")
		// TODO: 实现实际的版本检查逻辑
	}()
}

// createMenuBar 创建菜单栏
func (a *App) createMenuBar() *fyne.MainMenu {
	// 工具菜单
	toolsMenu := fyne.NewMenu("工具",
		fyne.NewMenuItem("安装 Npcap", func() {
			a.installNpcap()
		}),
		fyne.NewMenuItem("卸载 Npcap", func() {
			a.uninstallNpcap()
		}),
	)

	// 帮助菜单
	helpMenu := fyne.NewMenu("帮助",
		fyne.NewMenuItem("帮助中心", func() {
			a.openHelperCenter()
		}),
		fyne.NewMenuItem("检查软件更新", func() {
			a.checkUpdatesManually()
		}),
		fyne.NewMenuItem("GitHub 仓库", func() {
			a.openGitHubRepo()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("关于 (v1.0.0)", func() {
			a.ShowAbout()
		}),
	)

	// 贡献榜菜单项
	contributeMenuItem := fyne.NewMenuItem("贡献榜", func() {
		a.showContribute()
	})

	return fyne.NewMainMenu(toolsMenu, helpMenu, contributeMenuItem)
}

// createStatusBar 创建状态栏
func (a *App) createStatusBar() fyne.CanvasObject {
	// 左侧版本信息
	versionLabel := widget.NewLabel("版本: v1.0.0")

	// 自动检查更新复选框
	checkUpdateCheck := widget.NewCheck("启动时检查更新", func(checked bool) {
		a.checkUpdate = checked
		a.configManager.SetCheckUpdate(checked)
	})
	checkUpdateCheck.SetChecked(a.checkUpdate)

	// 右侧按钮组
	donationBtn := widget.NewButton("请作者喝杯咖啡", func() {
		a.showDonation()
	})

	disclaimerBtn := widget.NewButton("免责声明", func() {
		a.showDisclaimer()
	})

	helpBtn := widget.NewButton("使用说明", func() {
		a.showHelp()
	})

	// 组合状态栏布局
	leftContainer := container.NewHBox(versionLabel)
	rightContainer := container.NewHBox(checkUpdateCheck, donationBtn, disclaimerBtn, helpBtn)

	return container.NewBorder(nil, nil, leftContainer, rightContainer)
}



// 各种事件处理方法

// installNpcap 安装Npcap
func (a *App) installNpcap() {
	a.logger.Info("启动Npcap安装程序...")
	// TODO: 实现Npcap安装逻辑
}

// uninstallNpcap 卸载Npcap
func (a *App) uninstallNpcap() {
	a.logger.Info("启动Npcap卸载程序...")
	// TODO: 实现Npcap卸载逻辑
}

// openHelperCenter 打开帮助中心
func (a *App) openHelperCenter() {
	a.logger.Info("打开帮助中心...")
	// TODO: 实现打开帮助中心逻辑
}

// checkUpdatesManually 手动检查更新
func (a *App) checkUpdatesManually() {
	a.logger.Info("正在检查更新...")
	// TODO: 实现手动检查更新逻辑
}

// openGitHubRepo 打开GitHub仓库
func (a *App) openGitHubRepo() {
	a.logger.Info("打开GitHub仓库...")
	// TODO: 实现打开GitHub仓库逻辑
}

// showContribute 显示贡献榜
func (a *App) showContribute() {
	a.logger.Info("显示贡献榜...")
	// TODO: 实现贡献榜对话框
}

// showDonation 显示打赏对话框
func (a *App) showDonation() {
	content := widget.NewRichTextFromMarkdown(`
# 请作者喝杯咖啡

感谢您的支持！另请注意：此捐赠为纯自愿，非强制性，请根据自身情况自愿捐赠，谢谢！

无论多少都是心意，一分也是对我莫大的鼓励！
学生党或者直播没收益就不用啦！当然，大佬请随意~
预祝各位老师们大红大紫！
`)

	dialog.ShowCustom("感谢支持", "关闭", content, a.window)
}

// showDisclaimer 显示免责声明
func (a *App) showDisclaimer() {
	content := widget.NewRichTextFromMarkdown(`
# 免责声明

1. 本软件仅用于个人学习和测试使用，无需提供任何代价，并不可用于任何商业用途及目的（包括二次开发）；

2. 工具仅抓取公开传输的明文数据，未采用破解加密、伪装身份等技术手段；

3. 使用本软件时请遵守相关法律法规，不得用于任何违法用途。

4. 本工具推荐使用者使用官方直播工具，使用者需遵守抖音平台规定，若抖音平台禁止使用第三方工具，请立即停止工具的使用并删除工具；

5. 本软件开源免费，作者不对使用本软件造成的任何直接或间接损失负责。

6. 使用本软件即表示您同意本免责声明的所有条款。

7. 若本工具有造成任何可能的侵权行为，请进行联系，将会立即停止可能的侵权行为，并关闭代码仓库；

8. 本工具将根据抖音平台协议进行技术调整，确保合法合规；

9. 作者保留对本软件和免责声明的最终解释权。
`)

	dialog.ShowCustom("免责声明", "确定", content, a.window)
}

// showHelp 显示使用说明
func (a *App) showHelp() {
	content := widget.NewRichTextFromMarkdown(`
# 使用说明

## 使用步骤

1. **选择网络接口**: 在控制面板中选择要监听的网络接口
2. **开始捕获**: 点击"开始捕获"按钮
3. **开启直播**: 打开抖音直播伴侣开始直播
4. **获取地址**: 工具会自动捕获并显示推流地址

## 注意事项

- 请确保以管理员权限运行程序
- 选择正确的网络接口（通常是主网卡）
- 在开始捕获后再启动直播软件
- 如果无法捕获，请尝试勾选"日志模式（备选）"

## 常见问题

**Q: 为什么捕获不到推流地址？**
A: 请检查是否以管理员权限运行，并选择正确的网络接口。

**Q: 如何选择正确的网络接口？**
A: 通常选择显示为"以太网"或"本地连接"的接口。
`)

	dialog.ShowCustom("使用说明", "确定", content, a.window)
}

// ShowAbout 显示关于对话框
func (a *App) ShowAbout() {
	content := widget.NewRichTextFromMarkdown(`
# 抖音直播推流地址获取工具

**版本**: v1.0.0  
**作者**: 关水来了  
**技术栈**: Go + Fyne + GoPacket

## 功能特性

- 🔍 高性能网络数据包捕获
- 🎯 智能RTMP地址解析  
- 🖥️ 跨平台图形界面
- 📊 实时日志监控
- ⚙️ 灵活配置管理

## 开源许可

本项目采用 MIT 许可证开源

本工具仅供学习交流使用
`)

	dialog.ShowCustom("关于", "确定", content, a.window)
}

// StartCapture 开始捕获
func (a *App) StartCapture() error {
	if a.isCapturing {
		return fmt.Errorf("已在捕获中")
	}

	// 获取选择的接口
	selectedInterfaces := a.controlPanel.GetSelectedInterfaces()
	if len(selectedInterfaces) == 0 {
		dialog.ShowError(fmt.Errorf("请先选择网络接口"), a.window)
		return fmt.Errorf("未选择接口")
	}

	// 开始捕获
	err := a.captureManager.StartCapture(selectedInterfaces)
	if err != nil {
		dialog.ShowError(err, a.window)
		return err
	}

	a.isCapturing = true
	a.captureStartTime = time.Now()

	// 更新UI状态
	a.controlPanel.SetCaptureStatus(true)

	a.logger.Info(fmt.Sprintf("开始捕获，监听 %d 个接口", len(selectedInterfaces)))
	return nil
}

// StopCapture 停止捕获
func (a *App) StopCapture() error {
	if !a.isCapturing {
		return fmt.Errorf("未在捕获中")
	}

	err := a.captureManager.StopCapture()
	if err != nil {
		return err
	}

	a.isCapturing = false

	// 更新UI状态
	a.controlPanel.SetCaptureStatus(false)

	duration := time.Since(a.captureStartTime)
	a.logger.Info(fmt.Sprintf("停止捕获，持续时间: %v", duration))

	return nil
}

// RefreshInterfaces 刷新接口列表
func (a *App) RefreshInterfaces() {
	err := a.captureManager.RefreshInterfaces()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// 更新界面
	a.controlPanel.RefreshInterfaceList()
	a.logger.Info("已刷新网络接口列表")
}

// TestInterface 测试接口
func (a *App) TestInterface(interfaceName string) {
	a.logger.Info(fmt.Sprintf("正在测试接口: %s", interfaceName))

	// 显示进度对话框
	progress := dialog.NewProgressInfinite("测试中", "正在测试网络接口...", a.window)
	progress.Show()

	go func() {
		defer progress.Hide()

		stats, err := a.captureManager.TestInterface(interfaceName, 5*time.Second)
		if err != nil {
			a.fyneApp.SendNotification(&fyne.Notification{
				Title:   "测试失败",
				Content: err.Error(),
			})
			return
		}

		message := fmt.Sprintf("接口测试完成\n捕获数据包: %d\n数据量: %d 字节",
			stats.PacketCount, stats.ByteCount)

		if stats.Success {
			dialog.ShowInformation("测试成功", message, a.window)
		} else {
			dialog.ShowError(fmt.Errorf("测试失败: %s", message), a.window)
		}
	}()
}
func (a *App) onWindowClosing() {
	// 如果正在捕获，先停止
	if a.isCapturing {
		a.StopCapture()
	}

	// 保存窗口配置
	size := a.window.Canvas().Size()
	a.configManager.SetWindowSize(size.Width, size.Height)

	// 保存配置
	a.configManager.Save()

	// 关闭窗口
	a.window.Close()
}

// GetLogger 获取日志管理器
func (a *App) GetLogger() *logger.Logger {
	return a.logger
}

// GetConfigManager 获取配置管理器
func (a *App) GetConfigManager() *config.ConfigManager {
	return a.configManager
}

// GetCaptureManager 获取捕获管理器
func (a *App) GetCaptureManager() capture.CaptureManager {
	return a.captureManager
}

// GetWindow 获取窗口
func (a *App) GetWindow() fyne.Window {
	return a.window
}

// IsCapturing 检查是否正在捕获
func (a *App) IsCapturing() bool {
	return a.isCapturing
}

// GetCaptureStats 获取捕获统计信息
func (a *App) GetCaptureStats() *capture.CaptureManagerStats {
	return a.captureManager.GetStats()
}

// ExportResults 导出结果
func (a *App) ExportResults() {
	stats := a.GetCaptureStats()
	if !stats.RTMPFound {
		dialog.ShowInformation("提示", "尚未获取到RTMP推流信息", a.window)
		return
	}

	content := fmt.Sprintf(`# 抖音推流地址获取结果

## 推流信息
- **服务器地址**: %s
- **推流码**: %s  
- **完整地址**: %s/%s

## 统计信息
- **捕获时间**: %v
- **数据包数量**: %d
- **活跃接口**: %v

---
导出时间: %s
`,
		stats.ServerAddress,
		stats.StreamCode,
		stats.ServerAddress,
		stats.StreamCode,
		stats.Duration,
		stats.TotalPackets,
		stats.ActiveInterfaces,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// 显示导出内容
	textArea := widget.NewMultiLineEntry()
	textArea.SetText(content)
	textArea.Resize(fyne.NewSize(600, 400))

	dialog.ShowCustom("导出结果", "复制到剪贴板", container.NewScroll(textArea), a.window)
}

// ShowNetworkDiagnostic 显示网络诊断
func (a *App) ShowNetworkDiagnostic() {
	diagnostic, err := a.captureManager.GetNetworkDiagnostic()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	content := widget.NewRichTextFromMarkdown(fmt.Sprintf(`
# 网络诊断信息

## 系统信息
- **主机名**: %v
- **操作系统**: %v  
- **平台**: %v

## 接口统计
- **总接口数**: %v
- **活跃接口数**: %v

详细信息请查看日志面板。
`,
		diagnostic["hostname"],
		diagnostic["os"],
		diagnostic["platform"],
		diagnostic["interface_stats"],
		diagnostic["interface_stats"],
	))

	dialog.ShowCustom("网络诊断", "确定", content, a.window)
}

// PerformQuickTest 执行快速测试
func (a *App) PerformQuickTest() {
	a.logger.Info("开始执行快速网络测试...")

	progress := dialog.NewProgressInfinite("测试中", "正在执行网络连通性测试...", a.window)
	progress.Show()

	go func() {
		defer progress.Hide()

		results := a.captureManager.PerformConnectivityTest()

		message := "网络连通性测试结果:\n\n"
		for iface, connected := range results {
			status := "❌ 未连接"
			if connected {
				status = "✅ 已连接"
			}
			message += fmt.Sprintf("%s: %s\n", iface, status)
		}

		dialog.ShowInformation("测试结果", message, a.window)
	}()
}
