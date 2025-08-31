package gui

import (
	"fmt"
	"strings"
	"time"

	"douyin-rtmp/internal/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LogType 日志类型
type LogType int

const (
	LogTypeSystem LogType = iota
	LogTypePacket
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogPanel 日志面板组件
type LogPanel struct {
	app           *App
	
	// UI组件
	notebook      *container.AppTabs
	systemConsole *widget.RichText
	packetConsole *widget.RichText
	clearSystemBtn *widget.Button
	clearPacketBtn *widget.Button
}

// NewLogPanel 创建新的日志面板
func NewLogPanel(app *App) *LogPanel {
	lp := &LogPanel{
		app: app,
	}
	
	return lp
}

// CreateContent 创建日志面板内容
func (lp *LogPanel) CreateContent() fyne.CanvasObject {
	// 创建标签页控件
	lp.notebook = container.NewAppTabs()
	
	// 创建系统日志标签页
	systemTab := lp.createSystemLogTab()
	lp.notebook.AppendTab(container.NewTabItem("控制台输出", systemTab))
	
	// 创建数据包日志标签页
	packetTab := lp.createPacketLogTab()
	lp.notebook.AppendTab(container.NewTabItem("数据包监控", packetTab))
	
	// 设置日志控件
	lp.app.logger.SetConsoles(lp.systemConsole, lp.packetConsole)
	
	return lp.notebook
}

// RefreshDisplay 刷新日志显示（清除日志后更新界面）
func (lp *LogPanel) RefreshDisplay() {
	// 获取当前的日志内容
	consoleLogs := lp.app.logger.GetFormattedConsoleLogs()
	packetLogs := lp.app.logger.GetFormattedPacketLogs()
	
	// 更新系统控制台显示
	lp.systemConsole.ParseMarkdown(strings.Join(consoleLogs, "\n"))
	
	// 更新数据包控制台显示
	lp.packetConsole.ParseMarkdown(strings.Join(packetLogs, "\n"))
}

// createSystemLogTab 创建系统日志标签页
func (lp *LogPanel) createSystemLogTab() fyne.CanvasObject {
	// 创建系统日志显示区域
	lp.systemConsole = widget.NewRichText()
	lp.systemConsole.Wrapping = fyne.TextWrapWord
	lp.systemConsole.Scroll = container.ScrollVerticalOnly
	
	// 创建清除按钮
	lp.clearSystemBtn = widget.NewButton("清除控制台", func() {
		lp.clearSystemConsole()
	})
	
	// 创建滚动容器
	scrollContainer := container.NewScroll(lp.systemConsole)
	scrollContainer.SetMinSize(fyne.NewSize(400, 300))
	
	// 组合布局
	return container.NewBorder(
		nil,                    // top
		lp.clearSystemBtn,      // bottom
		nil,                    // left
		nil,                    // right
		scrollContainer,        // center
	)
}

// createPacketLogTab 创建数据包日志标签页
func (lp *LogPanel) createPacketLogTab() fyne.CanvasObject {
	// 创建数据包日志显示区域
	lp.packetConsole = widget.NewRichText()
	lp.packetConsole.Wrapping = fyne.TextWrapWord
	lp.packetConsole.Scroll = container.ScrollVerticalOnly
	
	// 创建清除按钮
	lp.clearPacketBtn = widget.NewButton("清除数据包日志", func() {
		lp.clearPacketConsole()
	})
	
	// 创建滚动容器
	scrollContainer := container.NewScroll(lp.packetConsole)
	scrollContainer.SetMinSize(fyne.NewSize(400, 300))
	
	// 组合布局
	return container.NewBorder(
		nil,                    // top
		lp.clearPacketBtn,      // bottom
		nil,                    // left
		nil,                    // right
		scrollContainer,        // center
	)
}

// clearSystemConsole 清除系统控制台内容
func (lp *LogPanel) clearSystemConsole() {
	lp.app.logger.ClearConsole()
	lp.app.logger.Info("系统日志已清除")
}

// clearPacketConsole 清除数据包控制台内容
func (lp *LogPanel) clearPacketConsole() {
	lp.app.logger.ClearPacketConsole()
	lp.app.logger.Info("数据包日志已清除") // 在主控制台显示清除提示
}

// UpdateLogs 更新日志显示
func (lp *LogPanel) UpdateLogs(entry *logger.LogEntry) {
	// 根据日志类型决定在哪个控制台显示
	// 由于我们直接使用logger的日志级别，需要判断是否为数据包日志
	// 这里我们通过消息内容来判断
	if lp.isPacketLog(entry.Message) {
		// 数据包日志显示在数据包控制台
		lp.appendToPacketConsole(entry)
		
		// 如果发现关键数据包，自动切换到数据包监控标签
		if lp.isImportantPacketLog(entry.Message) {
			lp.notebook.SelectTab(lp.notebook.Items[1]) // 切换到数据包监控标签
		}
	} else {
		// 系统日志显示在系统控制台
		lp.appendToSystemConsole(entry)
	}
}

// isPacketLog 判断是否为数据包日志
func (lp *LogPanel) isPacketLog(message string) bool {
	// 通过消息内容判断是否为数据包日志
	return strings.Contains(message, ":") && 
	       (strings.Contains(message, "->") || 
	        strings.Contains(message, "找到推流") || 
	        strings.Contains(message, "发现"))
}

// isImportantPacketLog 判断是否为重要数据包日志
func (lp *LogPanel) isImportantPacketLog(message string) bool {
	return strings.Contains(message, ">>> 发现") || 
	       strings.Contains(message, "找到推流服务器地址") || 
	       strings.Contains(message, "找到推流码")
}

// appendToSystemConsole 添加到系统控制台
func (lp *LogPanel) appendToSystemConsole(entry *logger.LogEntry) {
	// 格式化日志消息
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	levelStr := lp.formatLogLevel(entry.Level)
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, levelStr, entry.Message)
	
	// 添加到系统控制台
	currentText := lp.systemConsole.String()
	lp.systemConsole.ParseMarkdown(currentText + logMessage)
	
	// 自动滚动到底部
	lp.scrollToBottom(lp.systemConsole)
}

// appendToPacketConsole 添加到数据包控制台
func (lp *LogPanel) appendToPacketConsole(entry *logger.LogEntry) {
	// 格式化数据包日志消息
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, entry.Message)
	
	// 添加到数据包控制台
	currentText := lp.packetConsole.String()
	lp.packetConsole.ParseMarkdown(currentText + logMessage)
	
	// 自动滚动到底部
	lp.scrollToBottom(lp.packetConsole)
}

// formatLogLevel 格式化日志级别
func (lp *LogPanel) formatLogLevel(level logger.LogLevel) string {
	switch level {
	case logger.DEBUG:
		return "DEBUG"
	case logger.INFO:
		return "INFO"
	case logger.WARN:
		return "WARN"
	case logger.ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// scrollToBottom 滚动到底部
func (lp *LogPanel) scrollToBottom(richText *widget.RichText) {
	// Fyne的RichText组件会自动滚动到底部
	// 如果需要手动控制滚动，可以在这里实现
}

// LogToConsole 输出日志到控制台（保持与Python版本接口一致）
func (lp *LogPanel) LogToConsole(message string) {
	lp.app.logger.Info(message)
}

// LogPacket 记录数据包信息到数据包控制台（保持与Python版本接口一致）
func (lp *LogPanel) LogPacket(message string) {
	// 创建数据包日志条目
	entry := &logger.LogEntry{
		Level:   logger.INFO,
		Message: message,
		Time:    time.Now(),
	}
	
	lp.appendToPacketConsole(entry)
	
	// 如果发现关键数据包，自动切换到数据包监控标签
	if lp.isImportantPacketLog(message) {
		lp.notebook.SelectTab(lp.notebook.Items[1]) // 切换到数据包监控标签
	}
}

// ClearLogs 清除所有日志（保持与Python版本接口一致）
func (lp *LogPanel) ClearLogs() {
	lp.clearSystemConsole()
	lp.clearPacketConsole()
}

// GetNotebook 获取notebook引用（用于外部控制标签切换）
func (lp *LogPanel) GetNotebook() *container.AppTabs {
	return lp.notebook
}	case logger.INFO:
		return "INFO"
	case logger.WARN:
		return "WARN"
	case logger.ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// scrollToBottom 滚动到底部
func (lp *LogPanel) scrollToBottom(richText *widget.RichText) {
	// Fyne的RichText组件会自动滚动到底部
	// 如果需要手动控制滚动，可以在这里实现
}

// LogToConsole 输出日志到控制台（保持与Python版本接口一致）
func (lp *LogPanel) LogToConsole(message string) {
	lp.app.logger.Info(message)
}

// LogPacket 记录数据包信息到数据包控制台（保持与Python版本接口一致）
func (lp *LogPanel) LogPacket(message string) {
	// 创建数据包日志条目
	entry := &logger.LogEntry{
		Level:   logger.INFO,
		Message: message,
		Time:    time.Now(),
	}
	
	lp.appendToPacketConsole(entry)
	
	// 如果发现关键数据包，自动切换到数据包监控标签
	if lp.isImportantPacketLog(message) {
		lp.notebook.SelectTab(lp.notebook.Items[1]) // 切换到数据包监控标签
	}
}

// ClearLogs 清除所有日志（保持与Python版本接口一致）
func (lp *LogPanel) ClearLogs() {
	lp.clearSystemConsole()
	lp.clearPacketConsole()
}

// GetNotebook 获取notebook引用（用于外部控制标签切换）
func (lp *LogPanel) GetNotebook() *container.AppTabs {
	return lp.notebook
}