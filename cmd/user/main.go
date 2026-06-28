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

	database "base-api/opt/db"
	"base-api/pkg/jwt"
	"base-api/pkg/logs"
	"base-api/srv/user/domain"
	"base-api/srv/user/handlers"
	"base-api/srv/user/repositories"
	"base-api/srv/user/usecases"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	_ = godotenv.Load()

	port := getEnv("PORT", "8081")
	env := getEnv("ENV", "development")
	dbURL := getEnv("DATABASE_URL", "postgres://base_user:base_password@localhost:5432/base_db?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key-change-me")

	// Setup logging
	logs.Setup("info", "user-service", env)

	slog.Info("iniciando servicio de usuario...", slog.String("port", port))

	// Conectar a la base de datos
	db, err := database.Connect(dbURL, 30)
	if err != nil {
		slog.Error("error al conectar con base de datos", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Crear tabla de usuarios si no existe (automigración simple para iniciar rápido)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = db.NewCreateTable().
		Model((*domain.User)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		slog.Error("error al crear tabla de usuarios", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Inicializar generador JWT
	jwtGen := jwt.NewGenerator(jwt.Config{
		Secret:     jwtSecret,
		Expiration: 24 * time.Hour,
	})

	// Inicializar capas arquitectónicas
	userRepo := repositories.NewUserRepository(db)
	userUC := usecases.NewUserUsecase(userRepo, jwtGen)
	userHandler := handlers.NewUserHandler(userUC, jwtGen)

	// Iniciar Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(echomw.Recover())
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	// Registrar rutas de usuario
	userHandler.(*handlers.UserHandler).RegisterRoutes(e)

	// Servidor HTTP
	go func() {
		addr := fmt.Sprintf(":%s", port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("error en servidor de usuario", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("deteniendo servicio de usuario...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := e.Shutdown(ctxShutdown); err != nil {
		slog.Error("error durante apagado del servicio de usuario", slog.String("error", err.Error()))
	}

	slog.Info("servicio de usuario detenido")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
