package gui

import (
	"fmt"
	"strings"

	"douyin-rtmp/internal/capture"
	"douyin-rtmp/internal/network"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ControlPanel 控制面板组件
type ControlPanel struct {
	app                *App
	networkInterface   *network.NetworkInterface
	capture            *capture.PacketCapture
	logCapture         *capture.LogCapture

	// UI组件
	interfaceSelect    *widget.Select
	captureBtn         *widget.Button
	testBtn            *widget.Button
	statusLabel        *widget.Label
	listeningAllCheck  *widget.Check
	fileModeCheck      *widget.Check
	serverEntry        *widget.Entry
	streamEntry        *widget.Entry
	copyServerBtn      *widget.Button
	copyStreamBtn      *widget.Button

	// 状态变量
	isCapturing        bool
	selectedInterface  string
	listeningAll       bool
	fileMode           bool
	serverAddress      string
	streamCode         string
}

// NewControlPanel 创建新的控制面板
func NewControlPanel(app *App) *ControlPanel {
	cp := &ControlPanel{
		app:               app,
		networkInterface:  network.NewNetworkInterface(app.logger),
		capture:           capture.NewPacketCapture(app.logger),
		logCapture:        capture.NewLogCapture(app.logger),
		isCapturing:       false,
		listeningAll:      true, // 默认监听所有接口
		fileMode:          false, // 默认不启用日志模式
	}

	// 设置回调函数
	cp.capture.AddCallback(cp.updateStreamURL)
	cp.logCapture.AddCallback(cp.updateStreamURL)

	return cp
}

// CreateContent 创建控制面板内容
func (cp *ControlPanel) CreateContent() fyne.CanvasObject {
	// 创建控制面板框架
	controlFrame := container.NewVBox()

	// 网络接口选择区域
	interfaceContainer := container.NewGridWithColumns(3,
		widget.NewLabel("网络接口:"),
		cp.createInterfaceSelect(),
		cp.createRefreshButton(),
	)

	// 控制按钮区域
	controlContainer := container.NewHBox(
		cp.createCaptureButton(),
		cp.createTestButton(),
		widget.NewLabel("状态:"),
		cp.createStatusLabel(),
		cp.createListeningAllCheck(),
		cp.createFileModeCheck(),
	)

	// 服务器地址区域
	serverContainer := container.NewGridWithColumns(3,
		widget.NewLabel("推流服务器:"),
		cp.createServerEntry(),
		cp.createCopyServerButton(),
	)

	// 推流码区域
	streamContainer := container.NewGridWithColumns(3,
		widget.NewLabel("推流码:"),
		cp.createStreamEntry(),
		cp.createCopyStreamButton(),
	)

	// 组合所有区域
	controlFrame.Add(interfaceContainer)
	controlFrame.Add(controlContainer)
	controlFrame.Add(serverContainer)
	controlFrame.Add(streamContainer)

	return container.NewBorder(
		widget.NewCard("控制面板", "", controlFrame),
		nil, nil, nil,
	)
}

// createInterfaceSelect 创建接口选择框
func (cp *ControlPanel) createInterfaceSelect() *widget.Select {
	cp.interfaceSelect = widget.NewSelect([]string{}, func(selected string) {
		cp.selectedInterface = selected
	})
	cp.interfaceSelect.PlaceHolder = "请选择网络接口"
	return cp.interfaceSelect
}

// createRefreshButton 创建刷新按钮
func (cp *ControlPanel) createRefreshButton() *widget.Button {
	refreshBtn := widget.NewButton("刷新", func() {
		cp.RefreshInterfaces()
	})
	return refreshBtn
}

// createCaptureButton 创建捕获按钮
func (cp *ControlPanel) createCaptureButton() *widget.Button {
	cp.captureBtn = widget.NewButton("开始捕获", func() {
		cp.toggleCapture()
	})
	return cp.captureBtn
}

// createTestButton 创建测试按钮
func (cp *ControlPanel) createTestButton() *widget.Button {
	cp.testBtn = widget.NewButton("检测可用", func() {
		cp.testCapture()
	})
	return cp.testBtn
}

// createStatusLabel 创建状态标签
func (cp *ControlPanel) createStatusLabel() *widget.Label {
	cp.statusLabel = widget.NewLabel("待开始")
	return cp.statusLabel
}

// createListeningAllCheck 创建监听所有接口复选框
func (cp *ControlPanel) createListeningAllCheck() *widget.Check {
	cp.listeningAllCheck = widget.NewCheck("监听所有接口", func(checked bool) {
		cp.listeningAll = checked
		cp.onListeningChanged()
		cp.saveListeningConfig()
	})
	cp.listeningAllCheck.SetChecked(cp.listeningAll)
	return cp.listeningAllCheck
}

// createFileModeCheck 创建文件模式复选框
func (cp *ControlPanel) createFileModeCheck() *widget.Check {
	cp.fileModeCheck = widget.NewCheck("日志模式（备选）", func(checked bool) {
		cp.fileMode = checked
		cp.fileModeChanged()
	})
	cp.fileModeCheck.SetChecked(cp.fileMode)
	return cp.fileModeCheck
}

// createServerEntry 创建服务器地址输入框
func (cp *ControlPanel) createServerEntry() *widget.Entry {
	cp.serverEntry = widget.NewEntry()
	cp.serverEntry.Disable() // 只读
	return cp.serverEntry
}

// createStreamEntry 创建推流码输入框
func (cp *ControlPanel) createStreamEntry() *widget.Entry {
	cp.streamEntry = widget.NewEntry()
	cp.streamEntry.Disable() // 只读
	return cp.streamEntry
}

// createCopyServerButton 创建复制服务器地址按钮
func (cp *ControlPanel) createCopyServerButton() *widget.Button {
	cp.copyServerBtn = widget.NewButton("复制", func() {
		cp.copyToClipboard(cp.serverAddress)
	})
	return cp.copyServerBtn
}

// createCopyStreamButton 创建复制推流码按钮
func (cp *ControlPanel) createCopyStreamButton() *widget.Button {
	cp.copyStreamBtn = widget.NewButton("复制", func() {
		cp.copyToClipboard(cp.streamCode)
	})
	return cp.copyStreamBtn
}

// LoadInterfaces 加载网络接口列表
func (cp *ControlPanel) LoadInterfaces() {
	cp.app.logger.Info("正在加载网络接口列表...")

	result := cp.networkInterface.LoadInterfaces()
	interfaces := result["interfaces"].([]string)

	if len(interfaces) > 0 {
		// 更新下拉列表
		cp.interfaceSelect.Options = interfaces

		// 设置默认选择
		if defaultInterface, ok := result["default"].(string); ok && defaultInterface != "" {
			cp.interfaceSelect.SetSelected(defaultInterface)
			cp.selectedInterface = defaultInterface
			cp.app.logger.Info(fmt.Sprintf("自动选择以太网接口: %s", defaultInterface))
		} else {
			cp.interfaceSelect.SetSelected(interfaces[0])
			cp.selectedInterface = interfaces[0]
			cp.app.logger.Info(fmt.Sprintf("自动选择第一个可用接口: %s", interfaces[0]))
		}

		activeCount := result["active_count"].(int)
		cp.app.logger.Info(fmt.Sprintf("共找到 %d 个网络接口", len(interfaces)))
		cp.app.logger.Info(fmt.Sprintf("其中活动接口 %d 个", activeCount))
	} else {
		cp.app.logger.Error("未找到可用的网络接口")
	}
}

// RefreshInterfaces 刷新网络接口列表
func (cp *ControlPanel) RefreshInterfaces() {
	cp.app.logger.Info("正在刷新网络接口列表...")
	cp.LoadInterfaces()
	cp.app.logger.Info("网络接口列表刷新完成")
}

// toggleCapture 切换捕获状态
func (cp *ControlPanel) toggleCapture() {
	if !cp.isCapturing {
		// 开始捕获
		cp.startCapture()
	} else {
		// 停止捕获
		cp.stopCapture()
	}
}

// startCapture 开始捕获
func (cp *ControlPanel) startCapture() {
	cp.isCapturing = true
	cp.captureBtn.SetText("停止捕获")
	cp.interfaceSelect.Disable()
	cp.statusLabel.SetText("正在捕获")

	// 清空原有推流地址和推流码
	cp.updateStreamURL("", "")

	if cp.fileMode {
		// 启动日志模式抓取推流码
		cp.logCapture.Start()
		cp.app.logger.Info("已开启日志模式抓取推流")
	} else {
		cp.app.logger.Info("已开启抓包模式抓取推流")
		
		// 结束 MediaSDK_Server 进程（如果存在）
		// TODO: 实现进程终止逻辑

		if cp.listeningAll {
			// 获取所有接口的实际名称
			selectedInterfaces := []string{}
			for _, ifaceDisplay := range cp.interfaceSelect.Options {
				// 从显示名称中提取实际的接口名称
				actualName := strings.Split(ifaceDisplay, " [")[0]
				actualName = strings.TrimSpace(actualName)
				selectedInterfaces = append(selectedInterfaces, actualName)
			}
			// 启动多接口捕获
			cp.capture.StartMulti(selectedInterfaces)
		} else {
			// 获取选中接口的实际名称
			selectedDisplay := cp.selectedInterface
			if selectedDisplay == "" {
				dialog.ShowError(fmt.Errorf("请先选择网络接口"), cp.app.window)
				cp.stopCapture()
				return
			}
			// 从显示名称中提取实际的接口名称
			actualName := strings.Split(selectedDisplay, " [")[0]
			actualName = strings.TrimSpace(actualName)
			// 启动单接口捕获
			cp.capture.Start(actualName)
		}
	}
}

// stopCapture 停止捕获
func (cp *ControlPanel) stopCapture() {
	cp.isCapturing = false
	cp.captureBtn.SetText("开始捕获")
	cp.statusLabel.SetText("已停止")
	cp.onListeningChanged()

	// 停止捕获
	if cp.fileMode {
		cp.logCapture.Stop()
		cp.app.logger.Info("已停止日志模式抓取推流")
	} else {
		cp.app.logger.Info("已停止抓包模式抓取推流")
		cp.capture.Stop()
	}
}

// testCapture 测试接口捕获
func (cp *ControlPanel) testCapture() {
	if cp.isCapturing {
		dialog.ShowWarning("警告", "请先停止当前捕获", cp.app.window)
		return
	}

	cp.statusLabel.SetText("正在检测...")
	cp.testBtn.Disable()
	cp.captureBtn.Disable()
	cp.app.logger.Info("开始检测接口...")

	onTestComplete := func(hasData bool) {
		cp.statusLabel.SetText("待开始")
		cp.testBtn.Enable()
		cp.captureBtn.Enable()
		if hasData {
			dialog.ShowInformation("检测结果", 
				"接口可用，已检测到数据流。工具可正常使用，如果不能捕获到推流码，请检查本地是否开了过多软件（如浏览器看直播、视频等），或请尝试勾选日志模式后抓取推流信息。", 
				cp.app.window)
		} else {
			dialog.ShowInformation("检测结果", 
				"接口可能不可用，未检测到数据流，请更换其他接口尝试，如果勾选了监听所有接口，则本工具可能在您的环境下可能无法正常使用，或请尝试勾选日志模式后抓取推流信息。", 
				cp.app.window)
		}
	}

	// 获取要测试的接口
	if cp.listeningAll {
		selectedInterfaces := []string{}
		for _, ifaceDisplay := range cp.interfaceSelect.Options {
			actualName := strings.Split(ifaceDisplay, " [")[0]
			actualName = strings.TrimSpace(actualName)
			selectedInterfaces = append(selectedInterfaces, actualName)
		}
		// 启动多接口测试
		cp.capture.TestCapture(selectedInterfaces, onTestComplete)
	} else {
		selectedDisplay := cp.selectedInterface
		if selectedDisplay == "" {
			dialog.ShowError(fmt.Errorf("请先选择网络接口"), cp.app.window)
			return
		}
		actualName := strings.Split(selectedDisplay, " [")[0]
		actualName = strings.TrimSpace(actualName)
		// 启动单接口测试
		cp.capture.TestCapture([]string{actualName}, onTestComplete)
	}
}

// updateStreamURL 更新推流地址和推流码
func (cp *ControlPanel) updateStreamURL(serverAddress, streamCode string) {
	cp.serverAddress = serverAddress
	cp.streamCode = streamCode
	
	cp.serverEntry.SetText(serverAddress)
	cp.streamEntry.SetText(streamCode)

	// 如果获取到了地址和推流码，更新界面状态
	if serverAddress != "" && streamCode != "" {
		cp.stopCapture()
	}
}

// copyToClipboard 复制内容到剪贴板
func (cp *ControlPanel) copyToClipboard(text string) {
	if strings.TrimSpace(text) == "" {
		dialog.ShowInformation("提示", "没有内容可复制，请先按说明进行地址捕获", cp.app.window)
		return
	}

	// TODO: 实现剪贴板操作
	// fyne.CurrentApp().Driver().GetClipboard().SetContent(text)
	dialog.ShowInformation("成功", "已复制到剪贴板", cp.app.window)
}

// onListeningChanged 监听设置改变时的回调
func (cp *ControlPanel) onListeningChanged() {
	cp.saveListeningConfig()

	// 根据监听所有接口的状态设置接口选择的状态
	if cp.listeningAll {
		cp.interfaceSelect.Disable()
	} else {
		if !cp.isCapturing {
			cp.interfaceSelect.Enable()
		}
	}
}

// saveListeningConfig 保存监听配置
func (cp *ControlPanel) saveListeningConfig() {
	cp.app.configManager.SetListeningAll(cp.listeningAll)
}

// loadListeningConfig 加载监听配置
func (cp *ControlPanel) loadListeningConfig() {
	listeningAll := cp.app.configManager.GetListeningAll()
	cp.listeningAll = listeningAll
	cp.listeningAllCheck.SetChecked(listeningAll)
	cp.onListeningChanged()
}

// fileModeChanged 文件模式改变时的回调
func (cp *ControlPanel) fileModeChanged() {
	cp.app.configManager.SetFileMode(cp.fileMode)
}

// loadFileModeConfig 加载文件模式配置
func (cp *ControlPanel) loadFileModeConfig() {
	fileMode := cp.app.configManager.GetFileMode()
	cp.fileMode = fileMode
	cp.fileModeCheck.SetChecked(fileMode)
}

// SetCaptureStatus 设置捕获状态
func (cp *ControlPanel) SetCaptureStatus(capturing bool) {
	if capturing {
		cp.captureBtn.SetText("停止捕获")
		cp.statusLabel.SetText("正在捕获")
		cp.interfaceSelect.Disable()
	} else {
		cp.captureBtn.SetText("开始捕获")
		cp.statusLabel.SetText("已停止")
		cp.onListeningChanged()
	}
	cp.isCapturing = capturing
}

// GetSelectedInterfaces 获取选择的接口
func (cp *ControlPanel) GetSelectedInterfaces() []string {
	if cp.listeningAll {
		selectedInterfaces := []string{}
		for _, ifaceDisplay := range cp.interfaceSelect.Options {
			actualName := strings.Split(ifaceDisplay, " [")[0]
			actualName = strings.TrimSpace(actualName)
			selectedInterfaces = append(selectedInterfaces, actualName)
		}
		return selectedInterfaces
	} else {
		if cp.selectedInterface == "" {
			return []string{}
		}
		actualName := strings.Split(cp.selectedInterface, " [")[0]
		actualName = strings.TrimSpace(actualName)
		return []string{actualName}
	}
}

// UpdateCaptureResult 更新捕获结果
func (cp *ControlPanel) UpdateCaptureResult(rtmpInfo *capture.RTMPInfo) {
	cp.updateStreamURL(rtmpInfo.ServerAddress, rtmpInfo.StreamCode)
}

// RefreshInterfaceList 刷新接口列表
func (cp *ControlPanel) RefreshInterfaceList() {
	cp.LoadInterfaces()
}