package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 添加缺失的方法到App结构

// IsCapturing 检查是否正在捕获
func (a *App) IsCapturing() bool {
	return a.isCapturing
}

// GetCaptureStats 获取捕获统计信息
func (a *App) GetCaptureStats() *CaptureStats {
	if !a.isCapturing {
		return nil
	}

	stats := a.captureManager.GetStats()
	if stats == nil {
		return &CaptureStats{
			TotalPackets:     0,
			ActiveInterfaces: []string{},
			RTMPFound:        false,
		}
	}

	return &CaptureStats{
		TotalPackets:     stats.TotalPackets,
		ActiveInterfaces: stats.ActiveInterfaces,
		RTMPFound:        false, // 需要检查是否找到RTMP信息
	}
}

// CaptureStats 捕获统计信息
type CaptureStats struct {
	TotalPackets     int64    `json:"total_packets"`
	ActiveInterfaces []string `json:"active_interfaces"`
	RTMPFound        bool     `json:"rtmp_found"`
}

// ShowSettings 显示设置对话框
func (a *App) ShowSettings() {
	dialog.ShowInformation("设置", "设置功能尚未实现", a.window)
}

// ShowAbout 显示关于对话框
func (a *App) ShowAbout() {
	content := widget.NewRichTextFromMarkdown(`
# 抖音直播推流地址获取工具 - Go版本

**版本**: v1.0.0  
**开发语言**: Go  
**GUI框架**: Fyne  

## 主要功能

- 网络数据包捕获和分析
- RTMP推流地址自动提取
- 多网络接口支持
- 实时日志监控

## 技术特性

- 跨平台支持
- 高性能数据包处理
- 现代化GUI界面
- 模块化架构设计

## 开源协议

本项目采用开源协议，详情请参考项目仓库。

---

© 2024 抖音RTMP工具项目组
`)

	dialog.ShowCustom("关于", "确定", content, a.window)
}

// ShowNetworkDiagnostic 显示网络诊断
func (a *App) ShowNetworkDiagnostic() {
	// 获取网络诊断信息
	diagnostic, err := a.captureManager.GetNetworkDiagnostic()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	// 格式化诊断信息
	info := fmt.Sprintf(`
# 网络诊断信息

## 系统信息
- 主机名: %v
- 操作系统: %v
- 平台: %v

## 接口统计
- 总接口数: %v
- 活跃接口: %v
- 以太网接口: %v
- WiFi接口: %v
- VPN接口: %v
- 虚拟接口: %v

## 连通性测试
正在执行连通性测试...
`,
		diagnostic["hostname"],
		diagnostic["os"],
		diagnostic["platform"],
		diagnostic["interface_stats"].(map[string]interface{})["total"],
		diagnostic["interface_stats"].(map[string]interface{})["active"],
		diagnostic["interface_stats"].(map[string]interface{})["ethernet"],
		diagnostic["interface_stats"].(map[string]interface{})["wifi"],
		diagnostic["interface_stats"].(map[string]interface{})["vpn"],
		diagnostic["interface_stats"].(map[string]interface{})["virtual"],
	)

	content := widget.NewRichTextFromMarkdown(info)
	dialog.ShowCustom("网络诊断", "关闭", content, a.window)
}

// ExportResults 导出结果
func (a *App) ExportResults() {
	// TODO: 实现结果导出功能
	dialog.ShowInformation("导出结果", "导出功能尚未实现", a.window)
}

// onWindowClosing 窗口关闭处理
func (a *App) onWindowClosing() {
	// 如果正在捕获，先停止
	if a.isCapturing {
		a.StopCapture()
	}

	// 保存窗口配置
	size := a.window.Content().Size()
	a.configManager.SaveWindowConfig(size.Width, size.Height, 0, 0)

	// 关闭应用
	a.fyneApp.Quit()
}

// 添加缺失的组件方法到Components结构

// CreateToolbar 保留工具栏方法（作为备用）
func (c *Components) CreateToolbar() fyne.CanvasObject {
	// 刷新按钮
	refreshAction := widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
		c.app.RefreshInterfaces()
	})

	// 设置按钮
	settingsAction := widget.NewToolbarAction(theme.SettingsIcon(), func() {
		c.app.ShowSettings()
	})

	// 关于按钮
	aboutAction := widget.NewToolbarAction(theme.InfoIcon(), func() {
		c.app.ShowAbout()
	})

	// 网络诊断按钮
	diagnosticAction := widget.NewToolbarAction(theme.ComputerIcon(), func() {
		c.app.ShowNetworkDiagnostic()
	})

	toolbar := widget.NewToolbar(
		refreshAction,
		widget.NewToolbarSeparator(),
		diagnosticAction,
		widget.NewToolbarSeparator(),
		settingsAction,
		aboutAction,
	)

	return toolbar
}

// 确保progressBar被初始化
func (c *Components) ensureProgressBar() {
	if c.progressBar == nil {
		c.progressBar = widget.NewProgressBarInfinite()
		c.progressBar.Hide()
	}
}

// SetCaptureStatus 确保progress bar被正确初始化
func (c *Components) SetCaptureStatusSafe(capturing bool) {
	c.ensureProgressBar()
	c.SetCaptureStatus(capturing)
}
