# 安装脚本 (Linux/macOS)
#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== 抖音RTMP推流地址获取工具安装脚本 ===${NC}"

# 检查权限
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请以root权限运行此脚本 (sudo $0)${NC}"
    exit 1
fi

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/debian_version ]; then
            OS="ubuntu"
        elif [ -f /etc/redhat-release ]; then
            OS="centos"
        else
            OS="linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    else
        echo -e "${RED}不支持的操作系统: $OSTYPE${NC}"
        exit 1
    fi
    echo -e "${GREEN}检测到操作系统: $OS${NC}"
}

# 安装依赖
install_dependencies() {
    echo -e "${YELLOW}安装系统依赖...${NC}"
    
    case $OS in
        "ubuntu")
            apt-get update
            apt-get install -y libpcap-dev wget curl
            ;;
        "centos")
            yum install -y libpcap-devel wget curl
            ;;
        "macos")
            if ! command -v brew &> /dev/null; then
                echo -e "${RED}请先安装 Homebrew${NC}"
                exit 1
            fi
            brew install libpcap
            ;;
    esac
    
    echo -e "${GREEN}依赖安装完成${NC}"
}

# 下载程序
download_program() {
    echo -e "${YELLOW}下载程序...${NC}"
    
    # 这里应该从实际的发布地址下载
    DOWNLOAD_URL="https://github.com/example/douyin-rtmp/releases/latest/download"
    
    case $OS in
        "ubuntu"|"centos"|"linux")
            FILENAME="douyin-rtmp_v1.0.0_linux_amd64.tar.gz"
            ;;
        "macos")
            FILENAME="douyin-rtmp_v1.0.0_darwin_amd64.tar.gz"
            ;;
    esac
    
    # 下载文件
    wget -O "/tmp/$FILENAME" "$DOWNLOAD_URL/$FILENAME"
    
    # 解压
    cd /tmp
    tar -xzf "$FILENAME"
    
    echo -e "${GREEN}下载完成${NC}"
}

# 安装程序
install_program() {
    echo -e "${YELLOW}安装程序...${NC}"
    
    # 复制二进制文件
    cp douyin-rtmp /usr/local/bin/
    chmod +x /usr/local/bin/douyin-rtmp
    
    # 创建配置目录
    mkdir -p /etc/douyin-rtmp
    
    # 复制配置文件（如果存在）
    if [ -f configs/config.yaml ]; then
        cp configs/config.yaml /etc/douyin-rtmp/
    fi
    
    echo -e "${GREEN}程序安装完成${NC}"
}

# 创建systemd服务（仅Linux）
create_service() {
    if [[ "$OS" == "ubuntu" || "$OS" == "centos" ]]; then
        echo -e "${YELLOW}创建系统服务...${NC}"
        
        cat > /etc/systemd/system/douyin-rtmp.service << EOF
[Unit]
Description=抖音RTMP推流地址获取工具
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/douyin-rtmp -console
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

        systemctl daemon-reload
        echo -e "${GREEN}系统服务创建完成${NC}"
        echo -e "${BLUE}使用以下命令管理服务:${NC}"
        echo "  systemctl start douyin-rtmp    # 启动服务"
        echo "  systemctl stop douyin-rtmp     # 停止服务"
        echo "  systemctl enable douyin-rtmp   # 开机自启"
    fi
}

# 主安装流程
main() {
    detect_os
    install_dependencies
    download_program
    install_program
    create_service
    
    echo -e "${GREEN}=== 安装完成 ===${NC}"
    echo -e "${BLUE}程序已安装到: /usr/local/bin/douyin-rtmp${NC}"
    echo -e "${BLUE}配置文件位置: /etc/douyin-rtmp/config.yaml${NC}"
    echo -e "${BLUE}运行命令: douyin-rtmp${NC}"
}

# 运行主函数
main "$@"