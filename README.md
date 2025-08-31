# 抖音直播推流地址获取工具 (Go 版本)

一款基于 Go 语言开发的抖音直播推流地址获取工具，使用 gopacket 进行网络数据包捕获，获取到推流地址后，可以通过 OBS 等直播工具进行抖音直播。

## 功能特性

- 🔍 **网络数据包捕获** - 使用 gopacket 库进行高效的网络数据包捕获
- 🎯 **RTMP 地址解析** - 自动识别和提取抖音直播推流服务器地址和推流码
- 🖥️ **现代化 GUI** - 基于 Fyne 框架的跨平台图形界面
- 📡 **网络接口管理** - 智能识别和选择网络接口
- 📊 **实时日志显示** - 分离的操作日志和数据包监控日志
- ⚙️ **配置管理** - 支持配置文件持久化存储
- 🔄 **多接口并发捕获** - 支持同时在多个网络接口上进行捕获

## 系统要求

- **操作系统**: Windows 10/11, macOS 10.14+, Linux (Ubuntu 18.04+)
- **网络**: 需要管理员权限进行网络数据包捕获
- **依赖**: WinPcap/Npcap (Windows), libpcap (Linux/macOS)

## 快速开始

### 1. 安装依赖

#### Windows
```bash
# 下载并安装 Npcap: https://npcap.com/#download
# 或者使用 Chocolatey
choco install npcap
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt-get install libpcap-dev
```

#### macOS
```bash
brew install libpcap
```

### 2. 编译项目

```bash
# 克隆项目
git clone <repository-url>
cd go-rtmp

# 下载依赖
go mod download

# 编译
go build -o douyin-rtmp .

# Windows 下生成 exe 文件
go build -o douyin-rtmp.exe .
```

### 3. 运行程序

```bash
# Linux/macOS (需要 sudo 权限)
sudo ./douyin-rtmp

# Windows (以管理员身份运行)
douyin-rtmp.exe
```

## 使用说明

### 1. 推流码获取

1. 启动程序后，程序会自动检测可用的网络接口
2. 选择合适的网络接口（通常选择主要的网络连接）
3. 点击"开始捕获"按钮
4. 打开抖音直播伴侣开始直播
5. 程序会自动捕获并显示推流服务器地址和推流码

### 2. 日志查看

- **操作日志**: 显示程序运行状态和操作信息
- **数据包监控**: 显示网络数据包的详细信息
- 可以分别清除不同类型的日志

### 3. 配置管理

程序会自动保存以下配置：
- 上次选择的网络接口
- 窗口大小和位置
- 其他用户偏好设置

## 项目结构

```
go-rtmp/
├── main.go                 # 主程序入口
├── internal/
│   ├── capture/            # 数据包捕获模块
│   │   ├── capture.go      # 捕获核心逻辑
│   │   └── interfaces.go   # 接口定义
│   ├── network/            # 网络接口管理
│   │   ├── interfaces.go   # 网络接口检测
│   │   └── utils.go        # 网络工具函数
│   ├── rtmp/              # RTMP 协议解析
│   │   ├── parser.go      # RTMP 数据解析
│   │   └── extractor.go   # 地址提取器
│   ├── logger/            # 日志系统
│   │   └── logger.go      # 日志管理
│   ├── config/            # 配置管理
│   │   └── config.go      # 配置文件处理
│   └── gui/               # GUI 界面
│       ├── app.go         # 应用主窗口
│       ├── components.go  # UI 组件
│       └── handlers.go    # 事件处理
├── configs/               # 配置文件目录
│   └── config.yaml        # 默认配置
├── assets/                # 资源文件
│   └── icon.png          # 应用图标
├── build/                 # 构建脚本
│   ├── build.sh          # Linux/macOS 构建脚本
│   └── build.bat         # Windows 构建脚本
├── go.mod                 # Go 模块定义
├── go.sum                 # 依赖校验文件
└── README.md             # 项目说明
```

## 技术栈

- **语言**: Go 1.21+
- **网络捕获**: google/gopacket
- **GUI 框架**: Fyne v2.5
- **日志库**: sirupsen/logrus
- **配置管理**: gopkg.in/yaml.v3
- **系统信息**: shirou/gopsutil/v3

## 开发指南

### 本地开发环境设置

1. 安装 Go 1.21+ 
2. 安装网络捕获依赖 (见上述系统要求)
3. 克隆项目并安装依赖

```bash
git clone <repository-url>
cd go-rtmp
go mod download
```

### 代码结构说明

- `internal/capture`: 负责网络数据包的捕获和初步处理
- `internal/rtmp`: 专门处理 RTMP 协议相关的数据解析
- `internal/network`: 管理系统网络接口的检测和选择
- `internal/gui`: 基于 Fyne 的图形用户界面
- `internal/logger`: 统一的日志管理系统
- `internal/config`: 应用配置的读取和保存

### 构建和打包

```bash
# 开发模式运行
go run .

# 构建可执行文件
go build -ldflags="-s -w" -o douyin-rtmp .

# 交叉编译 (例如在 Linux 上编译 Windows 版本)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o douyin-rtmp.exe .
```

## 免责声明

1. 本软件仅用于个人学习和测试使用，不可用于任何商业用途
2. 工具仅捕获公开传输的网络数据包，未采用任何破解或逆向工程技术
3. 使用者需遵守抖音平台规定和相关法律法规
4. 开发者不承担因使用本工具而产生的任何责任

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！

## 更新日志

### v1.0.0 (Go 版本)
- 基于 Go 语言重写，性能大幅提升
- 使用 Fyne 框架实现跨平台 GUI
- 优化网络数据包捕获逻辑
- 增加并发捕获支持
- 改进 RTMP 地址解析算法