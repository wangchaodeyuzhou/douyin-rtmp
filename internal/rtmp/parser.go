package rtmp

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RTMPPacketAnalyzer RTMP数据包分析器
type RTMPPacketAnalyzer struct {
	extractor *RTMPExtractor
}

// NewRTMPPacketAnalyzer 创建RTMP数据包分析器
func NewRTMPPacketAnalyzer() *RTMPPacketAnalyzer {
	return &RTMPPacketAnalyzer{
		extractor: NewRTMPExtractor(),
	}
}

// AnalyzePacket 分析数据包
func (rpa *RTMPPacketAnalyzer) AnalyzePacket(payload []byte) *RTMPInfo {
	return rpa.extractor.ExtractFromPayload(payload)
}

// RTMPInfo RTMP推流信息
type RTMPInfo struct {
	ServerAddress string    `json:"server_address"`
	StreamCode    string    `json:"stream_code"`
	FullURL       string    `json:"full_url"`
	DetectedAt    time.Time `json:"detected_at"`
	Protocol      string    `json:"protocol"`
	Port          int       `json:"port"`
}

// RTMPExtractor RTMP地址提取器
type RTMPExtractor struct {
	serverRegexes []regexp.Regexp
	streamRegexes []regexp.Regexp
}

// NewRTMPExtractor 创建RTMP提取器
func NewRTMPExtractor() *RTMPExtractor {
	extractor := &RTMPExtractor{}
	extractor.initRegexes()
	return extractor
}

// initRegexes 初始化正则表达式
func (re *RTMPExtractor) initRegexes() {
	// 服务器地址匹配正则表达式
	serverPatterns := []string{
		`(rtmp://[a-zA-Z0-9\-\.]+/[^/\s\x00]+)`,          // 标准RTMP地址
		`(rtmps://[a-zA-Z0-9\-\.]+/[^/\s\x00]+)`,         // RTMPS地址
		`rtmp://([a-zA-Z0-9\-\.]+):?(\d*)/([^/\s\x00]+)`, // 带端口的RTMP地址
		`(rtmp://[^/\s\x00]+/live)`,                      // 抖音live路径
		`(rtmp://[^/\s\x00]+/rtmplive)`,                  // 抖音rtmplive路径
	}

	// 推流码匹配正则表达式
	streamPatterns := []string{
		`(stream-\d+\?[a-zA-Z0-9_]+=[a-zA-Z0-9\-]+(?:&[a-zA-Z0-9_]+=[a-zA-Z0-9\-]+)*)`, // 抖音推流码格式
		`([a-zA-Z0-9_\-]{20,})`,   // 长字符串推流码
		`(live_\d+_\d+)`,          // live格式推流码
		`([a-f0-9]{32})`,          // MD5格式推流码
		`(stream[a-zA-Z0-9_\-]+)`, // stream开头的推流码
	}

	// 编译正则表达式
	for _, pattern := range serverPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			re.serverRegexes = append(re.serverRegexes, *regex)
		}
	}

	for _, pattern := range streamPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			re.streamRegexes = append(re.streamRegexes, *regex)
		}
	}
}

// ExtractFromPayload 从数据包载荷中提取RTMP信息
func (re *RTMPExtractor) ExtractFromPayload(payload []byte) *RTMPInfo {
	// 尝试将载荷转换为字符串
	payloadStr := string(payload)

	// 清理字符串，移除null字符和不可见字符
	payloadStr = strings.ReplaceAll(payloadStr, "\x00", "")
	payloadStr = strings.TrimSpace(payloadStr)

	info := &RTMPInfo{
		DetectedAt: time.Now(),
		Protocol:   "rtmp",
		Port:       1935, // 默认RTMP端口
	}

	// 检测服务器地址
	serverAddress := re.extractServerAddress(payloadStr)
	if serverAddress != "" {
		info.ServerAddress = serverAddress

		// 从服务器地址中提取端口信息
		if port := re.extractPortFromAddress(serverAddress); port != 0 {
			info.Port = port
		}

		// 检测协议
		if strings.HasPrefix(serverAddress, "rtmps://") {
			info.Protocol = "rtmps"
		}
	}

	// 检测推流码
	streamCode := re.extractStreamCode(payloadStr)
	if streamCode != "" {
		// 清理推流码
		streamCode = re.cleanStreamCode(streamCode)
		info.StreamCode = streamCode
	}

	// 构建完整URL
	if info.ServerAddress != "" && info.StreamCode != "" {
		info.FullURL = info.ServerAddress + "/" + info.StreamCode
		return info
	}

	return nil
}

// extractServerAddress 提取服务器地址
func (re *RTMPExtractor) extractServerAddress(payload string) string {
	for _, regex := range re.serverRegexes {
		if matches := regex.FindStringSubmatch(payload); len(matches) > 1 {
			address := matches[1]
			// 清理地址，移除尾部的null字符
			if nullIndex := strings.Index(address, "\x00"); nullIndex != -1 {
				address = address[:nullIndex]
			}
			return address
		}
	}

	// 特殊处理：查找连接命令中的地址
	if strings.Contains(payload, "connect") {
		return re.extractConnectAddress(payload)
	}

	return ""
}

// extractConnectAddress 从connect命令中提取地址
func (re *RTMPExtractor) extractConnectAddress(payload string) string {
	// 查找connect命令后的URL
	connectIndex := strings.Index(payload, "connect")
	if connectIndex == -1 {
		return ""
	}

	// 在connect后查找rtmp://开头的地址
	remaining := payload[connectIndex:]
	rtmpIndex := strings.Index(remaining, "rtmp://")
	if rtmpIndex == -1 {
		return ""
	}

	// 提取URL直到空格或null字符
	urlStart := rtmpIndex
	urlEnd := len(remaining)

	for i := urlStart; i < len(remaining); i++ {
		char := remaining[i]
		if char == ' ' || char == '\x00' || char == '\n' || char == '\r' {
			urlEnd = i
			break
		}
	}

	if urlEnd > urlStart {
		return remaining[urlStart:urlEnd]
	}

	return ""
}

// extractStreamCode 提取推流码
func (re *RTMPExtractor) extractStreamCode(payload string) string {
	// 特殊处理FCPublish命令
	if strings.Contains(payload, "FCPublish") {
		return re.extractFCPublishStreamCode(payload)
	}

	// 特殊处理publish命令
	if strings.Contains(payload, "publish") {
		return re.extractPublishStreamCode(payload)
	}

	// 使用正则表达式匹配
	for _, regex := range re.streamRegexes {
		if matches := regex.FindStringSubmatch(payload); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractFCPublishStreamCode 从FCPublish命令中提取推流码
func (re *RTMPExtractor) extractFCPublishStreamCode(payload string) string {
	fcIndex := strings.Index(payload, "FCPublish")
	if fcIndex == -1 {
		return ""
	}

	// FCPublish后通常紧跟推流码
	remaining := payload[fcIndex+9:] // "FCPublish"长度为9

	// 查找推流码模式
	streamRegex := regexp.MustCompile(`(stream-\d+\?[a-zA-Z0-9_]+=[a-zA-Z0-9\-]+(?:&[a-zA-Z0-9_]+=[a-zA-Z0-9\-]+)*)`)
	if matches := streamRegex.FindStringSubmatch(remaining); len(matches) > 1 {
		return matches[1]
	}

	// 如果没找到，尝试提取任何长字符串
	parts := strings.Fields(remaining)
	for _, part := range parts {
		if len(part) > 10 && !strings.Contains(part, " ") {
			return part
		}
	}

	return ""
}

// extractPublishStreamCode 从publish命令中提取推流码
func (re *RTMPExtractor) extractPublishStreamCode(payload string) string {
	publishIndex := strings.Index(payload, "publish")
	if publishIndex == -1 {
		return ""
	}

	// publish后通常紧跟推流码
	remaining := payload[publishIndex+7:] // "publish"长度为7

	// 查找推流码
	parts := strings.Fields(remaining)
	for _, part := range parts {
		if len(part) > 8 && re.isValidStreamCode(part) {
			return part
		}
	}

	return ""
}

// isValidStreamCode 检查是否为有效的推流码
func (re *RTMPExtractor) isValidStreamCode(code string) bool {
	// 推流码通常包含字母、数字、下划线、连字符
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_\-\?&=]+$`)
	if !validChars.MatchString(code) {
		return false
	}

	// 推流码长度通常在10-200字符之间
	if len(code) < 10 || len(code) > 200 {
		return false
	}

	// 排除一些明显不是推流码的字符串
	excludePatterns := []string{
		"connect", "publish", "FCPublish", "createStream",
		"onStatus", "NetConnection", "NetStream",
	}

	lowerCode := strings.ToLower(code)
	for _, pattern := range excludePatterns {
		if strings.Contains(lowerCode, strings.ToLower(pattern)) {
			return false
		}
	}

	return true
}

// cleanStreamCode 清理推流码
func (re *RTMPExtractor) cleanStreamCode(code string) string {
	// 移除末尾的特殊字符
	code = strings.TrimRight(code, "C\x00\r\n\t ")

	// 移除开头的特殊字符
	code = strings.TrimLeft(code, "\x00\r\n\t ")

	return code
}

// extractPortFromAddress 从地址中提取端口号
func (re *RTMPExtractor) extractPortFromAddress(address string) int {
	// 匹配端口号
	portRegex := regexp.MustCompile(`:(\d+)/`)
	if matches := portRegex.FindStringSubmatch(address); len(matches) > 1 {
		if port := parseInt(matches[1]); port > 0 && port <= 65535 {
			return port
		}
	}
	return 0
}

// parseInt 字符串转整数
func parseInt(str string) int {
	result := 0
	for _, char := range str {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0
		}
	}
	return result
}

// AnalyzePacketAdvanced 高级数据包分析（带IP和端口信息）
func (rpa *RTMPPacketAnalyzer) AnalyzePacketAdvanced(srcIP, dstIP string, srcPort, dstPort int, payload []byte) *RTMPInfo {
	// 检查是否为RTMP端口
	if !rpa.isRTMPPort(srcPort) && !rpa.isRTMPPort(dstPort) {
		return nil
	}

	// 检查载荷是否包含RTMP相关关键词
	if !rpa.containsRTMPKeywords(payload) {
		return nil
	}

	// 提取RTMP信息
	return rpa.extractor.ExtractFromPayload(payload)
}

// isRTMPPort 检查是否为RTMP相关端口
func (rpa *RTMPPacketAnalyzer) isRTMPPort(port int) bool {
	rtmpPorts := []int{1935, 1936, 19350, 8080, 80, 443}
	for _, rtmpPort := range rtmpPorts {
		if port == rtmpPort {
			return true
		}
	}
	return false
}

// containsRTMPKeywords 检查载荷是否包含RTMP关键词
func (rpa *RTMPPacketAnalyzer) containsRTMPKeywords(payload []byte) bool {
	payloadStr := strings.ToLower(string(payload))

	keywords := []string{
		"rtmp://", "rtmps://", "connect", "publish", "fcpublish",
		"createstream", "netstram", "netconnection", "stream-",
	}

	for _, keyword := range keywords {
		if strings.Contains(payloadStr, keyword) {
			return true
		}
	}

	return false
}

// RTMPChunkParser RTMP块解析器
type RTMPChunkParser struct{}

// NewRTMPChunkParser 创建RTMP块解析器
func NewRTMPChunkParser() *RTMPChunkParser {
	return &RTMPChunkParser{}
}

// ParseRTMPChunk 解析RTMP块
func (parser *RTMPChunkParser) ParseRTMPChunk(data []byte) map[string]interface{} {
	result := make(map[string]interface{})

	if len(data) < 12 {
		return result
	}

	// 基本RTMP头解析
	chunkStreamID := data[0] & 0x3F
	fmt := (data[0] >> 6) & 0x03

	result["chunk_stream_id"] = chunkStreamID
	result["format"] = fmt

	// 根据格式解析时间戳和消息信息
	offset := 1
	if fmt <= 2 {
		// 包含时间戳
		if len(data) >= offset+3 {
			timestamp := binary.BigEndian.Uint32(append([]byte{0}, data[offset:offset+3]...))
			result["timestamp"] = timestamp
			offset += 3
		}
	}

	if fmt <= 1 {
		// 包含消息长度和类型
		if len(data) >= offset+4 {
			msgLength := binary.BigEndian.Uint32(append([]byte{0}, data[offset:offset+3]...))
			msgType := data[offset+3]
			result["message_length"] = msgLength
			result["message_type"] = msgType
			offset += 4
		}
	}

	if fmt == 0 {
		// 包含消息流ID
		if len(data) >= offset+4 {
			msgStreamID := binary.LittleEndian.Uint32(data[offset : offset+4])
			result["message_stream_id"] = msgStreamID
			offset += 4
		}
	}

	// 提取载荷
	if len(data) > offset {
		result["payload"] = data[offset:]
	}

	return result
}

// ValidateRTMPInfo 验证RTMP信息的完整性
func ValidateRTMPInfo(info *RTMPInfo) error {
	if info == nil {
		return fmt.Errorf("RTMP info is nil")
	}

	if info.ServerAddress == "" {
		return fmt.Errorf("server address is empty")
	}

	if info.StreamCode == "" {
		return fmt.Errorf("stream code is empty")
	}

	// 验证服务器地址格式
	if !strings.HasPrefix(info.ServerAddress, "rtmp://") && !strings.HasPrefix(info.ServerAddress, "rtmps://") {
		return fmt.Errorf("invalid server address format: %s", info.ServerAddress)
	}

	// 验证推流码格式
	if len(info.StreamCode) < 10 {
		return fmt.Errorf("stream code too short: %s", info.StreamCode)
	}

	// 验证端口范围
	if info.Port <= 0 || info.Port > 65535 {
		return fmt.Errorf("invalid port: %d", info.Port)
	}

	return nil
}

// FormatRTMPInfo 格式化RTMP信息为字符串
func FormatRTMPInfo(info *RTMPInfo) string {
	if info == nil {
		return ""
	}

	return fmt.Sprintf(
		"服务器地址: %s\n推流码: %s\n完整地址: %s\n协议: %s\n端口: %d\n检测时间: %s",
		info.ServerAddress,
		info.StreamCode,
		info.FullURL,
		info.Protocol,
		info.Port,
		info.DetectedAt.Format("2006-01-02 15:04:05"),
	)
}
