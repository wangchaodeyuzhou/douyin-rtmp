package logger

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time              `json:"time"`
	Level   LogLevel               `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// LogHandler 日志处理函数类型
type LogHandler func(entry *LogEntry)

// Logger 自定义日志管理器
type Logger struct {
	logrus      *logrus.Logger
	handlers    []LogHandler
	packetLogs  []LogEntry
	consoleLogs []LogEntry
	mu          sync.RWMutex
	maxEntries  int
}

// NewLogger 创建新的日志管理器
func NewLogger() *Logger {
	logger := &Logger{
		logrus:      logrus.New(),
		handlers:    make([]LogHandler, 0),
		packetLogs:  make([]LogEntry, 0),
		consoleLogs: make([]LogEntry, 0),
		maxEntries:  1000, // 最多保存1000条日志
	}

	// 配置 logrus
	logger.logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logger.logrus.SetLevel(logrus.InfoLevel)

	return logger
}

// AddHandler 添加日志处理器
func (l *Logger) AddHandler(handler LogHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// Debug 记录调试信息
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DEBUG, message, fields...)
}

// Info 记录普通信息
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(INFO, message, fields...)
}

// Warn 记录警告信息
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WARN, message, fields...)
}

// Error 记录错误信息
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(ERROR, message, fields...)
}

// Packet 记录数据包信息
func (l *Logger) Packet(message string, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Time:    time.Now(),
		Level:   INFO,
		Message: message,
	}

	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 添加到数据包日志
	l.packetLogs = append(l.packetLogs, *entry)

	// 限制日志条目数量
	if len(l.packetLogs) > l.maxEntries {
		l.packetLogs = l.packetLogs[1:]
	}

	// 通知处理器
	for _, handler := range l.handlers {
		go handler(entry)
	}
}

// log 内部日志记录方法
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	entry := &LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}

	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	// 记录到 logrus
	logrusEntry := l.logrus.WithTime(entry.Time)
	if entry.Fields != nil {
		logrusEntry = logrusEntry.WithFields(logrus.Fields(entry.Fields))
	}

	switch level {
	case DEBUG:
		logrusEntry.Debug(message)
	case INFO:
		logrusEntry.Info(message)
	case WARN:
		logrusEntry.Warn(message)
	case ERROR:
		logrusEntry.Error(message)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 添加到控制台日志
	l.consoleLogs = append(l.consoleLogs, *entry)

	// 限制日志条目数量
	if len(l.consoleLogs) > l.maxEntries {
		l.consoleLogs = l.consoleLogs[1:]
	}

	// 通知处理器
	for _, handler := range l.handlers {
		go handler(entry)
	}
}

// GetConsoleLogs 获取控制台日志
func (l *Logger) GetConsoleLogs() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 返回副本
	logs := make([]LogEntry, len(l.consoleLogs))
	copy(logs, l.consoleLogs)
	return logs
}

// GetPacketLogs 获取数据包日志
func (l *Logger) GetPacketLogs() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 返回副本
	logs := make([]LogEntry, len(l.packetLogs))
	copy(logs, l.packetLogs)
	return logs
}

// ClearConsoleLogs 清除控制台日志
func (l *Logger) ClearConsoleLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.consoleLogs = l.consoleLogs[:0]
}

// ClearPacketLogs 清除数据包日志
func (l *Logger) ClearPacketLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.packetLogs = l.packetLogs[:0]
}

// ClearAllLogs 清除所有日志
func (l *Logger) ClearAllLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.consoleLogs = l.consoleLogs[:0]
	l.packetLogs = l.packetLogs[:0]
}

// SetMaxEntries 设置最大日志条目数
func (l *Logger) SetMaxEntries(max int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxEntries = max
}

// FormatLogEntry 格式化日志条目为字符串
func (l *Logger) FormatLogEntry(entry *LogEntry) string {
	levelStr := ""
	switch entry.Level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}

	timeStr := entry.Time.Format("2006-01-02 15:04:05")

	result := fmt.Sprintf("[%s] [%s] %s", timeStr, levelStr, entry.Message)

	// 添加字段信息
	if entry.Fields != nil && len(entry.Fields) > 0 {
		result += " |"
		for key, value := range entry.Fields {
			result += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return result
}

// SetConsoles 设置控制台（与Fyne版本兼容）
func (l *Logger) SetConsoles(systemConsole, packetConsole *widget.RichText) {
	// 这里可以保存控制台的引用，但Fyne中我们使用回调机制
	// 所以这个方法可以为空实现，或者保存引用供其他地方使用
}

// ClearConsole 清除控制台日志
func (l *Logger) ClearConsole() {
	l.ClearConsoleLogs()
}

// ClearPacketConsole 清除数据包控制台日志
func (l *Logger) ClearPacketConsole() {
	l.ClearPacketLogs()
}

// GetFormattedConsoleLogs 获取格式化的控制台日志
func (l *Logger) GetFormattedConsoleLogs() []string {
	logs := l.GetConsoleLogs()
	formatted := make([]string, len(logs))
	for i, log := range logs {
		formatted[i] = l.FormatLogEntry(&log)
	}
	return formatted
}

// GetFormattedPacketLogs 获取格式化的数据包日志
func (l *Logger) GetFormattedPacketLogs() []string {
	logs := l.GetPacketLogs()
	formatted := make([]string, len(logs))
	for i, log := range logs {
		formatted[i] = l.FormatLogEntry(&log)
	}
	return formatted
}

// LogStats 日志统计信息
type LogStats struct {
	ConsoleLogCount int `json:"console_log_count"`
	PacketLogCount  int `json:"packet_log_count"`
	MaxEntries      int `json:"max_entries"`
}

// GetStats 获取日志统计信息
func (l *Logger) GetStats() LogStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return LogStats{
		ConsoleLogCount: len(l.consoleLogs),
		PacketLogCount:  len(l.packetLogs),
		MaxEntries:      l.maxEntries,
	}
}
