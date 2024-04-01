package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"
)

var jwtSecretKey string

type server struct {
	Address          string
	DatabaseURI      string
	LogLevel         string
	JWTSecretKey     string
	JWTTokenDuration time.Duration
}

type serverLoader interface {
	load() (*server, error)
}

type flagLoader struct{}

func (f *flagLoader) load() (*server, error) {
	serverAddress := flag.String("a", "", "address:port to run server")
	databaseURI := flag.String("d", "", "connection string for driver to establish connection to the conn")
	logLevel := flag.String("l", "", "logger level")
	jwtSecretKey := flag.String("js", "", "jwt secret key")
	jwtTokenDuration := flag.String("jt", "", "jwt secret key valid time")

	flag.Parse()

	duration, err := time.ParseDuration(*jwtTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("load: failed to decode jwt duration: %v", err)
	}

	return &server{
		Address:          *serverAddress,
		DatabaseURI:      *databaseURI,
		LogLevel:         *logLevel,
		JWTSecretKey:     *jwtSecretKey,
		JWTTokenDuration: duration,
	}, nil
}

type envLoader struct{}

func (e *envLoader) load() (*server, error) {
	duration, err := time.ParseDuration(os.Getenv("JWT_TOKEN_DURATION"))
	if err != nil {
		return nil, fmt.Errorf("load: failed to decode jwt duration: %v", err)
	}

	return &server{
		Address:          os.Getenv("SERVER_ADDRESS"),
		DatabaseURI:      os.Getenv("DATABASE_URI"),
		LogLevel:         os.Getenv("LOG_LEVEL"),
		JWTSecretKey:     os.Getenv("JWT_SECRET_KEY"),
		JWTTokenDuration: duration,
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
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Разыменовываем указатель
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		// Проверяем, что значение не равно нулевому значению для своего типа
		if reflect.DeepEqual(value.Interface(), reflect.Zero(field.Type).Interface()) {
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
	jwtSecretKey = config.JWTSecretKey
	return config, nil
}

func GetJWTSecretKey() string {
	return jwtSecretKey
}
