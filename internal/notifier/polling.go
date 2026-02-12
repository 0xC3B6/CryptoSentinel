// Package notifier Telegram长轮询命令处理
package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const pollTimeout = 30 // seconds

// telegramUpdate Telegram Update对象
type telegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *telegramMessage `json:"message,omitempty"`
}

// telegramMessage Telegram Message对象
type telegramMessage struct {
	Chat *telegramChat `json:"chat"`
	Text string        `json:"text"`
}

// telegramChat Telegram Chat对象
type telegramChat struct {
	ID int64 `json:"id"`
}

// getUpdatesResponse getUpdates API响应
type getUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []telegramUpdate `json:"result"`
}

// StartPolling 启动Telegram长轮询，监听用户命令
func (t *TelegramNotifier) StartPolling(ctx context.Context, onCommand func()) {
	// 创建长轮询专用HTTP客户端，超时 = pollTimeout + 10s
	pollClient := &http.Client{
		Timeout: time.Duration(pollTimeout+10) * time.Second,
	}
	// 复用主客户端的Transport（保持代理设置）
	if t.client.Transport != nil {
		pollClient.Transport = t.client.Transport
	}

	// 跳过离线期间积压的旧消息
	offset := t.skipOldUpdates(pollClient)
	log.Println("Telegram 轮询已启动，等待用户命令...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram 轮询已停止")
			return
		default:
		}

		updates, err := t.getUpdates(pollClient, offset)
		if err != nil {
			log.Printf("获取Telegram更新失败: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			offset = update.UpdateID + 1

			if update.Message == nil || update.Message.Text == "" {
				continue
			}

			// 只响应配置的chatID
			chatID := fmt.Sprintf("%d", update.Message.Chat.ID)
			if chatID != t.chatID {
				continue
			}

			if update.Message.Text == "查看今日建议" {
				log.Println("收到用户命令: 查看今日建议")
				_ = t.Send("⏳ 正在获取实时数据，请稍候...")
				onCommand()
			}
		}
	}
}

// skipOldUpdates 跳过离线期间积压的旧消息，返回起始offset
func (t *TelegramNotifier) skipOldUpdates(client *http.Client) int {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=-1&timeout=0", t.botToken)
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("跳过旧消息失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0
	}

	var result getUpdatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0
	}

	if len(result.Result) > 0 {
		latest := result.Result[len(result.Result)-1].UpdateID + 1
		log.Printf("跳过 %d 条旧消息", latest)
		return latest
	}
	return 0
}

// getUpdates 调用Telegram getUpdates API
func (t *TelegramNotifier) getUpdates(client *http.Client, offset int) ([]telegramUpdate, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d",
		t.botToken, offset, pollTimeout)

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求getUpdates失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取getUpdates响应失败: %w", err)
	}

	var result getUpdatesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析getUpdates响应失败: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("getUpdates返回错误")
	}

	return result.Result, nil
}
