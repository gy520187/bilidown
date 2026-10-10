package push

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// NapCatChannel 通过 OneBot v11 HTTP API 发送 QQ 消息。
type NapCatChannel struct{}

type oneBotResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Wording string `json:"wording"`
}

func (NapCatChannel) Send(cfg Config, text string) error {
	targetID, err := strconv.ParseInt(cfg.TargetID, 10, 64)
	if err != nil {
		return fmt.Errorf("推送目标 ID 不是数字: %q", cfg.TargetID)
	}

	action := "send_private_msg"
	payload, err := json.Marshal(map[string]interface{}{
		"user_id": targetID,
		"message": text,
	})
	if err != nil {
		return err
	}
	if cfg.TargetType == "group" {
		action = "send_group_msg"
		payload, err = json.Marshal(map[string]interface{}{
			"group_id": targetID,
			"message":  text,
		})
		if err != nil {
			return err
		}
	}

	url := strings.TrimRight(cfg.URL, "/") + "/" + action
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求推送地址失败: %v", err)
	}
	defer resp.Body.Close()

	var result oneBotResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("推送响应解析失败(HTTP %d): %v", resp.StatusCode, err)
	}
	// OneBot v11: retcode 0 表示同步成功，1 表示已转入异步处理
	if resp.StatusCode != http.StatusOK || (result.Retcode != 0 && result.Retcode != 1) {
		summary := result.Wording
		if summary == "" {
			summary = result.Status
		}
		return fmt.Errorf("推送被拒绝(HTTP %d, retcode %d): %s", resp.StatusCode, result.Retcode, summary)
	}
	return nil
}
