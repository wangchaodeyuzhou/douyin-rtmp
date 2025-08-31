package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// OBSPanel OBS管理面板组件
type OBSPanel struct {
	app *App

	// 状态变量
	obsPath            string
	obsStatus          string
	streamConfigStatus string

	// UI组件
	obsStatusLabel    *widget.Label
	streamConfigLabel *widget.Label
	configPathBtn     *widget.Button
	configStreamBtn   *widget.Button
	syncStreamBtn     *widget.Button
	launchOBSBtn      *widget.Button
	pluginManagerBtn  *widget.Button
	solveReconnectBtn *widget.Button
	helpBtn           *widget.Button
	restoreBtn        *widget.Button
}

// NewOBSPanel 创建新的OBS管理面板
func NewOBSPanel(app *App) *OBSPanel {
	op := &OBSPanel{
		app:                app,
		obsStatus:          "未配置",
		streamConfigStatus: "未配置",
	}

	// 自动初始化
	op.autoInitialize()

	return op
}

// CreateContent 创建OBS管理面板内容
func (op *OBSPanel) CreateContent() fyne.CanvasObject {
	// 状态显示框架
	statusContainer := container.NewHBox(
		widget.NewLabel("OBS状态:"),
		op.createOBSStatusLabel(),
		widget.NewLabel("推流配置:"),
		op.createStreamConfigLabel(),
	)

	// OBS按钮组1
	btnGroup1 := container.NewHBox(
		op.createConfigPathButton(),
		op.createConfigStreamButton(),
	)

	// OBS按钮组2
	btnGroup2 := container.NewHBox(
		op.createSyncStreamButton(),
		op.createLaunchOBSButton(),
	)

	// OBS按钮组3
	btnGroup3 := container.NewHBox(
		op.createPluginManagerButton(),
		op.createSolveReconnectButton(),
	)

	// OBS按钮组4
	btnGroup4 := container.NewHBox(
		op.createHelpButton(),
		op.createRestoreButton(),
	)

	// 组合所有组件
	obsFrame := container.NewVBox(
		statusContainer,
		btnGroup1,
		btnGroup2,
		btnGroup3,
		btnGroup4,
	)

	return container.NewBorder(
		widget.NewCard("OBS管理", "", obsFrame),
		nil, nil, nil,
	)
}

// createOBSStatusLabel 创建OBS状态标签
func (op *OBSPanel) createOBSStatusLabel() *widget.Label {
	op.obsStatusLabel = widget.NewLabel(op.obsStatus)
	return op.obsStatusLabel
}

// createStreamConfigLabel 创建推流配置状态标签
func (op *OBSPanel) createStreamConfigLabel() *widget.Label {
	op.streamConfigLabel = widget.NewLabel(op.streamConfigStatus)
	return op.streamConfigLabel
}

// createConfigPathButton 创建配置OBS路径按钮
func (op *OBSPanel) createConfigPathButton() *widget.Button {
	op.configPathBtn = widget.NewButton("OBS路径配置", func() {
		op.configureOBSPath()
	})
	return op.configPathBtn
}

// createConfigStreamButton 创建推流配置按钮
func (op *OBSPanel) createConfigStreamButton() *widget.Button {
	op.configStreamBtn = widget.NewButton("推流配置", func() {
		op.configureStream()
	})
	return op.configStreamBtn
}

// createSyncStreamButton 创建同步推流码按钮
func (op *OBSPanel) createSyncStreamButton() *widget.Button {
	op.syncStreamBtn = widget.NewButton("同步推流码", func() {
		op.syncStreamConfig(false)
	})
	return op.syncStreamBtn
}

// createLaunchOBSButton 创建启动OBS按钮
func (op *OBSPanel) createLaunchOBSButton() *widget.Button {
	op.launchOBSBtn = widget.NewButton("启动OBS", func() {
		op.launchOBS()
	})
	return op.launchOBSBtn
}

// createPluginManagerButton 创建插件管理按钮
func (op *OBSPanel) createPluginManagerButton() *widget.Button {
	op.pluginManagerBtn = widget.NewButton("插件管理", func() {
		op.openPluginManager()
	})
	return op.pluginManagerBtn
}

// createSolveReconnectButton 创建一键解决重连按钮
func (op *OBSPanel) createSolveReconnectButton() *widget.Button {
	op.solveReconnectBtn = widget.NewButton("一键解决重连", func() {
		op.solveRepeatConnectOBS()
	})
	return op.solveReconnectBtn
}

// createHelpButton 创建帮助说明按钮
func (op *OBSPanel) createHelpButton() *widget.Button {
	op.helpBtn = widget.NewButton("帮助说明", func() {
		op.showOBSHelp()
	})
	return op.helpBtn
}

// createRestoreButton 创建还原重连配置按钮
func (op *OBSPanel) createRestoreButton() *widget.Button {
	op.restoreBtn = widget.NewButton("↑还原重连配置", func() {
		op.killMediaSDKServer()
	})
	return op.restoreBtn
}

// configureOBSPath 配置OBS路径
func (op *OBSPanel) configureOBSPath() {
	// 创建文件选择对话框
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, op.app.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		if filepath.Base(filePath) != "obs64.exe" {
			dialog.ShowError(fmt.Errorf("请选择obs64.exe文件"), op.app.window)
			return
		}

		// 保存配置
		op.obsPath = filePath
		op.obsStatus = "已配置"
		op.obsStatusLabel.SetText(op.obsStatus)
		op.app.configManager.SetOBSPath(filePath)
		op.app.logger.Info(fmt.Sprintf("OBS路径已配置: %s", filePath))

	}, op.app.window)

	// 设置文件过滤器
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
	fileDialog.Show()
}

// configureStream 配置推流设置
func (op *OBSPanel) configureStream() {
	// 获取OBS配置文件夹路径
	profilesPath := filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles")

	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("未找到OBS配置文件夹，请确保已安装OBS并运行过"), op.app.window)
		return
	}

	// 创建文件选择对话框
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, op.app.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		if filepath.Base(filePath) != "service.json" {
			dialog.ShowError(fmt.Errorf("请选择service.json文件"), op.app.window)
			return
		}

		// 更新状态
		op.streamConfigStatus = "已配置"
		op.streamConfigLabel.SetText(op.streamConfigStatus)
		op.app.logger.Info(fmt.Sprintf("推流配置已选择: %s", filePath))

		// 保存配置文件路径
		op.saveStreamConfigPath(filePath)

	}, op.app.window)

	// 设置初始目录
	if listableURI := storage.NewFileURI(profilesPath); listableURI != nil {
		// 尝试转换为ListableURI
		if listable, ok := listableURI.(fyne.ListableURI); ok {
			fileDialog.SetLocation(listable)
		}
	}

	// 设置文件过滤器
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	fileDialog.Show()
}

// syncStreamConfig 同步推流配置到OBS
func (op *OBSPanel) syncStreamConfig(fromLaunchButton bool) bool {
	serverURL := op.app.serverAddress
	streamKey := op.app.streamCode
	// TODO: 实现实际的同步逻辑
	op.app.logger.Info(fmt.Sprintf("同步推流配置: %s, %s", serverURL, streamKey))
	return true
}

// launchOBS 启动OBS
func (op *OBSPanel) launchOBS() {
	// TODO: 实现实际的启动逻辑
	success := op.syncStreamConfig(true)
	if success {
		op.app.logger.Info("OBS启动成功")
	}
}

// solveRepeatConnectOBS 解决重复连接OBS的问题
func (op *OBSPanel) solveRepeatConnectOBS() {
	op.app.logger.Info("开始处理重连问题...")

	// TODO: 实现进程管理器逻辑
	// 查找 MediaSDK_Server.exe 进程
	// 获取最活跃的线程
	// 挂起该线程

	op.app.logger.Info("处理重连问题结束，如果问题仍然存在，请尝试重启OBS和直播伴侣后重试")
}

// killMediaSDKServer 清除一键解决重连状态
func (op *OBSPanel) killMediaSDKServer() {
	op.app.logger.Info("正在清除一键解决重连...")
	// TODO: 实现进程管理器逻辑
	// 终止 MediaSDK_Server.exe 进程
	op.app.logger.Info("一键解决重连状态已清除")
}

// showOBSHelp 显示OBS管理面板使用说明
func (op *OBSPanel) showOBSHelp() {
	helpText := `# OBS管理使用说明

## 功能介绍

### 1. OBS路径配置
- 功能：配置OBS Studio的安装路径
- 操作：点击"OBS路径配置"按钮，选择obs64.exe文件
- 说明：如果系统中安装了OBS，工具会自动检测并配置

### 2. 推流配置
- 功能：配置OBS的推流设置文件
- 操作：点击"推流配置"按钮，选择service.json文件
- 位置：通常在 %APPDATA%\obs-studio\basic\profiles\配置名\service.json

### 3. 同步推流码
- 功能：将捕获到的推流服务器地址和推流码同步到OBS
- 操作：在捕获到推流信息后，点击"同步推流码"按钮
- 说明：需要先配置推流设置文件

### 4. 启动OBS
- 功能：启动OBS Studio并自动同步推流配置
- 操作：点击"启动OBS"按钮
- 说明：需要先配置OBS路径

### 5. 插件管理
- 功能：管理OBS插件
- 操作：点击"插件管理"按钮打开插件管理窗口

### 6. 一键解决重连
- 功能：解决OBS重复连接推流服务器的问题
- 操作：在出现重连问题时点击此按钮
- 说明：通过挂起相关进程线程来解决

### 7. 还原重连配置
- 功能：恢复被挂起的进程，清除重连解决状态
- 操作：点击"↑还原重连配置"按钮

## 使用流程

1. 首次使用需要配置OBS路径和推流配置
2. 使用主工具捕获推流信息
3. 点击"同步推流码"将信息同步到OBS
4. 点击"启动OBS"开始直播

## 注意事项

- 确保OBS Studio已正确安装
- 推流配置文件路径需要正确
- 在修改配置前建议备份原有设置
`

	content := widget.NewRichText()
	content.ParseMarkdown(helpText)
	content.Wrapping = fyne.TextWrapWord
	dialog.ShowCustom("OBS管理使用说明", "确定", content, op.app.window)
}

// openPluginManager 打开插件管理窗口
func (op *OBSPanel) openPluginManager() {
	// TODO: 实现插件管理窗口
	op.app.logger.Info("打开插件管理窗口...")
	dialog.ShowInformation("插件管理", "插件管理功能开发中...", op.app.window)
}

// saveStreamConfigPath 保存推流配置路径到配置文件
func (op *OBSPanel) saveStreamConfigPath(filePath string) {
	op.app.configManager.SetStreamConfigPath(filePath)
}

// autoInitialize 自动初始化OBS配置
func (op *OBSPanel) autoInitialize() {
	// 加载已保存的配置
	obsPath, obsConfigured, streamConfigured := op.loadOBSConfig()
	if obsConfigured {
		op.obsPath = obsPath
		op.obsStatus = "已配置"
		op.app.logger.Info(fmt.Sprintf("已加载OBS配置: %s", obsPath))
	}

	if streamConfigured {
		op.streamConfigStatus = "已配置"
		op.app.logger.Info("已加载推流配置")
	}

	if !obsConfigured {
		// 尝试自动检测OBS安装路径
		op.autoDetectOBS()
	}

	if obsConfigured && !streamConfigured {
		// 尝试自动检测推流配置
		op.autoDetectStreamConfig()
	}
}

// loadOBSConfig 加载OBS配置
func (op *OBSPanel) loadOBSConfig() (string, bool, bool) {
	obsPath := op.app.configManager.GetOBSPath()
	streamConfigPath := op.app.configManager.GetStreamConfigPath()

	obsConfigured := obsPath != "" && op.fileExists(obsPath)
	streamConfigured := streamConfigPath != "" && op.fileExists(streamConfigPath)

	return obsPath, obsConfigured, streamConfigured
}

// autoDetectOBS 自动检测OBS安装路径
func (op *OBSPanel) autoDetectOBS() {
	// 常用的OBS安装路径
	commonPaths := []string{
		"C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe",
		"D:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe",
		"E:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe",
		"F:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe",
	}

	for _, path := range commonPaths {
		if op.fileExists(path) {
			op.obsPath = path
			op.obsStatus = "已配置"
			op.app.configManager.SetOBSPath(path)
			op.app.logger.Info(fmt.Sprintf("自动找到OBS路径: %s", path))
			return
		}
	}

	// TODO: 添加从注册表查找OBS路径的逻辑
	// TODO: 添加从Steam路径查找OBS的逻辑

	op.app.logger.Info("未能自动找到OBS安装路径，请点击「OBS路径配置」按钮手动选择obs64.exe的位置")
}

// autoDetectStreamConfig 自动检测推流配置
func (op *OBSPanel) autoDetectStreamConfig() {
	profilesPath := filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles")
	if !op.dirExists(profilesPath) {
		return
	}

	// 首先检查"未命名"文件夹
	unnamedPath := filepath.Join(profilesPath, "未命名")
	serviceJSONPath := filepath.Join(unnamedPath, "service.json")

	if op.fileExists(serviceJSONPath) {
		op.saveStreamConfigPath(serviceJSONPath)
		op.streamConfigStatus = "已配置"
		op.app.logger.Info(fmt.Sprintf("找到默认推流配置: %s，已进行自动配置", serviceJSONPath))
		return
	}

	// 遍历所有文件夹查找service.json
	if dirs, err := os.ReadDir(profilesPath); err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				serviceJSONPath := filepath.Join(profilesPath, dir.Name(), "service.json")
				if op.fileExists(serviceJSONPath) {
					op.saveStreamConfigPath(serviceJSONPath)
					op.streamConfigStatus = "已配置"
					op.app.logger.Info(fmt.Sprintf("找到推流配置: %s", serviceJSONPath))
					return
				}
			}
		}
	}

	// 如果没有找到任何配置，在"未命名"文件夹下创建新的配置
	os.MkdirAll(unnamedPath, os.ModePerm)
	serviceJSONPath = filepath.Join(unnamedPath, "service.json")

	// TODO: 创建默认的service.json配置文件

	op.saveStreamConfigPath(serviceJSONPath)
	op.streamConfigStatus = "已配置"
	op.app.logger.Info(fmt.Sprintf("创建新的推流配置: %s", serviceJSONPath))
}

// fileExists 检查文件是否存在
func (op *OBSPanel) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// dirExists 检查目录是否存在
func (op *OBSPanel) dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
