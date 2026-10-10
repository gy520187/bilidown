package push

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"bilidown/util"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Config 推送服务配置，持久化在 field 表的 push_* 键中。
type Config struct {
	Enabled    bool
	Channel    string // 空值视为 napcat
	URL        string // OneBot v11 HTTP 根地址
	Token      string // 可为空
	TargetType string // private | group，空值视为 private
	TargetID   string // QQ 号或群号
}

// LoadConfig 从数据库读取推送配置并填充缺省值。
func LoadConfig(db *sql.DB) (Config, error) {
	fields, err := util.GetFields(db, "push_enabled", "push_channel", "push_url", "push_token", "push_target_type", "push_target_id")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Enabled:    fields["push_enabled"] == "1",
		Channel:    fields["push_channel"],
		URL:        fields["push_url"],
		Token:      fields["push_token"],
		TargetType: fields["push_target_type"],
		TargetID:   fields["push_target_id"],
	}
	if cfg.Channel == "" {
		cfg.Channel = "napcat"
	}
	if cfg.TargetType == "" {
		cfg.TargetType = "private"
	}
	return cfg, nil
}

// validate 校验发送所需的必填配置。
func (cfg Config) validate() error {
	if cfg.URL == "" {
		return fmt.Errorf("推送地址为空")
	}
	if cfg.TargetID == "" {
		return fmt.Errorf("推送目标为空")
	}
	if cfg.TargetType != "private" && cfg.TargetType != "group" {
		return fmt.Errorf("推送目标类型无效: %q", cfg.TargetType)
	}
	return nil
}

// send 发送一条消息；未启用时静默跳过，发送失败由调用方记日志。
func send(cfg Config, text string) error {
	if !cfg.Enabled {
		return nil
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	channel, ok := channels[cfg.Channel]
	if !ok {
		return fmt.Errorf("未知推送通道 %q", cfg.Channel)
	}
	return channel.Send(cfg, text)
}

// TaskDoneAsync 推送任务完成消息。
func TaskDoneAsync(title, fileName string) {
	sendAsync(taskDoneText(title, fileName))
}

// TaskErrorAsync 推送任务失败消息。
func TaskErrorAsync(title, errMsg string) {
	sendAsync(taskErrorText(title, errMsg))
}

// SubscriptionUpdatedAsync 推送订阅新稿件提醒。
func SubscriptionUpdatedAsync(source string, count int) {
	sendAsync(subscriptionUpdatedText(source, count))
}

// sendAsync 读取配置后异步推送；任何失败仅记日志，不重试，不影响业务流程。
func sendAsync(text string) {
	go func() {
		db := util.MustGetDB()
		defer db.Close()
		cfg, err := LoadConfig(db)
		if err != nil {
			log.Printf("push: 读取配置失败: %v", err)
			return
		}
		if !cfg.Enabled {
			return
		}
		if err := send(cfg, text); err != nil {
			log.Printf("push: %v", err)
		}
	}()
}

// TestPush 使用传入配置（可含未保存修改）同步发送一条测试消息。
func TestPush(cfg Config) error {
	cfg.Enabled = true
	if cfg.Channel == "" {
		cfg.Channel = "napcat"
	}
	if cfg.TargetType == "" {
		cfg.TargetType = "private"
	}
	return send(cfg, testText())
}
