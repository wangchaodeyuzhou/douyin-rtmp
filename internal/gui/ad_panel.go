package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AdPanel 广告面板组件
type AdPanel struct {
	app         *App
	
	// UI组件
	adCard      *widget.Card
	adContent   *widget.RichText
	
	// 状态变量
	adLoaded    bool
}

// AdData 广告数据结构
type AdData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// NewAdPanel 创建新的广告面板
func NewAdPanel(app *App) *AdPanel {
	ap := &AdPanel{
		app:      app,
		adLoaded: false,
	}
	
	return ap
}

// CreateContent 创建广告面板内容
func (ap *AdPanel) CreateContent() fyne.CanvasObject {
	// 创建广告内容显示区域
	ap.adContent = widget.NewRichText()
	ap.adContent.Wrapping = fyne.TextWrapWord
	
	// 设置默认内容
	ap.adContent.ParseMarkdown("正在加载广告内容...")
	
	// 创建广告卡片
	ap.adCard = widget.NewCard("广告位", "", ap.adContent)
	ap.adCard.Hide() // 默认隐藏，加载成功后显示
	
	return ap.adCard
}

// AsyncFetchAdContent 异步获取广告内容
func (ap *AdPanel) AsyncFetchAdContent() {
	go func() {
		// 延迟1秒后开始获取广告内容
		time.Sleep(1 * time.Second)
		ap.fetchAdContent()
	}()
}

// fetchAdContent 获取广告内容
func (ap *AdPanel) fetchAdContent() {
	defer func() {
		if r := recover(); r != nil {
			ap.app.logger.Error(fmt.Sprintf("获取广告内容时发生错误: %v", r))
		}
	}()

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 尝试获取广告内容
	adURL := "https://10.192.168101.xyz/ad.json"
	resp, err := client.Get(adURL)
	if err != nil {
		ap.app.logger.Error(fmt.Sprintf("获取广告内容失败: %v", err))
		ap.hideAdPanel()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ap.app.logger.Error(fmt.Sprintf("获取广告内容失败，状态码: %d", resp.StatusCode))
		ap.hideAdPanel()
		return
	}

	// 解析广告数据
	var adData AdData
	if err := json.NewDecoder(resp.Body).Decode(&adData); err != nil {
		ap.app.logger.Error(fmt.Sprintf("解析广告数据失败: %v", err))
		ap.hideAdPanel()
		return
	}

	// 更新广告内容
	ap.updateAdContent(adData)
}

// updateAdContent 更新广告内容
func (ap *AdPanel) updateAdContent(adData AdData) {
	// 格式化广告内容
	adMarkdown := fmt.Sprintf(`
# %s

%s

[了解更多](%s)
`, adData.Title, adData.Content, adData.URL)

	// 更新显示内容
	ap.adContent.ParseMarkdown(adMarkdown)
	ap.adCard.SetTitle(adData.Title)
	
	// 显示广告面板
	ap.adCard.Show()
	ap.adLoaded = true
	
	ap.app.logger.Info("广告内容加载成功")
}

// hideAdPanel 隐藏广告面板
func (ap *AdPanel) hideAdPanel() {
	ap.adCard.Hide()
	ap.adLoaded = false
}

// IsAdLoaded 检查广告是否已加载
func (ap *AdPanel) IsAdLoaded() bool {
	return ap.adLoaded
}