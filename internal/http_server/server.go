// Package http_server
package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/fsd_server/packet"
	"github.com/half-nothing/fsd-server/internal/http_server/controller"
	middleware2 "github.com/half-nothing/fsd-server/internal/http_server/middleware"
	service2 "github.com/half-nothing/fsd-server/internal/http_server/service"
	"github.com/half-nothing/fsd-server/internal/interfaces/service"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/samber/slog-echo"
	"io"
	"log/slog"
	"net/http"
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

	ipPathLimiter := middleware2.NewSlidingWindowLimiter(
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

	e.Use(middleware2.RateLimitMiddleware(ipPathLimiter, middleware2.CombinedKeyFunc))

	jwtConfig := echojwt.Config{
		SigningKey:    []byte(httpConfig.JWT.Secret),
		TokenLookup:   "header:Authorization:Bearer ",
		SigningMethod: "HS512",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(service.Claims)
		},
		ErrorHandler: func(c echo.Context, err error) error {
			var data *service.ApiResponse[any]
			switch {
			case errors.Is(err, echojwt.ErrJWTMissing):
				data = service.NewApiResponse[any](&service.ErrMissingOrMalformedJwt, service.Unsatisfied, nil)
			case errors.Is(err, echojwt.ErrJWTInvalid):
				data = service.NewApiResponse[any](&service.ErrInvalidOrExpiredJwt, service.Unsatisfied, nil)
			default:
				data = service.NewApiResponse[any](&service.ErrUnknown, service.Unsatisfied, nil)
			}
			return data.Response(c)
		},
	}

	jwtMiddleware := echojwt.WithConfig(jwtConfig)

	emailService := service2.NewEmailService(config.Server.HttpServer.Email)
	service2.InitValidator(config.Server.HttpServer.Limits)

	var storeService service.StoreServiceInterface
	storeService = service2.NewLocalStoreService(httpConfig.Store)
	switch httpConfig.Store.StoreType {
	case 1:
		storeService = service2.NewALiYunOssStoreService(storeService, httpConfig.Store)
	case 2:
		storeService = service2.NewTencentCosStoreService(storeService, httpConfig.Store)
	}

	userService := service2.NewUserService(emailService, httpConfig)
	clientManager := packet.NewClientManager(config)
	clientService := service2.NewClientService(httpConfig, clientManager, emailService)
	serverService := service2.NewServerService(config.Server)
	activityService := service2.NewActivityService(httpConfig)

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

	serverGroup := apiGroup.Group("/fsd_server")
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
	c.InfoF("Starting %s fsd_server on %s", protocol, httpConfig.Address)
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
		c.Fatal("Http fsd_server error: %v", err)
	}
}
