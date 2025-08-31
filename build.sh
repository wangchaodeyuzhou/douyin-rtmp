#!/bin/bash

# 抖音RTMP推流地址获取工具构建脚本 (Linux/macOS)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="douyin-rtmp"
VERSION="v1.0.0"
AUTHOR="Go版本"

echo -e "${BLUE}=== 抖音RTMP推流地址获取工具构建脚本 ===${NC}"
echo -e "${BLUE}版本: ${VERSION}${NC}"
echo -e "${BLUE}作者: ${AUTHOR}${NC}"
echo ""

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: 未找到Go环境，请先安装Go 1.21+${NC}"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}Go版本: ${GO_VERSION}${NC}"
}

# 检查依赖
check_dependencies() {
    echo -e "${YELLOW}检查系统依赖...${NC}"
    
    # 检查libpcap
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if ! ldconfig -p | grep -q libpcap; then
            echo -e "${RED}警告: 未找到libpcap，请安装:${NC}"
            echo "  Ubuntu/Debian: sudo apt-get install libpcap-dev"
            echo "  CentOS/RHEL: sudo yum install libpcap-devel"
        else
            echo -e "${GREEN}libpcap: 已安装${NC}"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        if ! brew list libpcap &>/dev/null; then
            echo -e "${YELLOW}警告: 建议安装libpcap: brew install libpcap${NC}"
        else
            echo -e "${GREEN}libpcap: 已安装${NC}"
        fi
    fi
}

# 清理构建文件
clean() {
    echo -e "${YELLOW}清理构建文件...${NC}"
    rm -rf build/
    rm -f ${PROJECT_NAME}
    rm -f ${PROJECT_NAME}.exe
    echo -e "${GREEN}清理完成${NC}"
}

# 下载依赖
download_deps() {
    echo -e "${YELLOW}下载Go依赖...${NC}"
    go mod download
    go mod tidy
    echo -e "${GREEN}依赖下载完成${NC}"
}

# 运行测试
run_tests() {
    echo -e "${YELLOW}运行单元测试...${NC}"
    go test -v ./...
    echo -e "${GREEN}测试完成${NC}"
}

# 构建应用
build() {
    local target_os=${1:-$(go env GOOS)}
    local target_arch=${2:-$(go env GOARCH)}
    
    echo -e "${YELLOW}构建应用 (${target_os}/${target_arch})...${NC}"
    
    # 创建构建目录
    mkdir -p build/${target_os}_${target_arch}
    
    # 设置输出文件名
    local output_name="${PROJECT_NAME}"
    if [ "$target_os" = "windows" ]; then
        output_name="${PROJECT_NAME}.exe"
    fi
    
    # 设置构建标志
    local ldflags="-s -w -X main.AppVersion=${VERSION}"
    
    # 构建
    GOOS=${target_os} GOARCH=${target_arch} go build \
        -ldflags="${ldflags}" \
        -o build/${target_os}_${target_arch}/${output_name} \
        .
    
    # 复制资源文件（如果存在）
    if [ -d "assets" ]; then
        cp -r assets build/${target_os}_${target_arch}/
    fi
    
    echo -e "${GREEN}构建完成: build/${target_os}_${target_arch}/${output_name}${NC}"
}

# 构建所有平台
build_all() {
    echo -e "${YELLOW}构建所有平台版本...${NC}"
    
    # 常见平台组合
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "windows/amd64"
        "darwin/amd64"
        "darwin/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        echo -e "${BLUE}构建 ${os}/${arch}...${NC}"
        build "$os" "$arch"
    done
    
    echo -e "${GREEN}所有平台构建完成${NC}"
}

# 创建发布包
package() {
    echo -e "${YELLOW}创建发布包...${NC}"
    
    mkdir -p dist
    
    for dir in build/*/; do
        if [ -d "$dir" ]; then
            platform_name=$(basename "$dir")
            echo -e "${BLUE}打包 ${platform_name}...${NC}"
            
            # 创建压缩包
            cd build
            if [[ "$platform_name" == *"windows"* ]]; then
                zip -r ../dist/${PROJECT_NAME}_${VERSION}_${platform_name}.zip ${platform_name}/
            else
                tar -czf ../dist/${PROJECT_NAME}_${VERSION}_${platform_name}.tar.gz ${platform_name}/
            fi
            cd ..
        fi
    done
    
    echo -e "${GREEN}发布包创建完成，位于 dist/ 目录${NC}"
}

# 安装到系统
install() {
    echo -e "${YELLOW}安装到系统...${NC}"
    
    local install_dir="/usr/local/bin"
    local binary_name="${PROJECT_NAME}"
    
    if [ ! -f "build/$(go env GOOS)_$(go env GOARCH)/${binary_name}" ]; then
        echo -e "${RED}错误: 未找到构建文件，请先运行构建${NC}"
        exit 1
    fi
    
    # 复制到系统目录
    sudo cp "build/$(go env GOOS)_$(go env GOARCH)/${binary_name}" "${install_dir}/"
    sudo chmod +x "${install_dir}/${binary_name}"
    
    echo -e "${GREEN}安装完成: ${install_dir}/${binary_name}${NC}"
    echo -e "${BLUE}现在可以使用 '${binary_name}' 命令启动程序${NC}"
}

# 卸载
uninstall() {
    echo -e "${YELLOW}从系统卸载...${NC}"
    
    local install_dir="/usr/local/bin"
    local binary_name="${PROJECT_NAME}"
    
    if [ -f "${install_dir}/${binary_name}" ]; then
        sudo rm "${install_dir}/${binary_name}"
        echo -e "${GREEN}卸载完成${NC}"
    else
        echo -e "${YELLOW}未找到已安装的程序${NC}"
    fi
}

# 显示帮助
show_help() {
    echo "用法: $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  check-deps    检查系统依赖"
    echo "  clean         清理构建文件"
    echo "  deps          下载Go依赖"
    echo "  test          运行单元测试"
    echo "  build         构建当前平台版本"
    echo "  build-all     构建所有平台版本"
    echo "  package       创建发布包"
    echo "  install       安装到系统"
    echo "  uninstall     从系统卸载"
    echo "  help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 build          # 构建当前平台"
    echo "  $0 build-all      # 构建所有平台"
    echo "  $0 package        # 创建发布包"
}

# 主函数
main() {
    check_go
    
    case "${1:-build}" in
        "check-deps")
            check_dependencies
            ;;
        "clean")
            clean
            ;;
        "deps")
            download_deps
            ;;
        "test")
            run_tests
            ;;
        "build")
            download_deps
            build
            ;;
        "build-all")
            download_deps
            build_all
            ;;
        "package")
            package
            ;;
        "install")
            install
            ;;
        "uninstall")
            uninstall
            ;;
        "help")
            show_help
            ;;
        *)
            echo -e "${RED}未知命令: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"