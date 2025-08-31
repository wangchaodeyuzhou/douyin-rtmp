package capture

import (
	"context"
	"fmt"
	"time"

	"douyin-rtmp/internal/logger"
	"douyin-rtmp/internal/network"
)

// CaptureManager 捕获管理器接口
type CaptureManager interface {
	// 开始捕获
	StartCapture(interfaces []string) error

	// 停止捕获
	StopCapture() error

	// 获取状态
	GetStatus() CaptureStatus

	// 添加结果回调
	AddResultCallback(callback CaptureCallback)

	// 获取统计信息
	GetStats() *CaptureManagerStats

	// 刷新接口列表
	RefreshInterfaces() error

	// 获取可用接口列表
	GetAvailableInterfaces() []string

	// 获取活跃接口列表
	GetActiveInterfaces() []string

	// 获取默认接口
	GetDefaultInterface() (string, bool)

	// 测试接口
	TestInterface(interfaceDisplayName string, duration time.Duration) (*CaptureStats, error)

	// 获取网络诊断信息
	GetNetworkDiagnostic() (map[string]interface{}, error)

	// 执行连通性测试
	PerformConnectivityTest() map[string]bool
}

// CaptureManagerImpl 捕获管理器实现
type CaptureManagerImpl struct {
	logger        *logger.Logger
	netInterface  *network.NetworkInterface
	packetCapture *PacketCapture

	// 管理器状态
	isInitialized bool
	currentMode   CaptureMode

	// 统计信息
	stats *CaptureManagerStats
}

// CaptureMode 捕获模式
type CaptureMode int

const (
	ModeSingle CaptureMode = iota
	ModeMulti
	ModeAll
)

// CaptureManagerStats 捕获管理器统计信息
type CaptureManagerStats struct {
	TotalPackets     int64         `json:"total_packets"`
	ActiveInterfaces []string      `json:"active_interfaces"`
	StartTime        time.Time     `json:"start_time"`
	Duration         time.Duration `json:"duration"`
	RTMPFound        bool          `json:"rtmp_found"`
	ServerAddress    string        `json:"server_address"`
	StreamCode       string        `json:"stream_code"`
	LastUpdate       time.Time     `json:"last_update"`
}

// NewCaptureManager 创建捕获管理器
func NewCaptureManager(log *logger.Logger) CaptureManager {
	manager := &CaptureManagerImpl{
		logger:        log,
		netInterface:  network.NewNetworkInterface(log),
		packetCapture: NewPacketCapture(log),
		stats: &CaptureManagerStats{
			LastUpdate: time.Now(),
		},
	}

	// 初始化网络接口
	interfaces := manager.netInterface.LoadInterfaces()
	if len(interfaces["interfaces"].([]string)) > 0 {
		manager.isInitialized = true
		log.Info("网络接口初始化成功")
	} else {
		log.Error("网络接口初始化失败")
	}

	// 添加捕获回调
	manager.packetCapture.AddCallback(manager.handleCaptureResult)

	return manager
}

// StartCapture 开始捕获
func (cm *CaptureManagerImpl) StartCapture(interfaces []string) error {
	if !cm.isInitialized {
		return fmt.Errorf("capture manager not initialized")
	}

	if cm.GetStatus() == StatusRunning {
		return fmt.Errorf("capture already running")
	}

	// 验证接口
	validInterfaces, err := cm.validateInterfaces(interfaces)
	if err != nil {
		return err
	}

	if len(validInterfaces) == 0 {
		return fmt.Errorf("no valid interfaces to capture")
	}

	// 重置统计信息
	cm.resetStats()
	cm.stats.StartTime = time.Now()
	cm.stats.ActiveInterfaces = validInterfaces

	// 根据接口数量选择捕获模式
	if len(validInterfaces) == 1 {
		cm.currentMode = ModeSingle
		err = cm.packetCapture.StartSingleInterface(validInterfaces[0])
	} else {
		cm.currentMode = ModeMulti
		err = cm.packetCapture.StartMultiInterface(validInterfaces)
	}

	if err != nil {
		cm.logger.Error(fmt.Sprintf("启动捕获失败: %v", err))
		return err
	}

	cm.logger.Info(fmt.Sprintf("成功启动捕获，监听 %d 个接口", len(validInterfaces)))
	return nil
}

// StopCapture 停止捕获
func (cm *CaptureManagerImpl) StopCapture() error {
	if cm.GetStatus() != StatusRunning {
		return fmt.Errorf("capture not running")
	}

	cm.packetCapture.Stop()

	// 更新统计信息
	cm.stats.Duration = time.Since(cm.stats.StartTime)
	cm.stats.LastUpdate = time.Now()

	cm.logger.Info("捕获已停止")
	return nil
}

// GetStatus 获取状态
func (cm *CaptureManagerImpl) GetStatus() CaptureStatus {
	return cm.packetCapture.GetStatus()
}

// AddResultCallback 添加结果回调
func (cm *CaptureManagerImpl) AddResultCallback(callback CaptureCallback) {
	cm.packetCapture.AddCallback(callback)
}

// GetStats 获取统计信息
func (cm *CaptureManagerImpl) GetStats() *CaptureManagerStats {
	// 更新实时数据
	cm.stats.TotalPackets = cm.packetCapture.GetPacketCount()
	cm.stats.LastUpdate = time.Now()

	if cm.GetStatus() == StatusRunning {
		cm.stats.Duration = time.Since(cm.stats.StartTime)
	}

	// 获取RTMP信息
	serverAddr, streamCode, complete := cm.packetCapture.GetCapturedInfo()
	cm.stats.RTMPFound = complete
	cm.stats.ServerAddress = serverAddr
	cm.stats.StreamCode = streamCode

	return cm.stats
}

// validateInterfaces 验证接口列表
func (cm *CaptureManagerImpl) validateInterfaces(interfaces []string) ([]string, error) {
	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no interfaces specified")
	}

	validInterfaces := make([]string, 0)

	for _, interfaceDisplayName := range interfaces {
		// 从显示名称解析实际接口名
		actualName := cm.netInterface.ParseDisplayName(interfaceDisplayName)

		// 验证接口是否有效
		if cm.netInterface.IsValidInterface(actualName) {
			validInterfaces = append(validInterfaces, actualName)
			cm.logger.Debug(fmt.Sprintf("验证接口成功: %s -> %s", interfaceDisplayName, actualName))
		} else {
			cm.logger.Warn(fmt.Sprintf("无效的接口: %s", interfaceDisplayName))
		}
	}

	return validInterfaces, nil
}

// handleCaptureResult 处理捕获结果
func (cm *CaptureManagerImpl) handleCaptureResult(result *CaptureResult) {
	if result == nil {
		return
	}

	// 更新统计信息
	cm.stats.LastUpdate = time.Now()

	if result.Success && result.RTMPInfo != nil {
		rtmpInfo := result.RTMPInfo

		if rtmpInfo.ServerAddress != "" {
			cm.stats.ServerAddress = rtmpInfo.ServerAddress
		}

		if rtmpInfo.StreamCode != "" {
			cm.stats.StreamCode = rtmpInfo.StreamCode
		}

		// 检查是否获取到完整信息
		if rtmpInfo.FullURL != "" {
			cm.stats.RTMPFound = true
			cm.logger.Info("RTMP推流地址获取完成")
		}
	}
}

// resetStats 重置统计信息
func (cm *CaptureManagerImpl) resetStats() {
	cm.stats = &CaptureManagerStats{
		TotalPackets:     0,
		ActiveInterfaces: make([]string, 0),
		RTMPFound:        false,
		ServerAddress:    "",
		StreamCode:       "",
		LastUpdate:       time.Now(),
	}
}

// GetAvailableInterfaces 获取可用接口列表
func (cm *CaptureManagerImpl) GetAvailableInterfaces() []string {
	if !cm.isInitialized {
		return nil
	}

	// 刷新接口列表
	if err := cm.netInterface.RefreshInterfaces(); err != nil {
		cm.logger.Error(fmt.Sprintf("刷新接口列表失败: %v", err))
		return nil
	}

	return cm.netInterface.GetDisplayNames()
}

// GetActiveInterfaces 获取活跃接口列表
func (cm *CaptureManagerImpl) GetActiveInterfaces() []string {
	if !cm.isInitialized {
		return nil
	}

	return cm.netInterface.GetActiveDisplayNames()
}

// GetDefaultInterface 获取默认接口
func (cm *CaptureManagerImpl) GetDefaultInterface() (string, bool) {
	if !cm.isInitialized {
		return "", false
	}

	if iface, exists := cm.netInterface.GetDefaultInterface(); exists {
		return iface.DisplayName, true
	}

	return "", false
}

// TestInterface 测试接口
func (cm *CaptureManagerImpl) TestInterface(interfaceDisplayName string, duration time.Duration) (*CaptureStats, error) {
	if !cm.isInitialized {
		return nil, fmt.Errorf("capture manager not initialized")
	}

	// 解析接口名称
	actualName := cm.netInterface.ParseDisplayName(interfaceDisplayName)

	// 验证接口
	if !cm.netInterface.IsValidInterface(actualName) {
		return nil, fmt.Errorf("invalid interface: %s", interfaceDisplayName)
	}

	cm.logger.Info(fmt.Sprintf("开始测试接口 %s，持续时间 %v", interfaceDisplayName, duration))

	// 执行测试
	stats, err := cm.packetCapture.TestCapture(actualName, duration)
	if err != nil {
		cm.logger.Error(fmt.Sprintf("接口测试失败: %v", err))
		return stats, err
	}

	cm.logger.Info(fmt.Sprintf("接口测试完成，捕获 %d 个数据包，%d 字节", stats.PacketCount, stats.ByteCount))
	return stats, nil
}

// RefreshInterfaces 刷新接口列表
func (cm *CaptureManagerImpl) RefreshInterfaces() error {
	if !cm.isInitialized {
		return fmt.Errorf("capture manager not initialized")
	}

	return cm.netInterface.RefreshInterfaces()
}

// GetNetworkDiagnostic 获取网络诊断信息
func (cm *CaptureManagerImpl) GetNetworkDiagnostic() (map[string]interface{}, error) {
	if !cm.isInitialized {
		return nil, fmt.Errorf("capture manager not initialized")
	}

	return cm.netInterface.GetSystemInfo()
}

// SetCaptureTimeout 设置捕获超时
func (cm *CaptureManagerImpl) SetCaptureTimeout(timeout time.Duration) {
	cm.packetCapture.SetTimeout(timeout)
}

// PerformConnectivityTest 执行连通性测试
func (cm *CaptureManagerImpl) PerformConnectivityTest() map[string]bool {
	if !cm.isInitialized {
		return nil
	}

	result := make(map[string]bool)
	interfaces := cm.netInterface.GetActiveInterfaces()

	for _, iface := range interfaces {
		connected := cm.netInterface.TestConnectivity(iface.Name)
		result[iface.DisplayName] = connected
	}

	return result
}

// CaptureTaskConfig 捕获任务配置
type CaptureTaskConfig struct {
	Interfaces  []string      `json:"interfaces"`
	Timeout     time.Duration `json:"timeout"`
	AutoStop    bool          `json:"auto_stop"`    // 获取到完整信息后自动停止
	PacketLimit int64         `json:"packet_limit"` // 数据包限制
	MaxDuration time.Duration `json:"max_duration"` // 最大持续时间
}

// StartCaptureTask 启动捕获任务
func (cm *CaptureManagerImpl) StartCaptureTask(config *CaptureTaskConfig) error {
	if config == nil {
		return fmt.Errorf("capture task config is nil")
	}

	// 设置超时
	if config.Timeout > 0 {
		cm.SetCaptureTimeout(config.Timeout)
	}

	// 启动捕获
	err := cm.StartCapture(config.Interfaces)
	if err != nil {
		return err
	}

	// 如果设置了最大持续时间，启动定时器
	if config.MaxDuration > 0 {
		go func() {
			timer := time.NewTimer(config.MaxDuration)
			defer timer.Stop()

			select {
			case <-timer.C:
				cm.logger.Info("达到最大捕获时间，自动停止")
				cm.StopCapture()
			}
		}()
	}

	// 如果设置了数据包限制，启动监控
	if config.PacketLimit > 0 {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				if cm.GetStatus() != StatusRunning {
					return
				}

				if cm.packetCapture.GetPacketCount() >= config.PacketLimit {
					cm.logger.Info("达到数据包限制，自动停止")
					cm.StopCapture()
					return
				}
			}
		}()
	}

	return nil
}

// WaitForCompletion 等待捕获完成
func (cm *CaptureManagerImpl) WaitForCompletion(ctx context.Context) (*CaptureManagerStats, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return cm.GetStats(), ctx.Err()
		case <-ticker.C:
			status := cm.GetStatus()
			if status == StatusStopped || status == StatusError {
				return cm.GetStats(), nil
			}

			// 检查是否获取到完整信息
			stats := cm.GetStats()
			if stats.RTMPFound {
				return stats, nil
			}
		}
	}
}
