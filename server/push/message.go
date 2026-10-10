package push

import (
	"strconv"
	"time"
)

func nowText() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func taskDoneText(title, fileName string) string {
	return "【Bilidown】下载完成\n标题：" + title + "\n文件：" + fileName + "\n时间：" + nowText()
}

func taskErrorText(title, errMsg string) string {
	if errMsg == "" {
		errMsg = "未知错误"
	}
	return "【Bilidown】下载失败\n标题：" + title + "\n原因：" + errMsg + "\n时间：" + nowText()
}

func subscriptionUpdatedText(source string, count int) string {
	return "【Bilidown】订阅更新\n订阅源：" + source + "\n新增 " + strconv.Itoa(count) + " 个稿件已加入下载\n时间：" + nowText()
}

func testText() string {
	return "【Bilidown】测试消息：推送通道工作正常\n时间：" + nowText()
}
