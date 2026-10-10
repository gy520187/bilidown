package router

import (
	"encoding/json"
	"net/http"

	"bilidown/push"
	"bilidown/util"
)

// pushTest 使用表单当前配置（含未保存修改）发送一条测试推送消息。
func pushTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}
	defer r.Body.Close()
	var body struct {
		URL        string `json:"url"`
		Token      string `json:"token"`
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	cfg := push.Config{
		Channel:    "napcat",
		URL:        body.URL,
		Token:      body.Token,
		TargetType: body.TargetType,
		TargetID:   body.TargetID,
	}
	if cfg.TargetType == "" {
		cfg.TargetType = "private"
	}
	if err := push.TestPush(cfg); err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "测试消息已发送"}.Write(w)
}
