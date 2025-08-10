// Package server
package server

import (
	"context"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"time"
)

type HttpServerShutdownCallback struct {
	serverHandler *echo.Echo
}

func NewHttpServerShutdownCallback(serverHandler *echo.Echo) *HttpServerShutdownCallback {
	return &HttpServerShutdownCallback{
		serverHandler: serverHandler,
	}
}

func (hc *HttpServerShutdownCallback) Invoke(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return hc.serverHandler.Shutdown(timeoutCtx)
}

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

type FsdCloseCallback struct {
}

func NewFsdCloseCallback() *FsdCloseCallback {
	return &FsdCloseCallback{}
}

func (dc *FsdCloseCallback) Invoke(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		err := packet.GetClientManager().Shutdown(timeoutCtx)
		if err != nil {
			return
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}
