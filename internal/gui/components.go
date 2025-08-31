package gui

import (
	"fmt"
	"strings"
	"time"

	"douyin-rtmp/internal/config"
	"douyin-rtmp/internal/rtmp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	_ "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Components UI组件管理器 - 重构版本，与Python版本保持一致
type Components struct {
	app *App

	// 控制面板组件
	interfaceSelect *widget.Select
	captureButton   *widget.Button
	testButton      *widget.Button
	refreshButton   *widget.Button

	// 增加的控制组件（与Python版本保持一致）
	listenAllCheck *widget.Check // 监听所有接口
	logModeCheck   *widget.Check // 日志模式（备选）

	// 结果显示组件
	serverAddressEntry *widget.Entry
	streamCodeEntry    *widget.Entry
	fullURLEntry       *widget.Entry

	// 新增按钮
	copyServerButton *widget.Button
	copyStreamButton *widget.Button
	copyFullButton   *widget.Button

	// 日志组件
	consoleLog *widget.RichText
	packetLog  *widget.RichText
	logTabs    *container.AppTabs

	// 状态组件
	statusLabel *widget.Label
	progressBar *widget.ProgressBarInfinite
	statsLabel  *widget.Label
}

// NewComponents 创建UI组件管理器
func NewComponents(app *App) *Components {
	return &Components{
		app: app,
	}
}

// CreateMenuBar 创建菜单栏（与Python版本保持一致）
func (c *Components) CreateMenuBar() *fyne.MainMenu {
	// 工具菜单
	installNpcapItem := fyne.NewMenuItem("安装 Npcap", func() {
		// TODO: 实现安装Npcap
		c.app.logger.Info("安装 Npcap")
	})
	uninstallNpcapItem := fyne.NewMenuItem("卸载 Npcap", func() {
		// TODO: 实现卸载Npcap
		c.app.logger.Info("卸载 Npcap")
	})
	toolsMenu := fyne.NewMenu("工具", installNpcapItem, uninstallNpcapItem)

	// 帮助菜单
	helpCenterItem := fyne.NewMenuItem("帮助中心", func() {
		// TODO: 实现帮助中心
		c.app.logger.Info("打开帮助中心")
	})
	checkUpdatesItem := fyne.NewMenuItem("检查软件更新", func() {
		// TODO: 实现检查更新
		c.app.logger.Info("检查软件更新")
	})
	githubItem := fyne.NewMenuItem("GitHub 仓库", func() {
		// TODO: 打开GitHub
		c.app.logger.Info("打开GitHub仓库")
	})
	aboutItem := fyne.NewMenuItem("关于", func() {
		c.app.ShowAbout()
	})
	helpMenu := fyne.NewMenu("帮助", helpCenterItem, checkUpdatesItem, fyne.NewMenuItemSeparator(), githubItem, aboutItem)

	// 贡献榜菜单
	contributeMenu := fyne.NewMenu("贡献榜", fyne.NewMenuItem("查看贡献榜", func() {
		// TODO: 实现贡献榜
		c.app.logger.Info("查看贡献榜")
	}))

	return fyne.NewMainMenu(toolsMenu, helpMenu, contributeMenu)
}

// CreateControlPanel 创建控制面板（重构为与Python版本一致的布局）
func (c *Components) CreateControlPanel() fyne.CanvasObject {
	// 控制面板标题
	title := widget.NewRichTextFromMarkdown("## 控制面板")

	// 第一行：网络接口选择
	interfaceLabel := widget.NewLabel("网络接口:")
	c.interfaceSelect = widget.NewSelect([]string{}, func(selected string) {
		if selected != "" {
			c.app.logger.Info(fmt.Sprintf("选择接口: %s", selected))
		}
	})
	c.interfaceSelect.PlaceHolder = "选择网络接口..."

	c.refreshButton = widget.NewButton("刷新", func() {
		c.app.RefreshInterfaces()
	})

	interfaceRow := container.NewHBox(
		interfaceLabel,
		c.interfaceSelect,
		c.refreshButton,
	)

	// 第二行：状态和控制按钮
	c.captureButton = widget.NewButton("开始捕获", func() {
		if c.app.IsCapturing() {
			c.app.StopCapture()
		} else {
			c.app.StartCapture()
		}
	})
	c.captureButton.Importance = widget.HighImportance

	c.testButton = widget.NewButton("检测可用", func() {
		selected := c.interfaceSelect.Selected
		if selected == "" {
			return
		}
		c.app.TestInterface(selected)
	})

	statusTextLabel := widget.NewLabel("状态:")
	c.statusLabel = widget.NewLabel("待开始")

	// 监听所有接口复选框
	c.listenAllCheck = widget.NewCheck("监听所有接口", func(checked bool) {
		// TODO: 实现监听所有接口逻辑
		c.app.logger.Info(fmt.Sprintf("监听所有接口: %v", checked))
	})
	c.listenAllCheck.SetChecked(true) // 默认选中

	// 日志模式复选框
	c.logModeCheck = widget.NewCheck("日志模式（备选）", func(checked bool) {
		// TODO: 实现日志模式逻辑
		c.app.logger.Info(fmt.Sprintf("日志模式: %v", checked))
	})

	controlRow := container.NewHBox(
		c.captureButton,
		c.testButton,
		statusTextLabel,
		c.statusLabel,
		c.listenAllCheck,
		c.logModeCheck,
	)

	// 第三行：推流服务器
	serverLabel := widget.NewLabel("推流服务器:")
	c.serverAddressEntry = widget.NewEntry()
	c.serverAddressEntry.SetPlaceHolder("等待获取...")
	c.serverAddressEntry.Disable() // 只读

	c.copyServerButton = widget.NewButton("复制", func() {
		if c.serverAddressEntry.Text != "" {
			c.app.window.Clipboard().SetContent(c.serverAddressEntry.Text)
			c.app.logger.Info("已复制服务器地址到剪贴板")
		}
	})

	serverRow := container.NewHBox(
		serverLabel,
		c.serverAddressEntry,
		c.copyServerButton,
	)

	// 第四行：推流码
	streamLabel := widget.NewLabel("推流码:")
	c.streamCodeEntry = widget.NewEntry()
	c.streamCodeEntry.SetPlaceHolder("等待获取...")
	c.streamCodeEntry.Disable() // 只读

	c.copyStreamButton = widget.NewButton("复制", func() {
		if c.streamCodeEntry.Text != "" {
			c.app.window.Clipboard().SetContent(c.streamCodeEntry.Text)
			c.app.logger.Info("已复制推流码到剪贴板")
		}
	})

	streamRow := container.NewHBox(
		streamLabel,
		c.streamCodeEntry,
		c.copyStreamButton,
	)

	// 组合所有行
	controlPanel := container.NewVBox(
		title,
		widget.NewSeparator(),
		interfaceRow,
		controlRow,
		serverRow,
		streamRow,
	)

	// 初始化接口列表
	c.RefreshInterfaceList()

	return controlPanel
}

// CreateLogArea 创建日志区域（与Python版本一致的标签页结构）
func (c *Components) CreateLogArea() fyne.CanvasObject {
	// 创建控制台输出标签页
	c.consoleLog = widget.NewRichText()
	c.consoleLog.Wrapping = fyne.TextWrapWord

	clearConsoleButton := widget.NewButton("清除控制台", func() {
		c.consoleLog.ParseMarkdown("")
		c.app.logger.Info("控制台已清除")
	})

	consoleContent := container.NewBorder(
		nil,                               // top
		clearConsoleButton,                // bottom
		nil,                               // left
		nil,                               // right
		container.NewScroll(c.consoleLog), // center
	)

	// 创建数据包监控标签页
	c.packetLog = widget.NewRichText()
	c.packetLog.Wrapping = fyne.TextWrapWord

	clearPacketButton := widget.NewButton("清除数据包日志", func() {
		c.packetLog.ParseMarkdown("")
		c.app.logger.Info("数据包日志已清除")
	})

	packetContent := container.NewBorder(
		nil,                              // top
		clearPacketButton,                // bottom
		nil,                              // left
		nil,                              // right
		container.NewScroll(c.packetLog), // center
	)

	// 创建标签页容器
	c.logTabs = container.NewAppTabs(
		container.NewTabItem("控制台输出", consoleContent),
		container.NewTabItem("数据包监控", packetContent),
	)

	return c.logTabs
}

// CreateStatusBar 创建状态栏（与Python版本一致）
func (c *Components) CreateStatusBar() fyne.CanvasObject {
	// 左侧版本信息
	versionLabel := widget.NewLabel(fmt.Sprintf("版本: %s", config.VERSION))

	// 右侧按钮组
	// 自动检查更新复选框
	autoUpdateCheck := widget.NewCheck("启动时检查更新", func(checked bool) {
		// TODO: 实现自动更新配置
		c.app.logger.Info(fmt.Sprintf("自动检查更新: %v", checked))
	})
	autoUpdateCheck.SetChecked(true) // 默认选中

	// 打赏按钮
	donationButton := widget.NewButton("请作者喝杯咖啡", func() {
		// TODO: 实现打赏功能
		c.app.logger.Info("打开打赏页面")
	})

	// 免责声明按钮
	disclaimerButton := widget.NewButton("免责声明", func() {
		// TODO: 实现免责声明
		c.app.logger.Info("显示免责声明")
	})

	// 使用说明按钮
	helpButton := widget.NewButton("使用说明", func() {
		// TODO: 实现使用说明
		c.app.logger.Info("显示使用说明")
	})

	rightButtons := container.NewHBox(
		autoUpdateCheck,
		donationButton,
		disclaimerButton,
		helpButton,
	)

	return container.NewHBox(
		versionLabel,
		widget.NewSeparator(),
		rightButtons,
	)
}

// RefreshInterfaceList 刷新接口列表
func (c *Components) RefreshInterfaceList() {
	interfaces := c.app.captureManager.GetAvailableInterfaces()
	c.interfaceSelect.Options = interfaces

	// 尝试选择默认接口
	if defaultInterface, exists := c.app.captureManager.GetDefaultInterface(); exists {
		c.interfaceSelect.SetSelected(defaultInterface)
	}

	c.app.logger.Info(fmt.Sprintf("已加载 %d 个网络接口", len(interfaces)))
}

// GetSelectedInterfaces 获取选择的接口
func (c *Components) GetSelectedInterfaces() []string {
	// 如果选择了监听所有接口
	if c.listenAllCheck.Checked {
		return c.app.captureManager.GetAvailableInterfaces()
	}

	// 否则返回选择的单个接口
	selected := c.interfaceSelect.Selected
	if selected == "" {
		return []string{}
	}
	return []string{selected}
}

// SetCaptureStatus 设置捕获状态
func (c *Components) SetCaptureStatus(capturing bool) {
	// 初始化progressBar（如果尚未初始化）
	if c.progressBar == nil {
		c.progressBar = widget.NewProgressBarInfinite()
		c.progressBar.Hide()
	}

	if capturing {
		c.captureButton.SetText("停止捕获")
		c.captureButton.Importance = widget.DangerImportance
		c.statusLabel.SetText("正在捕获...")
		c.progressBar.Show()
		c.progressBar.Start() // 无限进度条有Start()方法
	} else {
		c.captureButton.SetText("开始捕获")
		c.captureButton.Importance = widget.HighImportance
		c.statusLabel.SetText("已停止")
		c.progressBar.Stop() // 无限进度条有Stop()方法
		c.progressBar.Hide()
	}
	c.captureButton.Refresh()
}

// UpdateCaptureResult 更新捕获结果
func (c *Components) UpdateCaptureResult(rtmpInfo *rtmp.RTMPInfo) {
	if rtmpInfo.ServerAddress != "" && c.serverAddressEntry.Text == "" {
		c.serverAddressEntry.SetText(rtmpInfo.ServerAddress)
	}

	if rtmpInfo.StreamCode != "" && c.streamCodeEntry.Text == "" {
		c.streamCodeEntry.SetText(rtmpInfo.StreamCode)
	}

	if rtmpInfo.FullURL != "" && c.fullURLEntry.Text == "" {
		c.fullURLEntry.SetText(rtmpInfo.FullURL)
	}
}

// UpdateLogs 更新日志显示
func (c *Components) UpdateLogs() {
	// 更新操作日志
	consoleLogs := c.app.logger.GetFormattedConsoleLogs()
	if len(consoleLogs) > 0 {
		consoleText := strings.Join(consoleLogs, "\n")
		c.consoleLog.ParseMarkdown(fmt.Sprintf("```\n%s\n```", consoleText))
	}

	// 更新数据包日志
	packetLogs := c.app.logger.GetFormattedPacketLogs()
	if len(packetLogs) > 0 {
		packetText := strings.Join(packetLogs, "\n")
		c.packetLog.ParseMarkdown(fmt.Sprintf("```\n%s\n```", packetText))
	}
}

// UpdateStats 更新统计信息
func (c *Components) UpdateStats() {
	stats := c.app.GetCaptureStats()
	if stats != nil {
		statsText := fmt.Sprintf("数据包: %d | 接口: %v",
			stats.TotalPackets,
			len(stats.ActiveInterfaces))
		c.statsLabel.SetText(statsText)

		if stats.RTMPFound {
			c.statusLabel.SetText("✅ 已获取推流地址")
		} else if c.app.IsCapturing() {
			c.statusLabel.SetText("🔍 正在搜索推流信息...")
		}
	}
}

// StartStatsUpdate 开始统计信息更新
func (c *Components) StartStatsUpdate() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if c.app.IsCapturing() {
				c.UpdateStats()
				c.UpdateLogs()
			}
		}
	}()
}

// ResetResults 重置结果显示
func (c *Components) ResetResults() {
	c.serverAddressEntry.SetText("")
	c.streamCodeEntry.SetText("")
	c.fullURLEntry.SetText("")
}

// ShowMessage 显示消息
func (c *Components) ShowMessage(message string, msgType MessageType) {
	var prefix string
	switch msgType {
	case MessageInfo:
		prefix = "ℹ️ "
	case MessageSuccess:
		prefix = "✅ "
	case MessageWarning:
		prefix = "⚠️ "
	case MessageError:
		prefix = "❌ "
	}

	c.statusLabel.SetText(prefix + message)
}

// MessageType 消息类型
type MessageType int

const (
	MessageInfo MessageType = iota
	MessageSuccess
	MessageWarning
	MessageError
)

// EnableControls 启用/禁用控件
func (c *Components) EnableControls(enabled bool) {
	if enabled {
		c.interfaceSelect.Enable()
		c.refreshButton.Enable()
		c.testButton.Enable()
	} else {
		c.interfaceSelect.Disable()
		c.refreshButton.Disable()
		c.testButton.Disable()
	}
}
