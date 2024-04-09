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
	userdata "github.com/vancho-go/lock-and-go/internal/service/user-data"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"log"
	"net/http"
)

const migrationsPath = "internal/repository/storage/psql/migrations"

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

	storage, err := psql.New(ctx, server.DatabaseURI, migrationsPath, logZap)
	if err != nil {
		log.Fatalf("error initialising storage: %v", err)
	}

	jwtManager := jwt.NewJWTManager(config.GetJWTSecretKey(), config.GetJWTTokenDuration())
	authRepo := psql.NewDefaultUserRepository(storage)
	authService := auth.NewUserAuthService(authRepo, *jwtManager)
	authController := handlers.NewUserController(authService, logZap)

	dataRepo := psql.NewDefaultUserDataRepository(storage)
	dataService := userdata.NewDataService(dataRepo, dataRepo, dataRepo)
	dataController := handlers.NewUserDataController(dataService, logZap)

	middles := middlewares.NewMiddlewares(logZap)

	r := chi.NewRouter()
	r.Post("/register", authController.Register)
	r.Post("/login", authController.Authenticate)

	r.Group(func(r chi.Router) {
		r.Use(middles.JWTMiddleware)
		r.Post("/data/sync", dataController.SyncDataChanges)
		r.Get("/data", dataController.GetData)
	})

	err = http.ListenAndServe(server.Address, r)
	if err != nil {
		logZap.Fatalf("error starting server: %v", err)
	}
}
