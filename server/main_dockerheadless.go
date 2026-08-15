//go:build dockerheadless

package main

import (
	"fmt"
	"log"
)

// main Docker 无头模式入口：跳过系统托盘，直接启动 HTTP 服务
func main() {
	checkFFmpeg()
	mustInitTables()
	mustRunServer()
	log.Printf("Bilidown %s 服务已启动，监听端口: %d", VERSION, HTTP_PORT)
	fmt.Printf("请在浏览器中访问 http://<容器地址>:%d\n", HTTP_PORT)
	// 保持运行
	select {}
}
