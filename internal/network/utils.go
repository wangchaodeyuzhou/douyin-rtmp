package network

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"
)

// GetLocalIPAddresses 获取本机IP地址列表
func GetLocalIPAddresses() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// GetDefaultGateway 获取默认网关
func GetDefaultGateway() (string, error) {
	// 这里简化实现，实际应该查询路由表
	// 在生产环境中建议使用更可靠的方法

	// 尝试连接到公网地址来确定默认路由
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIP := localAddr.IP.String()

	// 根据IP地址推断网关（这是一个简化的实现）
	parts := strings.Split(localIP, ".")
	if len(parts) == 4 {
		// 假设网关是 x.x.x.1
		gateway := fmt.Sprintf("%s.%s.%s.1", parts[0], parts[1], parts[2])
		return gateway, nil
	}

	return "", fmt.Errorf("unable to determine default gateway")
}

// PingHost 简单的ping测试
func PingHost(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "80"), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetMACAddress 获取指定接口的MAC地址
func GetMACAddress(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}

	return iface.HardwareAddr.String(), nil
}

// CheckPortAvailable 检查端口是否可用
func CheckPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// GetFreePort 获取一个可用的端口
func GetFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// ResolveHostname 解析主机名到IP地址
func ResolveHostname(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, ip := range ips {
		if ip.To4() != nil { // 只返回IPv4地址
			result = append(result, ip.String())
		}
	}

	return result, nil
}

// IsPrivateIP 判断是否为私有IP地址
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate()
}

// GetInterfaceSpeed 获取接口速度（简化实现）
func GetInterfaceSpeed(interfaceName string) (uint64, error) {
	// 在真实实现中，这需要调用系统API获取接口速度
	// 这里返回一个默认值
	switch runtime.GOOS {
	case "windows":
		// Windows下可以通过WMI查询
		return 1000, nil // 1000 Mbps 默认值
	case "linux":
		// Linux下可以读取 /sys/class/net/[interface]/speed
		return 1000, nil
	case "darwin":
		// macOS下可以使用system_profiler或ifconfig
		return 1000, nil
	default:
		return 1000, nil
	}
}

// NetworkDiagnostic 网络诊断信息
type NetworkDiagnostic struct {
	LocalIPs       []string          `json:"local_ips"`
	DefaultGateway string            `json:"default_gateway"`
	DNSServers     []string          `json:"dns_servers"`
	Connectivity   map[string]bool   `json:"connectivity"`
	Latency        map[string]string `json:"latency"`
}

// RunNetworkDiagnostic 运行网络诊断
func RunNetworkDiagnostic() (*NetworkDiagnostic, error) {
	diag := &NetworkDiagnostic{
		Connectivity: make(map[string]bool),
		Latency:      make(map[string]string),
	}

	// 获取本地IP
	if ips, err := GetLocalIPAddresses(); err == nil {
		diag.LocalIPs = ips
	}

	// 获取默认网关
	if gateway, err := GetDefaultGateway(); err == nil {
		diag.DefaultGateway = gateway
	}

	// 获取DNS服务器（简化实现）
	diag.DNSServers = []string{"8.8.8.8", "1.1.1.1"}

	// 测试连通性
	testHosts := []string{
		"8.8.8.8",   // Google DNS
		"1.1.1.1",   // Cloudflare DNS
		"baidu.com", // 百度
		"qq.com",    // 腾讯
	}

	for _, host := range testHosts {
		start := time.Now()
		connected := PingHost(host, 3*time.Second)
		duration := time.Since(start)

		diag.Connectivity[host] = connected
		if connected {
			diag.Latency[host] = duration.String()
		} else {
			diag.Latency[host] = "timeout"
		}
	}

	return diag, nil
}

// ValidateNetworkConfig 验证网络配置
func ValidateNetworkConfig(interfaceName string) error {
	// 检查接口是否存在
	if _, err := net.InterfaceByName(interfaceName); err != nil {
		return fmt.Errorf("interface %s not found: %w", interfaceName, err)
	}

	// 检查接口是否启用
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return err
	}

	if iface.Flags&net.FlagUp == 0 {
		return fmt.Errorf("interface %s is down", interfaceName)
	}

	// 检查是否有IP地址
	addrs, err := iface.Addrs()
	if err != nil {
		return fmt.Errorf("failed to get addresses for interface %s: %w", interfaceName, err)
	}

	hasIPv4 := false
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}
	}

	if !hasIPv4 {
		return fmt.Errorf("interface %s has no IPv4 address", interfaceName)
	}

	return nil
}

// GetInterfaceByIP 根据IP地址查找接口
func GetInterfaceByIP(targetIP string) (*net.Interface, error) {
	target := net.ParseIP(targetIP)
	if target == nil {
		return nil, fmt.Errorf("invalid IP address: %s", targetIP)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.Equal(target) {
					return &iface, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no interface found for IP %s", targetIP)
}

// MonitorNetworkChanges 网络变化监控（简化版本）
type NetworkChangeCallback func(change string)

// StartNetworkMonitor 开始网络监控（占位符实现）
func StartNetworkMonitor(callback NetworkChangeCallback) error {
	// 在实际实现中，这里应该使用系统特定的API来监控网络变化
	// 例如 Windows 的 WMI 事件，Linux 的 netlink，macOS 的 SystemConfiguration
	// 这里提供一个简化的轮询实现

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		lastInterfaces := make(map[string]bool)

		for range ticker.C {
			currentInterfaces := make(map[string]bool)

			interfaces, err := net.Interfaces()
			if err != nil {
				continue
			}

			for _, iface := range interfaces {
				if iface.Flags&net.FlagUp != 0 {
					currentInterfaces[iface.Name] = true
				}
			}

			// 检查新增的接口
			for name := range currentInterfaces {
				if !lastInterfaces[name] {
					callback(fmt.Sprintf("Interface %s is now up", name))
				}
			}

			// 检查移除的接口
			for name := range lastInterfaces {
				if !currentInterfaces[name] {
					callback(fmt.Sprintf("Interface %s is now down", name))
				}
			}

			lastInterfaces = currentInterfaces
		}
	}()

	return nil
}
