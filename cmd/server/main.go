package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/lock-and-go/internal/config"
	"github.com/vancho-go/lock-and-go/internal/controller/http/handlers"
	"github.com/vancho-go/lock-and-go/internal/controller/http/middlewares"
	"github.com/vancho-go/lock-and-go/internal/repository/storage/psql"
	"github.com/vancho-go/lock-and-go/internal/service/auth"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"log"
	"net/http"
)

func main() {
	ctx := context.TODO()

	// Определяем тип загрузчика конфигурации: "env" или "flag"
	loaderType := "flag"

	// Инициализируем конфигурацию сервера
	server, err := config.NewServer(loaderType)
	if err != nil {
		log.Fatalf("error building server configuration: %v", err)
	}

	// Используем полученную конфигурацию
	log.Printf("server configuration: %+v", server)

	logZap, err := logger.New(server.LogLevel)
	if err != nil {
		log.Fatalf("storage could not be closed: %v", err)
	}

	dbURL := "host=localhost port=5432 user=vancho password=vancho_pswd dbname=vancho_db sslmode=disable"
	migrationsPath := "internal/repository/storage/psql/migrations"

	storage, err := psql.New(ctx, dbURL, migrationsPath, logZap)
	if err != nil {
		log.Fatalf("error initialising storage: %v", err)
	}

	userAuthRepo := psql.NewDefaultUserRepository(storage)
	jwtManager := jwt.NewJWTManager(server.JWTSecretKey, server.JWTTokenDuration)
	userAuthService := auth.NewUserAuthService(userAuthRepo, *jwtManager)
	userController := handlers.NewUserController(userAuthService, logZap)

	middles := middlewares.NewMiddlewares(logZap)

	r := chi.NewRouter()
	r.Post("/register", userController.Register)
	r.Post("/login", userController.Authenticate)

	r.Group(func(r chi.Router) {
		r.Use(middles.JWTMiddleware)
		r.Get("/test", userController.Test)
	})

	err = http.ListenAndServe(server.Address, r)
	if err != nil {
		logZap.Fatalf("error starting server: %v", err)
	}
}
