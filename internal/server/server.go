package server

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/controller"
	. "github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
	gs "github.com/half-nothing/fsd-server/internal/server/grpc"
	mid "github.com/half-nothing/fsd-server/internal/server/middleware"
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"github.com/half-nothing/fsd-server/internal/server/service"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	slogecho "github.com/samber/slog-echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func StartHttpServer(config *c.Config) {
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	e.Logger.SetLevel(log.OFF)
	httpConfig := config.Server.HttpServer

	switch httpConfig.ProxyType {
	case 0:
		e.IPExtractor = echo.ExtractIPDirect()
	case 1:
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	case 2:
		e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	default:
		c.WarnF("Invalid proxy type %d, using default (direct)", httpConfig.ProxyType)
		e.IPExtractor = echo.ExtractIPDirect()
	}

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{Timeout: 30 * time.Second}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(ctx echo.Context, err error, stack []byte) error {
			c.ErrorF("Recovered from a fatal error: %v, stack: %s", err, string(stack))
			return err
		},
	}))

	loggerConfig := slogecho.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,
	}
	e.Use(slogecho.NewWithConfig(slog.Default(), loggerConfig))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            httpConfig.SSL.HstsExpiredTime,
		HSTSExcludeSubdomains: !httpConfig.SSL.IncludeDomain,
	}))
	e.Use(middleware.CORS())
	if httpConfig.BodyLimit != "" {
		e.Use(middleware.BodyLimit(httpConfig.BodyLimit))
	}
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	if httpConfig.Limits.RateLimit <= 0 {
		c.WarnF("Invalid rate limit value %d, using default 15", httpConfig.Limits.RateLimit)
		httpConfig.Limits.RateLimit = 15
	}

	if httpConfig.Limits.RateLimitDuration <= 0 {
		c.WarnF("Invalid rate limit duration %v, using default 1m", httpConfig.Limits.RateLimitDuration)
		httpConfig.Limits.RateLimitDuration = time.Minute
	}

	ipPathLimiter := mid.NewSlidingWindowLimiter(
		httpConfig.Limits.RateLimitDuration,
		httpConfig.Limits.RateLimit,
	)
	cleanupInterval := httpConfig.Limits.RateLimitDuration * 2
	if cleanupInterval > time.Hour {
		cleanupInterval = time.Hour
		c.InfoF("Limiting cleanup interval to 1 hour for efficiency")
	}
	ipPathLimiter.StartCleanup(cleanupInterval)

	whazzupContent := fmt.Sprintf("url0=%s/api/clients", httpConfig.WhazzupUrlHeader)

	e.Use(mid.RateLimitMiddleware(ipPathLimiter, mid.CombinedKeyFunc))

	jwtConfig := echojwt.Config{
		SigningKey:    []byte(httpConfig.JWT.Secret),
		TokenLookup:   "header:Authorization:Bearer ",
		SigningMethod: "HS512",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(Claims)
		},
		ErrorHandler: func(c echo.Context, err error) error {
			var data *ApiResponse[any]
			switch {
			case errors.Is(err, echojwt.ErrJWTMissing):
				data = NewApiResponse[any](&ErrMissingOrMalformedJwt, Unsatisfied, nil)
			case errors.Is(err, echojwt.ErrJWTInvalid):
				data = NewApiResponse[any](&ErrInvalidOrExpiredJwt, Unsatisfied, nil)
			default:
				data = NewApiResponse[any](&ErrUnknown, Unsatisfied, nil)
			}
			return data.Response(c)
		},
	}

	jwtMiddleware := echojwt.WithConfig(jwtConfig)

	emailService := service.NewEmailService(config.Server.HttpServer.Email)
	service.InitValidator(config.Server.HttpServer.Limits)

	var storeService StoreServiceInterface
	storeService = service.NewLocalStoreService(httpConfig.Store)
	switch httpConfig.Store.StoreType {
	case 1:
		storeService = service.NewALiYunOssStoreService(storeService, httpConfig.Store)
	case 2:
		storeService = service.NewTencentCosStoreService(storeService, httpConfig.Store)
	}

	userService := service.NewUserService(emailService, httpConfig)
	clientManager := packet.NewClientManager(config)
	clientService := service.NewClientService(httpConfig, clientManager, emailService)
	serverService := service.NewServerService(config.Server)
	activityService := service.NewActivityService(httpConfig)

	userController := controller.NewUserHandler(userService)
	emailController := controller.NewEmailController(emailService)
	clientController := controller.NewClientController(clientService)
	serverController := controller.NewServerController(serverService)
	activityController := controller.NewActivityController(activityService)
	fileController := controller.NewFileController(storeService)

	apiGroup := e.Group("/api")
	apiGroup.POST("/sessions", userController.UserLoginHandler)
	apiGroup.POST("/codes", emailController.SendVerifyEmail)
	apiGroup.GET("/profile", userController.GetCurrentUserProfileHandler, jwtMiddleware)
	apiGroup.PATCH("/profile", userController.EditCurrentProfileHandler, jwtMiddleware)

	userGroup := apiGroup.Group("/users")
	userGroup.POST("", userController.UserRegisterHandler)
	userGroup.GET("", userController.GetUsers, jwtMiddleware)
	userGroup.GET("/availability", userController.CheckUserAvailabilityHandler)
	userGroup.GET("/:uid/profile", userController.GetUserProfileHandler, jwtMiddleware)
	userGroup.PATCH("/:uid/profile", userController.EditProfileHandler, jwtMiddleware)
	userGroup.PATCH("/:uid/permission", userController.EditUserPermission, jwtMiddleware)
	userGroup.PUT("/:uid/rating", userController.EditUserRating, jwtMiddleware)

	clientGroup := apiGroup.Group("/clients")
	clientGroup.GET("/status", func(c echo.Context) error { return c.String(http.StatusOK, whazzupContent) })
	clientGroup.GET("", clientController.GetOnlineClients)
	clientGroup.POST("/:callsign/message", clientController.SendMessageToClient, jwtMiddleware)
	clientGroup.DELETE("/:callsign", clientController.KillClient, jwtMiddleware)

	serverGroup := apiGroup.Group("/server")
	serverGroup.GET("/config", serverController.GetServerConfig)

	activityGroup := apiGroup.Group("/activities")
	activityGroup.GET("", activityController.GetActivities)
	activityGroup.GET("/:id", activityController.GetActivityInfo)
	activityGroup.POST("", activityController.AddActivity, jwtMiddleware)

	fileGroup := apiGroup.Group("/files")
	fileGroup.POST("/images", fileController.UploadImages, jwtMiddleware)

	e.Use(middleware.Static(httpConfig.Store.LocalStorePath))

	c.GetCleaner().Add(NewHttpServerShutdownCallback(e))

	protocol := "http"
	if httpConfig.SSL.Enable {
		protocol = "https"
	}
	c.InfoF("Starting %s server on %s", protocol, httpConfig.Address)
	c.InfoF("Rate limit: %d requests per %v",
		httpConfig.Limits.RateLimit,
		httpConfig.Limits.RateLimitDuration)

	var err error
	if httpConfig.SSL.Enable {
		err = e.StartTLS(
			httpConfig.Address,
			httpConfig.SSL.CertFile,
			httpConfig.SSL.KeyFile,
		)
	} else {
		err = e.Start(httpConfig.Address)
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		c.Fatal("Http server error: %v", err)
	}
}

func StartGRPCServer(config *c.GRPCServerConfig) {
	ln, err := net.Listen("tcp", config.Address)
	if err != nil {
		c.FatalF("Fail to open grpc port: %v", err)
		return
	}
	c.InfoF("GRPC server listen on %s", ln.Addr().String())
	grpcServer := grpc.NewServer()
	gs.RegisterServerStatusServer(grpcServer, gs.NewGrpcServer(config.CacheDuration))
	reflection.Register(grpcServer)
	c.GetCleaner().Add(NewGrpcShutdownCallback(grpcServer))
	err = grpcServer.Serve(ln)
	if err != nil {
		c.FatalF("grpc failed to serve: %v", err)
		return
	}
}

// StartFSDServer 启动FSD服务器
func StartFSDServer(config *c.Config) {
	// 初始化客户端管理器
	cm := packet.NewClientManager(config)

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

	c.GetCleaner().Add(NewFsdCloseCallback(cm))

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
			connection := packet.NewConnectionHandler(conn, conn.RemoteAddr().String(), config, cm)
			connection.HandleConnection()
			// 释放信号量
			<-sem
		}(conn)
	}
}
