// Package http_server
package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	c "github.com/half-nothing/simple-fsd/internal/config"
	"github.com/half-nothing/simple-fsd/internal/fsd_server/packet"
	"github.com/half-nothing/simple-fsd/internal/http_server/controller"
	mid "github.com/half-nothing/simple-fsd/internal/http_server/middleware"
	impl "github.com/half-nothing/simple-fsd/internal/http_server/service"
	. "github.com/half-nothing/simple-fsd/internal/interfaces"
	"github.com/half-nothing/simple-fsd/internal/interfaces/service"
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

func StartHttpServer(applicationContent *ApplicationContent) {
	config := applicationContent.Config()
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

	emailService := impl.NewEmailService(config.Server.HttpServer.Email)
	impl.InitValidator(config.Server.HttpServer.Limits)

	var storeService service.StoreServiceInterface
	storeService = impl.NewLocalStoreService(httpConfig.Store)
	switch httpConfig.Store.StoreType {
	case 1:
		storeService = impl.NewALiYunOssStoreService(storeService, httpConfig.Store)
	case 2:
		storeService = impl.NewTencentCosStoreService(storeService, httpConfig.Store)
	}

	userService := impl.NewUserService(emailService, httpConfig, applicationContent.UserOperation(), applicationContent.HistoryOperation(), storeService)
	clientManager := packet.NewClientManager(applicationContent)
	clientService := impl.NewClientService(httpConfig, clientManager, emailService, applicationContent.UserOperation())
	serverService := impl.NewServerService(config.Server, applicationContent.UserOperation(), applicationContent.ActivityOperation())
	activityService := impl.NewActivityService(httpConfig, applicationContent.UserOperation(), applicationContent.ActivityOperation(), storeService)

	userController := controller.NewUserHandler(userService)
	emailController := controller.NewEmailController(emailService)
	clientController := controller.NewClientController(clientService)
	serverController := controller.NewServerController(serverService)
	activityController := controller.NewActivityController(activityService)
	fileController := controller.NewFileController(storeService)

	apiGroup := e.Group("/api")
	apiGroup.POST("/sessions", userController.UserLoginHandler)
	apiGroup.GET("/sessions", userController.GetToken, jwtMiddleware)
	apiGroup.POST("/codes", emailController.SendVerifyEmail)
	apiGroup.GET("/profile", userController.GetCurrentUserProfileHandler, jwtMiddleware)
	apiGroup.PATCH("/profile", userController.EditCurrentProfileHandler, jwtMiddleware)
	apiGroup.GET("/history", userController.GetUserHistory, jwtMiddleware)

	userGroup := apiGroup.Group("/users")
	userGroup.POST("", userController.UserRegisterHandler)
	userGroup.GET("", userController.GetUsers, jwtMiddleware)
	userGroup.GET("/controllers", userController.GetControllers, jwtMiddleware)
	userGroup.GET("/availability", userController.CheckUserAvailabilityHandler)
	userGroup.GET("/:uid/profile", userController.GetUserProfileHandler, jwtMiddleware)
	userGroup.PATCH("/:uid/profile", userController.EditProfileHandler, jwtMiddleware)
	userGroup.PATCH("/:uid/permission", userController.EditUserPermission, jwtMiddleware)
	userGroup.PUT("/:uid/rating", userController.EditUserRating, jwtMiddleware)

	clientGroup := apiGroup.Group("/clients")
	clientGroup.GET("/status", func(c echo.Context) error { return c.String(http.StatusOK, whazzupContent) })
	clientGroup.GET("", clientController.GetOnlineClients)
	clientGroup.GET("/paths", clientController.GetClientPath, jwtMiddleware)
	clientGroup.POST("/:callsign/message", clientController.SendMessageToClient, jwtMiddleware)
	clientGroup.DELETE("/:callsign", clientController.KillClient, jwtMiddleware)

	serverGroup := apiGroup.Group("/server")
	serverGroup.GET("/config", serverController.GetServerConfig)
	serverGroup.GET("/info", serverController.GetServerInfo, jwtMiddleware)
	serverGroup.GET("/rating", serverController.GetServerOnlineTime, jwtMiddleware)

	activityGroup := apiGroup.Group("/activities")
	activityGroup.GET("", activityController.GetActivities, jwtMiddleware)
	activityGroup.GET("/list", activityController.GetActivitiesPage, jwtMiddleware)
	activityGroup.GET("/:id", activityController.GetActivityInfo, jwtMiddleware)
	activityGroup.POST("", activityController.AddActivity, jwtMiddleware)
	activityGroup.DELETE("/:id", activityController.DeleteActivity, jwtMiddleware)
	activityGroup.POST("/:id/controllers/:facility_id", activityController.ControllerJoin, jwtMiddleware)
	activityGroup.DELETE("/:id/controllers/:facility_id", activityController.ControllerLeave, jwtMiddleware)
	activityGroup.POST("/:id/pilots", activityController.PilotJoin, jwtMiddleware)
	activityGroup.DELETE("/:id/pilots", activityController.PilotLeave, jwtMiddleware)
	activityGroup.PUT("/:id/status", activityController.EditActivityStatus, jwtMiddleware)
	activityGroup.PUT("/:id/pilots/:pilot_id/status", activityController.EditPilotStatus, jwtMiddleware)
	activityGroup.PUT("/:id", activityController.EditActivity, jwtMiddleware)

	fileGroup := apiGroup.Group("/files")
	fileGroup.POST("/images", fileController.UploadImages, jwtMiddleware)

	apiGroup.Use(middleware.Static(httpConfig.Store.LocalStorePath))

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
		c.Fatal("Http fsd_server error: %v", err)
	}
}
