package server

import (
	. "fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"net"
)

// StartServer 启动MQTT服务器
func StartServer() {
	config, _ := c.GetConfig()

	// 初始化客户端管理器
	_ = packet.GetClientManager()

	// 创建TCP监听器
	sem := make(chan struct{}, config.MaxWorkers)
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
				Client: nil,
				User:   nil,
			}
			connection.HandleConnection()
			// 释放信号量
			<-sem
		}(conn)
	}
}
