package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diegoaleon/test-app/pkg/logs"
	"github.com/diegoaleon/test-app/srv/gateway"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type Config struct {
	Port           int
	Env            string
	JWTSecret      string
	UserURL        string
	AllowedOrigins string
}

func main() {
	_ = godotenv.Load()

	cfg := loadConfig()

	// Setup logging
	logs.Setup("info", "gateway", cfg.Env)

	slog.Info("iniciando gateway...", slog.Int("port", cfg.Port), slog.String("user_url", cfg.UserURL))

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Configurar rutas y proxies
	gateway.Setup(e, gateway.Config{
		JWTSecret:      cfg.JWTSecret,
		UserURL:        cfg.UserURL,
		AllowedOrigins: cfg.AllowedOrigins,
	})

	// Iniciar servidor
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("error en gateway", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("deteniendo gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error al apagar gateway", slog.String("error", err.Error()))
	}

	slog.Info("gateway detenido")
}

func loadConfig() Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &port); err != nil {
			slog.Warn("puerto inválido, usando default 8080", "value", p, "error", err)
		}
	}

	return Config{
		Port:           port,
		Env:            getEnv("ENV", "development"),
		JWTSecret:      getEnv("JWT_SECRET", "super-secret-key-change-me"),
		UserURL:        getEnv("USER_URL", "http://localhost:8081"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
