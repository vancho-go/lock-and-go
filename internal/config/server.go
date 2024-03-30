package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
)

type server struct {
	Address     string
	DatabaseURI string
	LogLevel    string
}

type serverLoader interface {
	load() (server, error)
}

type flagLoader struct{}

func (f *flagLoader) load() (server, error) {
	serverAddress := flag.String("a", "", "address:port to run server")
	databaseURI := flag.String("d", "", "connection string for driver to establish connection to the conn")
	logLevel := flag.String("l", "", "logger level")
	flag.Parse()

	return server{
		Address:     *serverAddress,
		DatabaseURI: *databaseURI,
		LogLevel:    *logLevel,
	}, nil
}

type envLoader struct{}

func (e *envLoader) load() (server, error) {
	return server{
		Address:     os.Getenv("SERVER_ADDRESS"),
		DatabaseURI: os.Getenv("DATABASE_URI"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
	}, nil
}

func newConfigLoader(loaderType string) (serverLoader, error) {
	switch loaderType {
	case "env":
		return &envLoader{}, nil
	case "flag":
		return &flagLoader{}, nil
	default:
		return nil, fmt.Errorf("newConfigLoader: Unsupported config loader type: %s", loaderType)
	}
}

func isConfigFull(config any) error {
	val := reflect.ValueOf(config)
	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		// Предполагаем, что все поля - строки. Проверяем только строки на пустоту.
		if field.Type.Kind() == reflect.String && value.String() == "" {
			return fmt.Errorf("isConfigFull: поле %s не должно быть пустым", field.Name)
		}
	}
	return nil
}

func NewServer(loaderType string) (*server, error) {
	loader, err := newConfigLoader(loaderType)
	if err != nil {
		return nil, fmt.Errorf("newServer: %w", err)
	}

	config, err := loader.load()
	if err != nil {
		return nil, fmt.Errorf("newServer: %w", err)
	}

	err = isConfigFull(config)
	if err != nil {
		return nil, fmt.Errorf("newServer: %w", err)
	}

	return &config, nil
}
