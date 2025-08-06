package server

import (
	. "fmt"
	c "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/packet"
	"net"
)

// sem 用于控制并发连接数的信号量
var (
	sem    chan struct{}
	config *c.Config
)

// StartServer 启动MQTT服务器
// port: 服务器监听的端口号
func StartServer() {
	config, _ := c.GetConfig()
	// 创建TCP监听器
	sem = make(chan struct{}, config.MaxWorkers)
	ln, err := net.Listen("tcp", Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port))
	if err != nil {
		c.FatalF("FSD Server Start error: %v", err)
	}
	c.InfoF("FSD Server Listen On " + ln.Addr().String())

	// 确保在函数退出时关闭监听器
	defer func() {
		err := ln.Close()
		if err != nil {
			c.ErrorF("Server close error: %v", err)
		}
	}()

	// 循环接受新的连接
	for {
		conn, err := ln.Accept()
		if err != nil {
			c.ErrorF("Accept connection error: %v", err)
			continue
		}

		c.DebugF("Accepted new connection from %s", conn.RemoteAddr().String())

		// 使用信号量控制并发连接数
		sem <- struct{}{}
		go func(c net.Conn) {
			connection := &packet.ConnectionHandler{
				Conn:   conn,
				ConnId: conn.RemoteAddr().String(),
			}
			connection.HandleConnection()
			// 释放信号量
			<-sem
		}(conn)
	}
}
