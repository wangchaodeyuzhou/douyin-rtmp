# 抖音RTMP推流地址获取工具开发指南

## 项目架构

本项目采用模块化架构设计，主要包含以下模块：

### 1. 核心模块 (internal/)

#### 1.1 网络管理 (network/)
- **功能**: 网络接口检测、管理和验证
- **主要文件**:
  - `interfaces.go`: 网络接口管理核心逻辑
  - `utils.go`: 网络工具函数
- **关键特性**:
  - 自动检测可用网络接口
  - 智能分类接口类型（以太网、WiFi、VPN等）
  - 网络连通性测试
  - 跨平台兼容

#### 1.2 数据包捕获 (capture/)
- **功能**: 网络数据包捕获和处理
- **主要文件**:
  - `capture.go`: 数据包捕获核心引擎
  - `interfaces.go`: 捕获管理器接口定义
- **关键特性**:
  - 基于gopacket的高性能数据包捕获
  - 多接口并发捕获支持
  - 实时数据包分析
  - 捕获状态管理

#### 1.3 RTMP解析 (rtmp/)
- **功能**: RTMP协议数据解析和地址提取
- **主要文件**:
  - `parser.go`: RTMP数据解析器
- **关键特性**:
  - 智能RTMP地址识别
  - 推流码提取算法
  - 多种RTMP格式支持
  - 数据包有效性验证

#### 1.4 配置管理 (config/)
- **功能**: 应用配置的读取、保存和管理
- **主要文件**:
  - `config.go`: 配置管理核心
- **关键特性**:
  - YAML格式配置文件
  - 热配置更新
  - 配置验证和默认值
  - 跨平台配置路径

#### 1.5 日志系统 (logger/)
- **功能**: 统一的日志管理和记录
- **主要文件**:
  - `logger.go`: 日志管理器
- **关键特性**:
  - 分级日志记录
  - 控制台和数据包日志分离
  - 日志缓存和限制
  - 实时日志更新

#### 1.6 图形界面 (gui/)
- **功能**: 基于Fyne的跨平台GUI
- **主要文件**:
  - `app.go`: 主应用程序
  - `components.go`: UI组件管理
  - `handlers.go`: 事件处理器
- **关键特性**:
  - 现代化界面设计
  - 响应式布局
  - 实时状态更新
  - 主题支持

### 2. 主程序 (main.go)
- **功能**: 程序入口和模式选择
- **特性**:
  - 命令行参数解析
  - GUI/控制台模式切换
  - 环境检查和权限验证

## 开发环境设置

### 1. 前置要求
```bash
# Go版本要求
go version # >= 1.21

# 系统依赖
# Windows: Npcap
# Linux: libpcap-dev
# macOS: libpcap (通过Homebrew)
```

### 2. 项目初始化
```bash
# 克隆项目
git clone <repository-url>
cd go-rtmp

# 安装依赖
go mod download

# 验证构建
make build
```

### 3. 开发工具推荐
- **IDE**: VS Code + Go扩展 / GoLand
- **调试**: Delve debugger
- **代码检查**: golangci-lint
- **测试**: go test + testify
- **文档**: godoc

## 构建和部署

### 1. 本地开发
```bash
# 开发模式运行
make dev

# 运行测试
make test

# 代码检查
make lint

# 格式化代码
make fmt
```

### 2. 构建发布版本
```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 创建发布包
make package
```

### 3. Docker部署
```bash
# 构建Docker镜像
make docker-build

# 运行容器
docker run --privileged -it douyin-rtmp:latest
```

## 代码规范

### 1. Go代码规范
- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 添加适当的注释和文档
- 函数命名使用驼峰命名法
- 包名使用小写单词

### 2. 项目规范
```go
// 示例：错误处理
func processPacket(packet gopacket.Packet) error {
    if packet == nil {
        return fmt.Errorf("packet is nil")
    }
    
    // 处理逻辑
    return nil
}

// 示例：日志记录
func (c *Capturer) Start() error {
    c.logger.Info("开始数据包捕获")
    
    if err := c.validateConfig(); err != nil {
        c.logger.Error(fmt.Sprintf("配置验证失败: %v", err))
        return err
    }
    
    return nil
}
```

### 3. 错误处理
- 使用标准库的错误处理模式
- 提供有意义的错误信息
- 适当的错误传播和包装
- 关键操作添加恢复机制

### 4. 并发安全
```go
// 示例：使用互斥锁保护共享数据
type SafeCounter struct {
    mu    sync.RWMutex
    value int64
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *SafeCounter) Value() int64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.value
}
```

## 测试策略

### 1. 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/network

# 测试覆盖率
go test -cover ./...
```

### 2. 集成测试
```bash
# 运行集成测试
go test -tags=integration ./...
```

### 3. 性能测试
```bash
# 基准测试
go test -bench=. ./...

# 内存分析
go test -benchmem -memprofile=mem.prof ./...
```

## 性能优化

### 1. 数据包捕获优化
- 使用缓冲通道避免阻塞
- 合理设置捕获过滤器
- 并发处理数据包
- 及时释放资源

### 2. 内存管理
- 使用对象池减少GC压力
- 避免不必要的内存分配
- 及时清理大对象
- 监控内存使用情况

### 3. 并发优化
```go
// 示例：工作池模式
type WorkerPool struct {
    workers   int
    jobQueue  chan Job
    wg        sync.WaitGroup
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    for job := range wp.jobQueue {
        job.Process()
    }
}
```

## 调试技巧

### 1. 日志调试
```go
// 使用不同级别的日志
logger.Debug("详细调试信息")
logger.Info("一般信息")
logger.Warn("警告信息")
logger.Error("错误信息")
```

### 2. 性能分析
```bash
# CPU性能分析
go tool pprof cpu.prof

# 内存分析
go tool pprof mem.prof

# 竞态检测
go run -race .
```

### 3. 网络调试
```bash
# 使用Wireshark分析数据包
# 检查BPF过滤器效果
# 验证网络接口状态
```

## 部署注意事项

### 1. 权限要求
- Windows: 管理员权限
- Linux: root权限或CAP_NET_RAW能力
- macOS: root权限

### 2. 依赖安装
```bash
# Windows
# 下载安装Npcap

# Ubuntu/Debian
sudo apt-get install libpcap-dev

# CentOS/RHEL
sudo yum install libpcap-devel

# macOS
brew install libpcap
```

### 3. 防火墙配置
- 确保网络捕获不被防火墙阻止
- 添加程序到防火墙白名单
- 配置适当的网络访问权限

## 故障排除

### 1. 常见问题
- **权限不足**: 确保以管理员/root权限运行
- **依赖缺失**: 检查libpcap/Npcap安装
- **接口选择错误**: 验证网络接口配置
- **防火墙阻止**: 添加程序到白名单

### 2. 调试命令
```bash
# 检查网络接口
ip addr show  # Linux
ipconfig      # Windows
ifconfig      # macOS

# 测试网络捕获
tcpdump -i any # Linux/macOS
```

### 3. 日志分析
- 查看应用日志文件
- 检查系统日志
- 分析错误堆栈信息

## 贡献指南

### 1. 提交规范
- 使用有意义的提交信息
- 遵循约定式提交规范
- 小步快跑，频繁提交
- 提交前运行测试

### 2. PR流程
1. Fork项目
2. 创建功能分支
3. 开发和测试
4. 提交PR
5. 代码审查
6. 合并主分支

### 3. Issue报告
- 使用Issue模板
- 提供详细的复现步骤
- 包含环境信息
- 添加相关日志和截图

## 扩展开发

### 1. 添加新功能
```go
// 1. 定义接口
type NewFeature interface {
    Process() error
    Configure(config Config) error
}

// 2. 实现接口
type NewFeatureImpl struct {
    // 字段定义
}

// 3. 注册到主程序
func init() {
    RegisterFeature("new-feature", NewFeatureImpl{})
}
```

### 2. 插件系统
- 支持动态加载插件
- 定义插件接口规范
- 提供插件开发文档
- 实现插件管理机制

### 3. 协议扩展
- 添加新的协议解析器
- 扩展数据包分析能力
- 支持自定义协议格式
- 提供协议配置选项