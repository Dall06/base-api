package gateway

import (
	"net/http"
	"strings"

	"base-api/opt/middlewares"
	"base-api/srv/gateway/handlers"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

type Config struct {
	JWTSecret      string
	UserURL        string
	AllowedOrigins string
}

func Setup(e *echo.Echo, cfg Config) {
	e.Use(echomw.Recover())
	e.Use(middlewares.GatewayRequestLogger())
	e.Use(middlewares.ScanRequestMiddleware())
	e.Use(middlewares.AuditMiddleware())

	origins := strings.Split(cfg.AllowedOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     origins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
	}))

	proxy := handlers.NewProxyHandler(cfg.UserURL)

	// Health check
	healthHandler := func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok", "service": "gateway"})
	}
	e.GET("/health", healthHandler)
	e.GET("/api/health", healthHandler)

	api := e.Group("/api/v1")

	// Auth routes (public)
	api.POST("/auth/signup", proxy.ProxyToUser)
	api.POST("/auth/login", proxy.ProxyToUser)

	// User routes (protected)
	// Gateway doesn't validate JWT here, it proxies the request including Authorization header,
	// so the target service (user-service) validates it. This is standard in microservices!
	api.GET("/users/me", proxy.ProxyToUser)
}
