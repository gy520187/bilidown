//go:build !dockerheadless

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
)

func main() {
	checkFFmpeg()
	// 启动托盘程序
	systray.Run(onReady, nil)
}

func onReady() {
	// 设置托盘图标
	setIcon()
	// 设置托盘标题
	setTitle()
	// 设置托盘菜单
	setMenuItem()
	// 初始化数据表
	mustInitTables()
	// 配置和启动 HTTP 服务器
	mustRunServer()
	// 调用默认浏览器访问端口
	time.Sleep(time.Millisecond * 1000)
	openBrowser(urlLocalUnix)
	// 保持运行
	select {}
}

// openBrowser 调用系统默认浏览器打开指定 URL
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		log.Printf("openBrowser: %v.", errors.New("unsupported operating system"))
	}
	if err := cmd.Start(); err != nil {
		log.Printf("openBrowser: %v.", err)
	}
	fmt.Printf("Opened in default browser: %s.\n", url)
}

// setIcon 设置托盘图标
func setIcon() {
	var path string
	if runtime.GOOS == "windows" {
		path = "static/favicon.ico"
	} else {
		path = "static/favicon-32x32.png"
	}
	systray.SetIcon(mustReadFile(path))
}

// mustReadFile 返回文件字节内容
func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("os.ReadFile:", err)
	}
	return data
}

// setTitle 设置托盘标题和工具提示
func setTitle() {
	title := "Bilidown"
	tooltip := fmt.Sprintf("%s 视频解析器 %s (port:%d)", title, VERSION, HTTP_PORT)
	// only available on Mac and Windows.
	systray.SetTooltip(tooltip)
}

// setMenuItem 设置托盘菜单
func setMenuItem() {
	openBrowserItemText := fmt.Sprintf("打开主界面 (port:%d)", HTTP_PORT)
	openBrowserItem := systray.AddMenuItem(openBrowserItemText, openBrowserItemText)
	go func() {
		for {
			<-openBrowserItem.ClickedCh
			openBrowser(urlLocalUnix)
		}
	}()

	aboutItemText := "Github 项目主页"
	aboutItem := systray.AddMenuItem(aboutItemText, aboutItemText)
	go func() {
		for {
			<-aboutItem.ClickedCh
			openBrowser("https://github.com/iuroc/bilidown")
		}
	}()

	exitItemText := "退出应用"
	exitItem := systray.AddMenuItem(exitItemText, exitItemText)
	go func() {
		<-exitItem.ClickedCh
		log.Printf("Bilidown has exited.")
		systray.Quit()
	}()
}
