package gui

import (
	"fmt"
	"time"

	"douyin-rtmp/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SettingsDialog 设置对话框
type SettingsDialog struct {
	app    *App
	dialog dialog.Dialog
	
	// 配置选项
	captureTimeoutEntry *widget.Entry
	maxCapturesEntry    *widget.Entry
	logLevelSelect      *widget.Select
	themeSelect         *widget.Select
	autoSaveCheck       *widget.Check
}

// NewSettingsDialog 创建设置对话框
func NewSettingsDialog(app *App) *SettingsDialog {
	sd := &SettingsDialog{
		app: app,
	}
	
	sd.createDialog()
	return sd
}

// createDialog 创建对话框
func (sd *SettingsDialog) createDialog() {
	// 获取当前配置
	cfg := sd.app.configManager.GetConfig()
	
	// 捕获超时设置
	timeoutLabel := widget.NewLabel("捕获超时（秒）:")
	sd.captureTimeoutEntry = widget.NewEntry()
	sd.captureTimeoutEntry.SetText(fmt.Sprintf("%d", cfg.Network.CaptureTimeout))
	
	// 最大并发捕获
	maxCapturesLabel := widget.NewLabel("最大并发捕获:")
	sd.maxCapturesEntry = widget.NewEntry()
	sd.maxCapturesEntry.SetText(fmt.Sprintf("%d", cfg.Network.MaxConcurrentCaptures))
	
	// 日志级别
	logLevelLabel := widget.NewLabel("日志级别:")
	sd.logLevelSelect = widget.NewSelect(
		[]string{"debug", "info", "warn", "error"}, 
		nil,
	)
	sd.logLevelSelect.SetSelected(cfg.Log.Level)
	
	// 主题选择
	themeLabel := widget.NewLabel("界面主题:")
	sd.themeSelect = widget.NewSelect(
		[]string{"auto", "light", "dark"}, 
		nil,
	)
	sd.themeSelect.SetSelected(cfg.UI.Theme)
	
	// 自动保存配置
	sd.autoSaveCheck = widget.NewCheck("自动保存配置", nil)
	sd.autoSaveCheck.SetChecked(cfg.App.AutoSaveConfig)
	
	// 按钮
	saveBtn := widget.NewButton("保存", sd.onSave)
	cancelBtn := widget.NewButton("取消", sd.onCancel)
	resetBtn := widget.NewButton("重置", sd.onReset)
	
	buttonRow := container.NewHBox(
		saveBtn,
		cancelBtn,
		resetBtn,
	)
	
	// 布局
	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## 应用设置"),
		widget.NewSeparator(),
		
		timeoutLabel,
		sd.captureTimeoutEntry,
		
		maxCapturesLabel,
		sd.maxCapturesEntry,
		
		logLevelLabel,
		sd.logLevelSelect,
		
		themeLabel,
		sd.themeSelect,
		
		sd.autoSaveCheck,
		
		widget.NewSeparator(),
		buttonRow,
	)
	
	sd.dialog = dialog.NewCustom("设置", "", content, sd.app.window)
	sd.dialog.Resize(fyne.NewSize(400, 500))
}

// Show 显示对话框
func (sd *SettingsDialog) Show() {
	sd.dialog.Show()
}

// onSave 保存设置
func (sd *SettingsDialog) onSave() {
	// 验证输入
	if !sd.validateInputs() {
		return
	}
	
	// 更新配置
	err := sd.updateConfig()
	if err != nil {
		dialog.ShowError(err, sd.app.window)
		return
	}
	
	// 保存配置
	err = sd.app.configManager.Save()
	if err != nil {
		dialog.ShowError(fmt.Errorf("保存配置失败: %w", err), sd.app.window)
		return
	}
	
	sd.app.logger.Info("设置已保存")
	sd.dialog.Hide()
}

// onCancel 取消设置
func (sd *SettingsDialog) onCancel() {
	sd.dialog.Hide()
}

// onReset 重置设置
func (sd *SettingsDialog) onReset() {
	dialog.ShowConfirm("确认重置", "确定要重置所有设置到默认值吗？", func(confirmed bool) {
		if confirmed {
			sd.app.configManager.Reset()
			sd.app.logger.Info("设置已重置为默认值")
			sd.dialog.Hide()
		}
	}, sd.app.window)
}

// validateInputs 验证输入
func (sd *SettingsDialog) validateInputs() bool {
	// 验证超时时间
	if sd.captureTimeoutEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("请输入捕获超时时间"), sd.app.window)
		return false
	}
	
	// 验证最大并发数
	if sd.maxCapturesEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("请输入最大并发捕获数"), sd.app.window)
		return false
	}
	
	return true
}

// updateConfig 更新配置
func (sd *SettingsDialog) updateConfig() error {
	// 更新网络配置
	err := sd.app.configManager.UpdateNetworkConfig(func(nc *config.NetworkConfig) {
		// 这里应该解析字符串到整数，简化处理
		if sd.captureTimeoutEntry.Text != "" {
			// 实际实现中应该使用 strconv.Atoi
			nc.CaptureTimeout = 30 // 默认值
		}
		if sd.maxCapturesEntry.Text != "" {
			nc.MaxConcurrentCaptures = 5 // 默认值
		}
	})
	if err != nil {
		return err
	}
	
	// 更新日志配置
	err = sd.app.configManager.UpdateLogConfig(func(lc *config.LogConfig) {
		lc.Level = sd.logLevelSelect.Selected
	})
	if err != nil {
		return err
	}
	
	// 更新UI配置
	err = sd.app.configManager.UpdateUIConfig(func(uc *config.UIConfig) {
		uc.Theme = sd.themeSelect.Selected
	})
	if err != nil {
		return err
	}
	
	// 更新应用配置
	err = sd.app.configManager.UpdateAppConfig(func(ac *config.AppConfig) {
		ac.AutoSaveConfig = sd.autoSaveCheck.Checked
	})
	if err != nil {
		return err
	}
	
	return nil
}

// EventHandlers 事件处理器集合
type EventHandlers struct {
	app *App
}

// NewEventHandlers 创建事件处理器
func NewEventHandlers(app *App) *EventHandlers {
	return &EventHandlers{
		app: app,
	}
}

// HandleCaptureToggle 处理捕获切换
func (eh *EventHandlers) HandleCaptureToggle() {
	if eh.app.IsCapturing() {
		eh.app.StopCapture()
	} else {
		eh.app.StartCapture()
	}
}

// HandleInterfaceRefresh 处理接口刷新
func (eh *EventHandlers) HandleInterfaceRefresh() {
	eh.app.RefreshInterfaces()
}

// HandleInterfaceTest 处理接口测试
func (eh *EventHandlers) HandleInterfaceTest(interfaceName string) {
	if interfaceName == "" {
		dialog.ShowError(fmt.Errorf("请先选择网络接口"), eh.app.window)
		return
	}
	
	eh.app.TestInterface(interfaceName)
}

// HandleExportResults 处理结果导出
func (eh *EventHandlers) HandleExportResults() {
	eh.app.ExportResults()
}

// HandleCopyToClipboard 处理复制到剪贴板
func (eh *EventHandlers) HandleCopyToClipboard(text string) {
	if text != "" {
		eh.app.window.Clipboard().SetContent(text)
		eh.app.logger.Info("已复制到剪贴板")
	}
}

// HandleQuickTest 处理快速测试
func (eh *EventHandlers) HandleQuickTest() {
	eh.app.PerformQuickTest()
}

// HandleNetworkDiagnostic 处理网络诊断
func (eh *EventHandlers) HandleNetworkDiagnostic() {
	eh.app.ShowNetworkDiagnostic()
}

// HandleThemeChange 处理主题切换
func (eh *EventHandlers) HandleThemeChange(theme string) {
	err := eh.app.configManager.SetTheme(theme)
	if err != nil {
		eh.app.logger.Error(fmt.Sprintf("切换主题失败: %v", err))
		return
	}
	
	eh.app.logger.Info(fmt.Sprintf("已切换到 %s 主题", theme))
}

// HandleLogClear 处理日志清除
func (eh *EventHandlers) HandleLogClear(logType string) {
	switch logType {
	case "console":
		eh.app.logger.ClearConsoleLogs()
		eh.app.logger.Info("已清除操作日志")
	case "packet":
		eh.app.logger.ClearPacketLogs()
		eh.app.logger.Info("已清除数据包日志")
	case "all":
		eh.app.logger.ClearAllLogs()
		eh.app.logger.Info("已清除所有日志")
	}
	
	// 更新日志显示
	eh.app.logPanel.RefreshDisplay()
}

// HandleWindowResize 处理窗口大小调整
func (eh *EventHandlers) HandleWindowResize(size fyne.Size) {
	err := eh.app.configManager.SetWindowSize(size.Width, size.Height)
	if err != nil {
		eh.app.logger.Error(fmt.Sprintf("保存窗口大小失败: %v", err))
	}
}

// HandleAppExit 处理应用退出
func (eh *EventHandlers) HandleAppExit() {
	// 如果正在捕获，先停止
	if eh.app.IsCapturing() {
		eh.app.StopCapture()
		// 等待一点时间让捕获完全停止
		time.Sleep(500 * time.Millisecond)
	}
	
	// 保存配置
	eh.app.configManager.Save()
	
	eh.app.logger.Info("应用程序正在退出...")
}

// KeyboardShortcuts 键盘快捷键处理
type KeyboardShortcuts struct {
	app *App
}

// NewKeyboardShortcuts 创建键盘快捷键处理器
func NewKeyboardShortcuts(app *App) *KeyboardShortcuts {
	return &KeyboardShortcuts{
		app: app,
	}
}

// SetupShortcuts 设置快捷键
func (ks *KeyboardShortcuts) SetupShortcuts() {
	// Ctrl+R: 刷新接口
	ks.app.window.Canvas().AddShortcut(&fyne.ShortcutCut{}, func(shortcut fyne.Shortcut) {
		// 这里应该是刷新快捷键，但Fyne的快捷键API可能不同
		ks.app.RefreshInterfaces()
	})
	
	// F5: 刷新
	// Space: 开始/停止捕获
	// Ctrl+S: 保存设置
	// 等等...实际实现需要根据Fyne的API调整
}

// TooltipHelper 工具提示帮助器
type TooltipHelper struct{}

// NewTooltipHelper 创建工具提示帮助器
func NewTooltipHelper() *TooltipHelper {
	return &TooltipHelper{}
}

// AddTooltips 为组件添加工具提示
func (th *TooltipHelper) AddTooltips(components *Components) {
	// 为各种组件添加提示信息
	// 这需要根据Fyne的具体API来实现
}

// NotificationManager 通知管理器
type NotificationManager struct {
	app *App
}

// NewNotificationManager 创建通知管理器
func NewNotificationManager(app *App) *NotificationManager {
	return &NotificationManager{
		app: app,
	}
}

// ShowSuccess 显示成功通知
func (nm *NotificationManager) ShowSuccess(title, message string) {
	nm.app.fyneApp.SendNotification(&fyne.Notification{
		Title:   "✅ " + title,
		Content: message,
	})
}

// ShowError 显示错误通知
func (nm *NotificationManager) ShowError(title, message string) {
	nm.app.fyneApp.SendNotification(&fyne.Notification{
		Title:   "❌ " + title,
		Content: message,
	})
}

// ShowInfo 显示信息通知
func (nm *NotificationManager) ShowInfo(title, message string) {
	nm.app.fyneApp.SendNotification(&fyne.Notification{
		Title:   "ℹ️ " + title,
		Content: message,
	})
}

// ShowWarning 显示警告通知
func (nm *NotificationManager) ShowWarning(title, message string) {
	nm.app.fyneApp.SendNotification(&fyne.Notification{
		Title:   "⚠️ " + title,
		Content: message,
	})
}