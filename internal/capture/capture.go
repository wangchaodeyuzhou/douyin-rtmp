package capture

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"douyin-rtmp/internal/logger"
	"douyin-rtmp/internal/rtmp"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// CaptureStatus 捕获状态
type CaptureStatus int

const (
	StatusStopped CaptureStatus = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusError
)

// PacketInfo 数据包信息
type PacketInfo struct {
	Timestamp     time.Time `json:"timestamp"`
	SrcIP         string    `json:"src_ip"`
	DstIP         string    `json:"dst_ip"`
	SrcPort       int       `json:"src_port"`
	DstPort       int       `json:"dst_port"`
	Protocol      string    `json:"protocol"`
	Length        int       `json:"length"`
	PayloadSize   int       `json:"payload_size"`
	InterfaceName string    `json:"interface_name"`
}

// CaptureResult 捕获结果
type CaptureResult struct {
	RTMPInfo   *rtmp.RTMPInfo `json:"rtmp_info"`
	PacketInfo *PacketInfo    `json:"packet_info"`
	Success    bool           `json:"success"`
	Error      string         `json:"error,omitempty"`
}

// CaptureCallback 捕获回调函数
type CaptureCallback func(result *CaptureResult)

// PacketCapture 数据包捕获器
type PacketCapture struct {
	logger         *logger.Logger
	rtmpAnalyzer   *rtmp.RTMPPacketAnalyzer
	status         CaptureStatus
	callbacks      []CaptureCallback
	handles        map[string]*pcap.Handle
	contexts       map[string]context.Context
	cancelFuncs    map[string]context.CancelFunc
	mu             sync.RWMutex
	packetCount    int64
	captureTimeout time.Duration

	// 捕获到的RTMP信息
	serverAddress   string
	streamCode      string
	captureComplete bool
}

// NewPacketCapture 创建数据包捕获器
func NewPacketCapture(log *logger.Logger) *PacketCapture {
	return &PacketCapture{
		logger:         log,
		rtmpAnalyzer:   rtmp.NewRTMPPacketAnalyzer(),
		status:         StatusStopped,
		callbacks:      make([]CaptureCallback, 0),
		handles:        make(map[string]*pcap.Handle),
		contexts:       make(map[string]context.Context),
		cancelFuncs:    make(map[string]context.CancelFunc),
		captureTimeout: 30 * time.Second,
	}
}

// AddCallback 添加捕获回调
func (pc *PacketCapture) AddCallback(callback CaptureCallback) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.callbacks = append(pc.callbacks, callback)
}

// SetTimeout 设置捕获超时时间
func (pc *PacketCapture) SetTimeout(timeout time.Duration) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.captureTimeout = timeout
}

// GetStatus 获取捕获状态
func (pc *PacketCapture) GetStatus() CaptureStatus {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.status
}

// GetPacketCount 获取捕获的数据包数量
func (pc *PacketCapture) GetPacketCount() int64 {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.packetCount
}

// StartSingleInterface 开始单接口捕获
func (pc *PacketCapture) StartSingleInterface(interfaceName string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.status == StatusRunning {
		return fmt.Errorf("capture already running")
	}

	pc.resetCaptureState()
	pc.status = StatusStarting

	// 创建捕获上下文
	ctx, cancel := context.WithTimeout(context.Background(), pc.captureTimeout)
	pc.contexts[interfaceName] = ctx
	pc.cancelFuncs[interfaceName] = cancel

	// 打开网络接口
	handle, err := pc.openInterface(interfaceName)
	if err != nil {
		pc.status = StatusError
		cancel()
		return fmt.Errorf("failed to open interface %s: %w", interfaceName, err)
	}

	pc.handles[interfaceName] = handle
	pc.status = StatusRunning

	// 启动捕获goroutine
	go pc.captureLoop(ctx, interfaceName, handle)

	pc.logger.Info(fmt.Sprintf("开始在接口 %s 上捕获数据包", interfaceName))
	return nil
}

// StartMultiInterface 开始多接口捕获
func (pc *PacketCapture) StartMultiInterface(interfaceNames []string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.status == StatusRunning {
		return fmt.Errorf("capture already running")
	}

	if len(interfaceNames) == 0 {
		return fmt.Errorf("no interfaces specified")
	}

	pc.resetCaptureState()
	pc.status = StatusStarting

	// 创建全局超时上下文
	globalCtx, globalCancel := context.WithTimeout(context.Background(), pc.captureTimeout)
	defer func() {
		if pc.status != StatusRunning {
			globalCancel()
		}
	}()

	// 为每个接口创建捕获
	successCount := 0
	for _, interfaceName := range interfaceNames {
		// 为每个接口创建子上下文
		ctx, cancel := context.WithCancel(globalCtx)
		pc.contexts[interfaceName] = ctx
		pc.cancelFuncs[interfaceName] = cancel

		// 打开网络接口
		handle, err := pc.openInterface(interfaceName)
		if err != nil {
			pc.logger.Error(fmt.Sprintf("无法打开接口 %s: %v", interfaceName, err))
			cancel()
			continue
		}

		pc.handles[interfaceName] = handle
		successCount++

		// 启动捕获goroutine
		go pc.captureLoop(ctx, interfaceName, handle)
	}

	if successCount == 0 {
		pc.status = StatusError
		return fmt.Errorf("failed to open any interface")
	}

	pc.status = StatusRunning

	// 启动全局超时监控
	go func() {
		<-globalCtx.Done()
		if globalCtx.Err() == context.DeadlineExceeded {
			pc.logger.Warn("捕获超时，正在停止...")
			pc.Stop()
		}
	}()

	pc.logger.Info(fmt.Sprintf("开始在 %d 个接口上捕获数据包", successCount))
	return nil
}

// Stop 停止捕获
func (pc *PacketCapture) Stop() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.status != StatusRunning {
		return
	}

	pc.status = StatusStopping
	pc.logger.Info("正在停止数据包捕获...")

	// 取消所有上下文
	for interfaceName, cancel := range pc.cancelFuncs {
		if cancel != nil {
			cancel()
		}
		delete(pc.cancelFuncs, interfaceName)
		delete(pc.contexts, interfaceName)
	}

	// 关闭所有句柄
	for interfaceName, handle := range pc.handles {
		if handle != nil {
			handle.Close()
		}
		delete(pc.handles, interfaceName)
	}

	pc.status = StatusStopped
	pc.logger.Info("数据包捕获已停止")
}

// resetCaptureState 重置捕获状态
func (pc *PacketCapture) resetCaptureState() {
	pc.packetCount = 0
	pc.serverAddress = ""
	pc.streamCode = ""
	pc.captureComplete = false

	// 清理之前的资源
	for _, cancel := range pc.cancelFuncs {
		if cancel != nil {
			cancel()
		}
	}
	for _, handle := range pc.handles {
		if handle != nil {
			handle.Close()
		}
	}

	pc.handles = make(map[string]*pcap.Handle)
	pc.contexts = make(map[string]context.Context)
	pc.cancelFuncs = make(map[string]context.CancelFunc)
}

// openInterface 打开网络接口
func (pc *PacketCapture) openInterface(interfaceName string) (*pcap.Handle, error) {
	// 首先尝试用指定的名称直接打开
	handle, err := pcap.OpenLive(interfaceName, 65536, true, pcap.BlockForever)
	if err != nil {
		// 如果直接打开失败，尝试通过设备名称转换
		deviceName, findErr := pc.findDeviceByName(interfaceName)
		if findErr != nil {
			return nil, fmt.Errorf("无法找到接口 %s: %w (original error: %v)", interfaceName, findErr, err)
		}

		// 使用找到的设备名称重新尝试
		handle, err = pcap.OpenLive(deviceName, 65536, true, pcap.BlockForever)
		if err != nil {
			return nil, fmt.Errorf("无法打开接口 %s (设备名: %s): %w", interfaceName, deviceName, err)
		}
	}

	// 设置BPF过滤器，只捕获TCP数据包
	err = handle.SetBPFFilter("tcp")
	if err != nil {
		handle.Close()
		return nil, fmt.Errorf("failed to set BPF filter: %w", err)
	}

	return handle, nil
}

// findDeviceByName 通过接口名称查找对应的设备名称
func (pc *PacketCapture) findDeviceByName(interfaceName string) (string, error) {
	// 获取所有pcap设备
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return "", fmt.Errorf("获取设备列表失败: %w", err)
	}

	// 按优先级查找匹配的设备
	for _, device := range devices {
		// 1. 精确匹配设备名称
		if device.Name == interfaceName {
			return device.Name, nil
		}

		// 2. 匹配描述信息（如果有）
		if device.Description != "" {
			// 尝试匹配描述中的关键信息
			if strings.Contains(device.Description, interfaceName) {
				return device.Name, nil
			}

			// 去除特殊字符后匹配
			cleanInterfaceName := strings.ReplaceAll(interfaceName, "*", "")
			cleanInterfaceName = strings.TrimSpace(cleanInterfaceName)
			if cleanInterfaceName != "" && strings.Contains(device.Description, cleanInterfaceName) {
				return device.Name, nil
			}
		}
	}

	// 如果都没找到，返回错误
	return "", fmt.Errorf("未找到匹配的pcap设备: %s", interfaceName)
}

// captureLoop 捕获循环
func (pc *PacketCapture) captureLoop(ctx context.Context, interfaceName string, handle *pcap.Handle) {
	defer func() {
		if r := recover(); r != nil {
			pc.logger.Error(fmt.Sprintf("捕获循环异常: %v", r))
		}
	}()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-packetSource.Packets():
			if packet == nil {
				continue
			}

			pc.processPacket(packet, interfaceName)

			// 检查是否已经获取到完整的RTMP信息
			pc.mu.RLock()
			complete := pc.captureComplete
			pc.mu.RUnlock()

			if complete {
				pc.logger.Info("已获取完整的RTMP信息，停止捕获")
				go pc.Stop() // 异步停止，避免死锁
				return
			}
		}
	}
}

// processPacket 处理数据包
func (pc *PacketCapture) processPacket(packet gopacket.Packet, interfaceName string) {
	pc.mu.Lock()
	pc.packetCount++
	currentCount := pc.packetCount
	pc.mu.Unlock()

	// 解析网络层
	networkLayer := packet.NetworkLayer()
	if networkLayer == nil {
		return
	}

	ipv4Layer, ok := networkLayer.(*layers.IPv4)
	if !ok {
		return
	}

	// 解析传输层
	transportLayer := packet.TransportLayer()
	if transportLayer == nil {
		return
	}

	tcpLayer, ok := transportLayer.(*layers.TCP)
	if !ok {
		return
	}

	// 获取应用层数据
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer == nil {
		return
	}

	payload := applicationLayer.Payload()
	if len(payload) == 0 {
		return
	}

	// 创建数据包信息
	packetInfo := &PacketInfo{
		Timestamp:     packet.Metadata().Timestamp,
		SrcIP:         ipv4Layer.SrcIP.String(),
		DstIP:         ipv4Layer.DstIP.String(),
		SrcPort:       int(tcpLayer.SrcPort),
		DstPort:       int(tcpLayer.DstPort),
		Protocol:      "TCP",
		Length:        len(packet.Data()),
		PayloadSize:   len(payload),
		InterfaceName: interfaceName,
	}

	// 记录数据包信息
	if currentCount%100 == 0 { // 每100个包记录一次
		pc.logger.Packet(fmt.Sprintf("[%s] %s:%d -> %s:%d (长度: %d)",
			packetInfo.Timestamp.Format("15:04:05"),
			packetInfo.SrcIP, packetInfo.SrcPort,
			packetInfo.DstIP, packetInfo.DstPort,
			packetInfo.Length))
	}

	// 分析RTMP数据
	rtmpInfo := pc.rtmpAnalyzer.AnalyzePacket(payload)

	result := &CaptureResult{
		RTMPInfo:   rtmpInfo,
		PacketInfo: packetInfo,
		Success:    rtmpInfo != nil,
	}

	// 如果发现RTMP信息，更新状态
	if rtmpInfo != nil {
		pc.mu.Lock()

		if rtmpInfo.ServerAddress != "" && pc.serverAddress == "" {
			pc.serverAddress = rtmpInfo.ServerAddress
			pc.logger.Info(fmt.Sprintf("\n>>> 找到推流服务器地址 <<<\n地址: %s", pc.serverAddress))
		}

		if rtmpInfo.StreamCode != "" && pc.streamCode == "" {
			pc.streamCode = rtmpInfo.StreamCode
			pc.logger.Info(fmt.Sprintf("\n>>> 找到推流码 <<<\n推流码: %s", pc.streamCode))
		}

		// 检查是否已获取完整信息
		if pc.serverAddress != "" && pc.streamCode != "" && !pc.captureComplete {
			pc.captureComplete = true
			fullURL := pc.serverAddress + "/" + pc.streamCode
			pc.logger.Info(fmt.Sprintf("\n>>> 获取完整推流地址 <<<\n完整地址: %s", fullURL))

			// 更新结果中的完整信息
			result.RTMPInfo.ServerAddress = pc.serverAddress
			result.RTMPInfo.StreamCode = pc.streamCode
			result.RTMPInfo.FullURL = fullURL
		}
		pc.mu.Unlock()
	}

	// 调用回调函数
	pc.mu.RLock()
	callbacks := make([]CaptureCallback, len(pc.callbacks))
	copy(callbacks, pc.callbacks)
	pc.mu.RUnlock()

	for _, callback := range callbacks {
		go func(cb CaptureCallback) {
			defer func() {
				if r := recover(); r != nil {
					pc.logger.Error(fmt.Sprintf("回调函数异常: %v", r))
				}
			}()
			cb(result)
		}(callback)
	}
}

// GetCapturedInfo 获取已捕获的RTMP信息
func (pc *PacketCapture) GetCapturedInfo() (serverAddress, streamCode string, complete bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.serverAddress, pc.streamCode, pc.captureComplete
}

// IsRunning 检查是否正在运行
func (pc *PacketCapture) IsRunning() bool {
	return pc.GetStatus() == StatusRunning
}

// GetActiveInterfaces 获取当前活跃的接口列表
func (pc *PacketCapture) GetActiveInterfaces() []string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	interfaces := make([]string, 0, len(pc.handles))
	for interfaceName := range pc.handles {
		interfaces = append(interfaces, interfaceName)
	}
	return interfaces
}

// TestCapture 测试捕获功能
func (pc *PacketCapture) TestCapture(interfaceName string, duration time.Duration) (*CaptureStats, error) {
	stats := &CaptureStats{
		InterfaceName: interfaceName,
		StartTime:     time.Now(),
		Duration:      duration,
	}

	// 打开接口
	handle, err := pc.openInterface(interfaceName)
	if err != nil {
		stats.Error = err.Error()
		return stats, err
	}
	defer handle.Close()

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-ctx.Done():
			stats.EndTime = time.Now()
			stats.Success = stats.PacketCount > 0
			return stats, nil
		case packet := <-packetSource.Packets():
			if packet != nil {
				stats.PacketCount++
				if stats.PacketCount == 1 {
					stats.FirstPacketTime = packet.Metadata().Timestamp
				}
				stats.LastPacketTime = packet.Metadata().Timestamp
				stats.ByteCount += int64(len(packet.Data()))
			}
		}
	}
}

// CaptureStats 捕获统计信息
type CaptureStats struct {
	InterfaceName   string        `json:"interface_name"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	PacketCount     int64         `json:"packet_count"`
	ByteCount       int64         `json:"byte_count"`
	FirstPacketTime time.Time     `json:"first_packet_time"`
	LastPacketTime  time.Time     `json:"last_packet_time"`
	Success         bool          `json:"success"`
	Error           string        `json:"error,omitempty"`
}
