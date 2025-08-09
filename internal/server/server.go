package server

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	__ "github.com/half-nothing/fsd-server/internal/grpc"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func StartHttpServer() {

}

func StartGRPCServer() {
	config, _ := c.GetConfig()
	ln, err := net.Listen("tcp", config.Server.GRPCServer.Address)
	if err != nil {
		c.FatalF("Fail to open grpc port: %v", err)
		return
	}
	grpcServer := grpc.NewServer()
	__.RegisterServerStatusServer(grpcServer, __.NewGrpcServer(config.Server.GRPCServer.CacheDuration))
	reflection.Register(grpcServer)
	c.NewCleaner().Add(__.NewGrpcShutdownCallback(grpcServer))
	c.InfoF("GRPC server listen on %s", ln.Addr().String())
	err = grpcServer.Serve(ln)
	if err != nil {
		c.FatalF("grpc failed to serve: %v", err)
		return
	}
}

// StartFSDServer 启动FSD服务器
func StartFSDServer() {
	config, _ := c.GetConfig()

	// 初始化客户端管理器
	_ = packet.GetClientManager()

	// 创建TCP监听器
	sem := make(chan struct{}, config.Server.FSDServer.MaxWorkers)
	ln, err := net.Listen("tcp", config.Server.FSDServer.Address)
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
