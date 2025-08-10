package server

import (
	"errors"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/controller"
	gs "github.com/half-nothing/fsd-server/internal/server/grpc"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
)

func StartHttpServer() {
	config, _ := c.GetConfig()
	e := echo.New()
	if config.Server.HttpServer.EnableSSL && (config.Server.HttpServer.CertFile == "" || config.Server.HttpServer.KeyFile == "") {
		c.WarnF("https server request a cert file and a key file, fallback to http server")
		config.Server.HttpServer.EnableSSL = false
	}

	apiGroup := e.Group("/api")
	apiGroup.POST("/users", controller.UserRegister)

	c.GetCleaner().Add(NewHttpServerShutdownCallback(e))

	if config.Server.HttpServer.EnableSSL {
		if err := e.StartTLS(config.Server.HttpServer.Address, config.Server.HttpServer.CertFile, config.Server.HttpServer.KeyFile); !errors.Is(err, http.ErrServerClosed) {
			c.Fatal("Error %v", err)
		}
	} else {
		if err := e.Start(config.Server.HttpServer.Address); !errors.Is(err, http.ErrServerClosed) {
			c.Fatal("Error %v", err)
		}
	}
}

func StartGRPCServer() {
	config, _ := c.GetConfig()
	ln, err := net.Listen("tcp", config.Server.GRPCServer.Address)
	if err != nil {
		c.FatalF("Fail to open grpc port: %v", err)
		return
	}
	c.InfoF("GRPC server listen on %s", ln.Addr().String())
	grpcServer := grpc.NewServer()
	gs.RegisterServerStatusServer(grpcServer, gs.NewGrpcServer(config.Server.GRPCServer.CacheDuration))
	reflection.Register(grpcServer)
	c.GetCleaner().Add(NewGrpcShutdownCallback(grpcServer))
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
		return
	}
	c.InfoF("FSD Server Listen On " + ln.Addr().String())

	// 确保在函数退出时关闭监听器
	defer func() {
		err := ln.Close()
		if err != nil {
			c.ErrorF("Server close error: %v", err)
		}
	}()

	c.GetCleaner().Add(NewFsdCloseCallback())

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
