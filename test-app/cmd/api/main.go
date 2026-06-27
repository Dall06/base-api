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

	"github.com/diegoaleon/test-app/pkg/jwt"
	"github.com/diegoaleon/test-app/pkg/logs"
	"github.com/diegoaleon/test-app/srv/user/handlers"
	"github.com/diegoaleon/test-app/srv/user/repositories"
	"github.com/diegoaleon/test-app/srv/user/usecases"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Cargar .env si existe
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")
	jwtSecret := getEnv("JWT_SECRET", "change-me-please")

	// Configurar Slog estructurado
	logs.Setup("info", "base-api", env)

	slog.Info("iniciando servidor api...", slog.String("env", env), slog.String("port", port))

	// Inicializar generador de JWT
	jwtGen := jwt.NewGenerator(jwt.Config{
		Secret:     jwtSecret,
		Expiration: 24 * time.Hour,
	})

	// Inicializar capas de arquitectura (Hexagonal)
	userRepo := repositories.NewMemoryUserRepository()
	userUC := usecases.NewUserUsecase(userRepo, jwtGen)

	// Crear servidor Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares globales
	e.Use(echomw.Recover())
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
	}))

	// Registrar rutas de servicios
	handlers.RegisterRoutes(e, userUC, jwtGen)

	// Iniciar servidor asíncronamente
	go func() {
		addr := fmt.Sprintf(":%s", port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("error en el servidor", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Esperar señal de interrupción para apagado gracioso
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("apagando el servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error durante el apagado", slog.String("error", err.Error()))
	}

	slog.Info("servidor detenido")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
