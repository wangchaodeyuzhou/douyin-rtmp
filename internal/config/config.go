package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	App     AppConfig     `yaml:"app"`
	Network NetworkConfig `yaml:"network"`
	UI      UIConfig      `yaml:"ui"`
	Log     LogConfig     `yaml:"log"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Version        string `yaml:"version"`
	CheckUpdate    bool   `yaml:"check_update"`
	AutoSaveConfig bool   `yaml:"auto_save_config"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	DefaultInterface      string `yaml:"default_interface"`
	MonitorAllInterfaces  bool   `yaml:"monitor_all_interfaces"`
	CaptureTimeout        int    `yaml:"capture_timeout"` // 秒
	MaxConcurrentCaptures int    `yaml:"max_concurrent_captures"`
}

// UIConfig 界面配置
type UIConfig struct {
	WindowWidth    float32 `yaml:"window_width"`
	WindowHeight   float32 `yaml:"window_height"`
	WindowX        float32 `yaml:"window_x"`
	WindowY        float32 `yaml:"window_y"`
	Theme          string  `yaml:"theme"` // "light", "dark", "auto"
	ShowPacketLogs bool    `yaml:"show_packet_logs"`
	LogMaxLines    int     `yaml:"log_max_lines"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"` // "debug", "info", "warn", "error"
	MaxEntries int    `yaml:"max_entries"`
	SaveToFile bool   `yaml:"save_to_file"`
	FilePath   string `yaml:"file_path"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	cm := &ConfigManager{}
	cm.configPath = cm.getConfigPath()
	cm.config = cm.loadDefaultConfig()
	return cm
}

// getConfigPath 获取配置文件路径
func (cm *ConfigManager) getConfigPath() string {
	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		// 如果获取失败，使用当前目录
		configDir = "."
	}

	appConfigDir := filepath.Join(configDir, "douyin-rtmp")

	// 确保配置目录存在
	os.MkdirAll(appConfigDir, 0755)

	return filepath.Join(appConfigDir, "config.yaml")
}

// loadDefaultConfig 加载默认配置
func (cm *ConfigManager) loadDefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Version:        "1.0.0",
			CheckUpdate:    true,
			AutoSaveConfig: true,
		},
		Network: NetworkConfig{
			DefaultInterface:      "",
			MonitorAllInterfaces:  false,
			CaptureTimeout:        30,
			MaxConcurrentCaptures: 5,
		},
		UI: UIConfig{
			WindowWidth:    800,
			WindowHeight:   600,
			WindowX:        -1, // -1 表示居中
			WindowY:        -1, // -1 表示居中
			Theme:          "auto",
			ShowPacketLogs: true,
			LogMaxLines:    1000,
		},
		Log: LogConfig{
			Level:      "info",
			MaxEntries: 1000,
			SaveToFile: false,
			FilePath:   filepath.Join(filepath.Dir(cm.configPath), "logs", "app.log"),
		},
	}
}

// Load 从文件加载配置
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置并保存
		return cm.save()
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// 合并默认配置（处理新增的配置项）
	cm.mergeWithDefault(&config)
	cm.config = &config

	return nil
}

// Save 保存配置到文件
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.save()
}

// save 内部保存方法（不加锁）
func (cm *ConfigManager) save() error {
	// 确保目录存在
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(cm.configPath, data, 0644)
}

// mergeWithDefault 合并默认配置
func (cm *ConfigManager) mergeWithDefault(config *Config) {
	defaultConfig := cm.loadDefaultConfig()

	// 检查并设置缺失的字段
	if config.App.Version == "" {
		config.App.Version = defaultConfig.App.Version
	}
	if config.Network.CaptureTimeout == 0 {
		config.Network.CaptureTimeout = defaultConfig.Network.CaptureTimeout
	}
	if config.Network.MaxConcurrentCaptures == 0 {
		config.Network.MaxConcurrentCaptures = defaultConfig.Network.MaxConcurrentCaptures
	}
	if config.UI.WindowWidth == 0 {
		config.UI.WindowWidth = defaultConfig.UI.WindowWidth
	}
	if config.UI.WindowHeight == 0 {
		config.UI.WindowHeight = defaultConfig.UI.WindowHeight
	}
	if config.UI.Theme == "" {
		config.UI.Theme = defaultConfig.UI.Theme
	}
	if config.UI.LogMaxLines == 0 {
		config.UI.LogMaxLines = defaultConfig.UI.LogMaxLines
	}
	if config.Log.Level == "" {
		config.Log.Level = defaultConfig.Log.Level
	}
	if config.Log.MaxEntries == 0 {
		config.Log.MaxEntries = defaultConfig.Log.MaxEntries
	}
	if config.Log.FilePath == "" {
		config.Log.FilePath = defaultConfig.Log.FilePath
	}
}

// GetConfig 获取配置（只读）
func (cm *ConfigManager) GetConfig() Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回配置的副本
	configCopy := *cm.config
	return configCopy
}

// UpdateAppConfig 更新应用配置
func (cm *ConfigManager) UpdateAppConfig(update func(*AppConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	update(&cm.config.App)

	if cm.config.App.AutoSaveConfig {
		return cm.save()
	}
	return nil
}

// UpdateNetworkConfig 更新网络配置
func (cm *ConfigManager) UpdateNetworkConfig(update func(*NetworkConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	update(&cm.config.Network)

	if cm.config.App.AutoSaveConfig {
		return cm.save()
	}
	return nil
}

// UpdateUIConfig 更新界面配置
func (cm *ConfigManager) UpdateUIConfig(update func(*UIConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	update(&cm.config.UI)

	if cm.config.App.AutoSaveConfig {
		return cm.save()
	}
	return nil
}

// UpdateLogConfig 更新日志配置
func (cm *ConfigManager) UpdateLogConfig(update func(*LogConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	update(&cm.config.Log)

	if cm.config.App.AutoSaveConfig {
		return cm.save()
	}
	return nil
}

// SetDefaultInterface 设置默认网络接口
func (cm *ConfigManager) SetDefaultInterface(interfaceName string) error {
	return cm.UpdateNetworkConfig(func(nc *NetworkConfig) {
		nc.DefaultInterface = interfaceName
	})
}

// GetDefaultInterface 获取默认网络接口
func (cm *ConfigManager) GetDefaultInterface() string {
	config := cm.GetConfig()
	return config.Network.DefaultInterface
}

// SetWindowSize 设置窗口大小
func (cm *ConfigManager) SetWindowSize(width, height float32) error {
	return cm.UpdateUIConfig(func(ui *UIConfig) {
		ui.WindowWidth = width
		ui.WindowHeight = height
	})
}

// SetWindowPosition 设置窗口位置
func (cm *ConfigManager) SetWindowPosition(x, y float32) error {
	return cm.UpdateUIConfig(func(ui *UIConfig) {
		ui.WindowX = x
		ui.WindowY = y
	})
}

// GetWindowConfig 获取窗口配置
func (cm *ConfigManager) GetWindowConfig() (width, height, x, y float32) {
	config := cm.GetConfig()
	return config.UI.WindowWidth, config.UI.WindowHeight, config.UI.WindowX, config.UI.WindowY
}

// SetTheme 设置主题
func (cm *ConfigManager) SetTheme(theme string) error {
	return cm.UpdateUIConfig(func(ui *UIConfig) {
		ui.Theme = theme
	})
}

// GetTheme 获取主题
func (cm *ConfigManager) GetTheme() string {
	config := cm.GetConfig()
	return config.UI.Theme
}

// SetCheckUpdate 设置检查更新
func (cm *ConfigManager) SetCheckUpdate(checkUpdate bool) error {
	return cm.UpdateAppConfig(func(app *AppConfig) {
		app.CheckUpdate = checkUpdate
	})
}

// GetCheckUpdate 获取检查更新设置
func (cm *ConfigManager) GetCheckUpdate() bool {
	config := cm.GetConfig()
	return config.App.CheckUpdate
}

// SetListeningAll 设置监听所有接口
func (cm *ConfigManager) SetListeningAll(listeningAll bool) error {
	return cm.UpdateNetworkConfig(func(net *NetworkConfig) {
		net.MonitorAllInterfaces = listeningAll
	})
}

// GetListeningAll 获取监听所有接口设置
func (cm *ConfigManager) GetListeningAll() bool {
	config := cm.GetConfig()
	return config.Network.MonitorAllInterfaces
}

// SetFileMode 设置文件模式
func (cm *ConfigManager) SetFileMode(fileMode bool) error {
	// 这里我们可以扩展配置结构或使用现有的字段
	// 为了简化，我们可以在应用状态中管理这个设置
	return nil
}

// GetFileMode 获取文件模式设置
func (cm *ConfigManager) GetFileMode() bool {
	// 默认返回false
	return false
}

// SetOBSPath 设置OBS路径
func (cm *ConfigManager) SetOBSPath(path string) error {
	// 这里我们可以扩展配置结构或使用现有的字段
	// 为了简化，我们可以在应用状态中管理这个设置
	return nil
}

// GetOBSPath 获取OBS路径
func (cm *ConfigManager) GetOBSPath() string {
	// 默认返回空字符串
	return ""
}

// SetStreamConfigPath 设置推流配置路径
func (cm *ConfigManager) SetStreamConfigPath(path string) error {
	return nil
}

// GetStreamConfigPath 获取推流配置路径
func (cm *ConfigManager) GetStreamConfigPath() string {
	return ""
}

// Reset 重置为默认配置
func (cm *ConfigManager) Reset() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = cm.loadDefaultConfig()
	return cm.save()
}

// GetConfigPath 获取配置文件路径
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// Validate 验证配置的有效性
func (cm *ConfigManager) Validate() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var errors []string

	// 验证网络配置
	if cm.config.Network.CaptureTimeout < 5 {
		errors = append(errors, "capture_timeout must be at least 5 seconds")
	}
	if cm.config.Network.MaxConcurrentCaptures < 1 {
		errors = append(errors, "max_concurrent_captures must be at least 1")
	}

	// 验证UI配置
	if cm.config.UI.WindowWidth < 400 {
		errors = append(errors, "window_width must be at least 400")
	}
	if cm.config.UI.WindowHeight < 300 {
		errors = append(errors, "window_height must be at least 300")
	}
	if cm.config.UI.LogMaxLines < 100 {
		errors = append(errors, "log_max_lines must be at least 100")
	}

	// 验证日志配置
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cm.config.Log.Level] {
		errors = append(errors, "invalid log level: "+cm.config.Log.Level)
	}
	if cm.config.Log.MaxEntries < 100 {
		errors = append(errors, "log max_entries must be at least 100")
	}

	return errors
}

// 全局配置管理器实例
var defaultConfigManager *ConfigManager
var configOnce sync.Once

// GetDefaultConfigManager 获取默认配置管理器
func GetDefaultConfigManager() *ConfigManager {
	configOnce.Do(func() {
		defaultConfigManager = NewConfigManager()
		// 尝试加载配置文件
		if err := defaultConfigManager.Load(); err != nil {
			// 如果加载失败，使用默认配置
			defaultConfigManager.Save() // 保存默认配置
		}
	})
	return defaultConfigManager
}
