package push

// Channel 是推送通道抽象；新增通道（如 Bark、Telegram）在此注册实现即可接入全部事件。
type Channel interface {
	Send(cfg Config, text string) error
}

var channels = map[string]Channel{
	"napcat": NapCatChannel{},
}
