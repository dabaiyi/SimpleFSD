// Package grpc_server
package grpc_server

import (
	"context"
	c "github.com/half-nothing/fsd-server/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type GrpcShutdownCallback struct {
	grpcServer *grpc.Server
}

func NewGrpcShutdownCallback(grpcServer *grpc.Server) *GrpcShutdownCallback {
	return &GrpcShutdownCallback{
		grpcServer: grpcServer,
	}
}

func (g *GrpcShutdownCallback) Invoke(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		g.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		g.grpcServer.Stop()
		return timeoutCtx.Err()
	}
}

func StartGRPCServer(config *c.GRPCServerConfig) {
	ln, err := net.Listen("tcp", config.Address)
	if err != nil {
		c.FatalF("Fail to open grpc_server port: %v", err)
		return
	}
	c.InfoF("GRPC fsd_server listen on %s", ln.Addr().String())
	grpcServer := grpc.NewServer()
	RegisterServerStatusServer(grpcServer, NewGrpcServer(config.CacheDuration))
	reflection.Register(grpcServer)
	c.GetCleaner().Add(NewGrpcShutdownCallback(grpcServer))
	err = grpcServer.Serve(ln)
	if err != nil {
		c.FatalF("grpc_server failed to serve: %v", err)
		return
	}
}
