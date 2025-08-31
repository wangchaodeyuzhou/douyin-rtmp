package network

import (
	"douyin-rtmp/internal/logger"
	"strings"
)

// NetworkInterface 网络接口管理器
type NetworkInterface struct {
	logger *logger.Logger
}

// NewNetworkInterface 创建网络接口管理器
func NewNetworkInterface(log *logger.Logger) *NetworkInterface {
	return &NetworkInterface{
		logger: log,
	}
}

// LoadInterfaces 加载网络接口列表
func (ni *NetworkInterface) LoadInterfaces() map[string]interface{} {
	// TODO: 实现网络接口枚举逻辑
	// 这里应该使用gopacket或其他方式获取网络接口列表

	// 临时返回模拟数据
	interfaces := []string{
		"以太网 [已连接] - Realtek PCIe GBE Family Controller",
		"WiFi [已连接] - Intel Wireless-AC 9560 160MHz",
		"蓝牙 [未连接] - Bluetooth Device",
	}

	return map[string]interface{}{
		"interfaces":   interfaces,
		"default":      interfaces[0], // 默认选择第一个
		"active_count": 2,
	}
}

// ParseDisplayName 从显示名称中解析实际接口名
func (ni *NetworkInterface) ParseDisplayName(displayName string) string {
	// 从显示名称中提取实际的接口名称
	parts := strings.Split(displayName, " [")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return displayName
}

// IsValidInterface 验证接口是否有效
func (ni *NetworkInterface) IsValidInterface(interfaceName string) bool {
	// TODO: 实现实际的接口验证逻辑
	// 临时返回true
	return interfaceName != ""
}

// RefreshInterfaces 刷新接口列表
func (ni *NetworkInterface) RefreshInterfaces() error {
	// TODO: 实现刷新逻辑
	ni.logger.Info("刷新网络接口列表")
	return nil
}

// GetDisplayNames 获取所有显示名称
func (ni *NetworkInterface) GetDisplayNames() []string {
	result := ni.LoadInterfaces()
	return result["interfaces"].([]string)
}

// GetActiveDisplayNames 获取活跃接口显示名称
func (ni *NetworkInterface) GetActiveDisplayNames() []string {
	// TODO: 实现活跃接口过滤
	return ni.GetDisplayNames()
}

// GetDefaultInterface 获取默认接口
func (ni *NetworkInterface) GetDefaultInterface() (NetworkInterfaceInfo, bool) {
	result := ni.LoadInterfaces()
	if defaultIface, ok := result["default"].(string); ok && defaultIface != "" {
		return NetworkInterfaceInfo{
			Name:        ni.ParseDisplayName(defaultIface),
			DisplayName: defaultIface,
		}, true
	}
	return NetworkInterfaceInfo{}, false
}

// GetSystemInfo 获取系统信息
func (ni *NetworkInterface) GetSystemInfo() (map[string]interface{}, error) {
	// TODO: 实现系统信息收集
	return map[string]interface{}{
		"hostname":        "localhost",
		"os":              "windows",
		"platform":        "x64",
		"interface_stats": 3,
	}, nil
}

// NetworkInterfaceInfo 网络接口信息
type NetworkInterfaceInfo struct {
	Name        string
	DisplayName string
}

// GetActiveInterfaces 获取活跃接口列表
func (ni *NetworkInterface) GetActiveInterfaces() []NetworkInterfaceInfo {
	// TODO: 实现活跃接口枚举
	displayNames := ni.GetActiveDisplayNames()
	interfaces := make([]NetworkInterfaceInfo, len(displayNames))

	for i, displayName := range displayNames {
		interfaces[i] = NetworkInterfaceInfo{
			Name:        ni.ParseDisplayName(displayName),
			DisplayName: displayName,
		}
	}

	return interfaces
}

// TestConnectivity 测试连通性
func (ni *NetworkInterface) TestConnectivity(interfaceName string) bool {
	// TODO: 实现实际的连通性测试
	// 临时返回true表示连通
	return true
}
