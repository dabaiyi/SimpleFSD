package server

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/controller"
	"github.com/half-nothing/fsd-server/internal/server/defination/interfaces"
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

var (
	ErrMissingOrMalformedJwt = interfaces.ApiStatus{StatusName: "MISSING_OR_MALFORMED_JWT", Description: "缺少JWT令牌或者令牌格式错误", HttpCode: interfaces.BadRequest}
	ErrInvalidOrExpiredJwt   = interfaces.ApiStatus{StatusName: "INVALID_OR_EXPIRED_JWT", Description: "无效或过期的JWT令牌", HttpCode: interfaces.Unauthorized}
	ErrUnknown               = interfaces.ApiStatus{StatusName: "UNKNOWN_JWT_ERROR", Description: "未知的JWT解析错误", HttpCode: interfaces.ServerInternalError}
)

func StartHttpServer() {
	config, _ := c.GetConfig()
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	e.Logger.SetLevel(log.OFF)

	if config.Server.HttpServer.SSL.Enable {
		if config.Server.HttpServer.SSL.CertFile == "" || config.Server.HttpServer.SSL.KeyFile == "" {
			c.WarnF("HTTPS server requires both cert and key files. Cert: %s, Key: %s. Falling back to HTTP",
				config.Server.HttpServer.SSL.CertFile,
				config.Server.HttpServer.SSL.KeyFile)
			config.Server.HttpServer.SSL.Enable = false
		}
	} else if config.Server.HttpServer.SSL.EnableHSTS {
		c.Warn("You can enable HSTS when ssl is not enable!")
		config.Server.HttpServer.SSL.EnableHSTS = false
		config.Server.HttpServer.SSL.HstsExpiredTime = 0
		config.Server.HttpServer.SSL.IncludeDomain = true
	}

	switch config.Server.HttpServer.ProxyType {
	case 0:
		e.IPExtractor = echo.ExtractIPDirect()
	case 1:
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	case 2:
		e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	default:
		c.WarnF("Invalid proxy type %d, using default (direct)", config.Server.HttpServer.ProxyType)
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
		HSTSMaxAge:            config.Server.HttpServer.SSL.HstsExpiredTime,
		HSTSExcludeSubdomains: !config.Server.HttpServer.SSL.IncludeDomain,
	}))
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit("64KB"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	if config.Server.HttpServer.RateLimit <= 0 {
		c.WarnF("Invalid rate limit value %d, using default 100", config.Server.HttpServer.RateLimit)
		config.Server.HttpServer.RateLimit = 100
	}

	if config.Server.HttpServer.RateLimitDuration <= 0 {
		c.WarnF("Invalid rate limit duration %v, using default 1m", config.Server.HttpServer.RateLimitDuration)
		config.Server.HttpServer.RateLimitDuration = time.Minute
	}

	ipPathLimiter := mid.NewSlidingWindowLimiter(
		config.Server.HttpServer.RateLimitDuration,
		config.Server.HttpServer.RateLimit,
	)
	cleanupInterval := config.Server.HttpServer.RateLimitDuration * 2
	if cleanupInterval > time.Hour {
		cleanupInterval = time.Hour
		c.InfoF("Limiting cleanup interval to 1 hour for efficiency")
	}
	ipPathLimiter.StartCleanup(cleanupInterval)

	e.Use(mid.RateLimitMiddleware(ipPathLimiter, mid.CombinedKeyFunc))

	jwtConfig := echojwt.Config{
		SigningKey:    []byte(config.Server.HttpServer.JWT.Secret),
		TokenLookup:   "header:Authorization:Bearer ",
		SigningMethod: "HS512",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(interfaces.Claims)
		},
		ErrorHandler: func(c echo.Context, err error) error {
			var data *interfaces.ApiResponse[any]

			switch {
			case errors.Is(err, echojwt.ErrJWTMissing):
				data = interfaces.NewApiResponse[any](&ErrMissingOrMalformedJwt, interfaces.Unsatisfied, nil)
			case errors.Is(err, echojwt.ErrJWTInvalid):
				data = interfaces.NewApiResponse[any](&ErrInvalidOrExpiredJwt, interfaces.Unsatisfied, nil)
			default:
				data = interfaces.NewApiResponse[any](&ErrUnknown, interfaces.Unsatisfied, nil)
			}

			return data.Response(c)
		},
	}

	jwtMiddleware := echojwt.WithConfig(jwtConfig)

	emailService := service.NewEmailService(config)
	service.InitValidator()
	userService := service.NewUserService(emailService, config)
	clientManager := packet.GetClientManager()
	clientService := service.NewClientService(config, clientManager, emailService)
	userController := controller.NewUserHandler(userService)
	emailController := controller.NewEmailController(emailService)
	clientController := controller.NewClientController(clientService)

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
	clientGroup.GET("", clientController.GetOnlineClients)
	clientGroup.POST("/:callsign/message", clientController.SendMessageToClient, jwtMiddleware)
	clientGroup.DELETE("/:callsign", clientController.KillClient, jwtMiddleware)

	c.GetCleaner().Add(NewHttpServerShutdownCallback(e))

	protocol := "http"
	if config.Server.HttpServer.SSL.Enable {
		protocol = "https"
	}
	c.InfoF("Starting %s server on %s", protocol, config.Server.HttpServer.Address)
	c.InfoF("Rate limit: %d requests per %v",
		config.Server.HttpServer.RateLimit,
		config.Server.HttpServer.RateLimitDuration)

	var err error
	if config.Server.HttpServer.SSL.Enable {
		err = e.StartTLS(
			config.Server.HttpServer.Address,
			config.Server.HttpServer.SSL.CertFile,
			config.Server.HttpServer.SSL.KeyFile,
		)
	} else {
		err = e.Start(config.Server.HttpServer.Address)
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		c.Fatal("Http server error: %v", err)
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
			connection := packet.NewConnectionHandler(conn, conn.RemoteAddr().String())
			connection.HandleConnection()
			// 释放信号量
			<-sem
		}(conn)
	}
}
