package main

import (
	"context"
	"github.com/vancho-go/lock-and-go/internal/config"
	"github.com/vancho-go/lock-and-go/internal/repository/storage/psql"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"log"
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

	database, err := psql.New(ctx, dbURL, migrationsPath, logZap)
	if err != nil {
		log.Fatalf("error initialising database: %v", err)
	}
	_ = database
}
