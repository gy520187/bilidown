package push

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func newOneBotMock(t *testing.T, respBody string, status int) (*httptest.Server, *string, *string, *map[string]interface{}) {
	t.Helper()
	var gotPath, gotAuth string
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(server.Close)
	return server, &gotPath, &gotAuth, &gotBody
}

func TestNapCatPrivateSend(t *testing.T) {
	server, path, auth, body := newOneBotMock(t, `{"status":"ok","retcode":0}`, http.StatusOK)
	cfg := Config{Enabled: true, Channel: "napcat", URL: server.URL, Token: "secret", TargetType: "private", TargetID: "10001"}
	if err := TestPush(cfg); err != nil {
		t.Fatalf("TestPush: %v", err)
	}
	if *path != "/send_private_msg" {
		t.Fatalf("path = %q", *path)
	}
	if *auth != "Bearer secret" {
		t.Fatalf("auth = %q", *auth)
	}
	if (*body)["user_id"].(float64) != 10001 {
		t.Fatalf("user_id = %v", (*body)["user_id"])
	}
	if msg, _ := (*body)["message"].(string); !strings.Contains(msg, "测试消息") {
		t.Fatalf("message = %q", msg)
	}
}

func TestNapCatGroupSend(t *testing.T) {
	server, path, _, body := newOneBotMock(t, `{"status":"async","retcode":1}`, http.StatusOK)
	cfg := Config{Enabled: true, Channel: "napcat", URL: server.URL, TargetType: "group", TargetID: "20002"}
	if err := TestPush(cfg); err != nil {
		t.Fatalf("TestPush: %v", err)
	}
	if *path != "/send_group_msg" {
		t.Fatalf("path = %q", *path)
	}
	if _, hasUser := (*body)["user_id"]; hasUser {
		t.Fatalf("group send should not carry user_id: %v", *body)
	}
	if (*body)["group_id"].(float64) != 20002 {
		t.Fatalf("group_id = %v", (*body)["group_id"])
	}
}

func TestNapCatRejectsFailure(t *testing.T) {
	server, _, _, _ := newOneBotMock(t, `{"status":"failed","retcode":1400,"wording":"API不存在"}`, http.StatusNotFound)
	cfg := Config{Enabled: true, Channel: "napcat", URL: server.URL, TargetType: "private", TargetID: "10001"}
	err := TestPush(cfg)
	if err == nil {
		t.Fatal("expected error on retcode 1400")
	}
	if !strings.Contains(err.Error(), "1400") || !strings.Contains(err.Error(), "API不存在") {
		t.Fatalf("error should carry retcode and wording: %v", err)
	}
}

func TestNapCatRejectsInvalidTargetID(t *testing.T) {
	cfg := Config{Enabled: true, Channel: "napcat", URL: "http://127.0.0.1:1", TargetType: "private", TargetID: "abc"}
	if err := TestPush(cfg); err == nil || !strings.Contains(err.Error(), "不是数字") {
		t.Fatalf("expected invalid target id error, got %v", err)
	}
}

func TestSendSkipsWhenDisabled(t *testing.T) {
	// 指向一个必然不可达的地址，若未静默跳过会返回错误
	cfg := Config{Enabled: false, Channel: "napcat", URL: "http://127.0.0.1:1", TargetType: "private", TargetID: "10001"}
	if err := send(cfg, "x"); err != nil {
		t.Fatalf("disabled push should skip silently, got %v", err)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE "field" ("name" TEXT PRIMARY KEY NOT NULL, "value" TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO "field" ("name", "value") VALUES ('push_enabled', '1'), ('push_url', 'http://127.0.0.1:3000'), ('push_target_id', '10086')`); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(db)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Fatal("enabled should be true")
	}
	if cfg.Channel != "napcat" {
		t.Fatalf("channel default = %q", cfg.Channel)
	}
	if cfg.TargetType != "private" {
		t.Fatalf("target type default = %q", cfg.TargetType)
	}
	if cfg.URL != "http://127.0.0.1:3000" || cfg.TargetID != "10086" {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestMessageTexts(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		contain []string
	}{
		{"done", taskDoneText("凡人修仙传.S01E194", "凡人修仙传.S01E194.mp4"), []string{"下载完成", "凡人修仙传.S01E194", ".mp4", "时间："}},
		{"error", taskErrorText("标题A", "os.Rename: no space"), []string{"下载失败", "标题A", "no space"}},
		{"error empty", taskErrorText("标题A", ""), []string{"未知错误"}},
		{"subscription", subscriptionUpdatedText("间谍过家家", 3), []string{"订阅更新", "间谍过家家", "新增 3 个稿件"}},
		{"test", testText(), []string{"测试消息"}},
	}
	for _, c := range cases {
		for _, sub := range c.contain {
			if !strings.Contains(c.text, sub) {
				t.Fatalf("%s message %q should contain %q", c.name, c.text, sub)
			}
		}
	}
}
