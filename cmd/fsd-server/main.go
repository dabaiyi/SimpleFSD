package main

import (
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	__ "github.com/half-nothing/fsd-server/internal/grpc"
	"github.com/half-nothing/fsd-server/internal/server"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {
	config, err := c.ReadConfig()
	if err != nil {
		c.FatalF("Error occurred while reading config %v", err)
		return
	}
	loggerCallback := c.Init()
	c.Info("Application initializing...")
	cleaner := c.NewCleaner()
	cleaner.Init(loggerCallback)
	defer cleaner.Clean()
	err = database.ConnectDatabase()
	if err != nil {
		c.FatalF("Error occurred while initializing database, details: %v", err)
		return
	}
	c.Info("Database initialized and connection established")
	packet.SyncRatingConfig()
	if config.ServerConfig.EnableGRPC {
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.GRPCPort))
		if err != nil {
			c.FatalF("Fail to open grpc port: %v", err)
			return
		}
		grpcServer := grpc.NewServer()
		__.RegisterServerStatusServer(grpcServer, __.NewGrpcServer(config.ServerConfig.GRPCCacheDuration))
		reflection.Register(grpcServer)
		cleaner.Add(__.NewGrpcShutdownCallback(grpcServer))
		go func() {
			c.DebugF("GRPC server listen on %s", listener.Addr().String())
			err := grpcServer.Serve(listener)
			if err != nil {
				c.FatalF("grpc failed to serve: %v", err)
				return
			}
		}()
	}
	server.StartServer()
}
