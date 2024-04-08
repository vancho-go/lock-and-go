package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"
)

var (
	jwtSecretKey     string
	jwtTokenDuration time.Duration
)

type Server struct {
	Address     string
	DatabaseURI string
	LogLevel    string
}

type serverLoader interface {
	load() (*Server, error)
}

type flagServerLoader struct{}

func (f *flagServerLoader) load() (*Server, error) {
	serverAddress := flag.String("a", "", "address:port to run Server")
	databaseURI := flag.String("d", "", "connection string for driver to establish connection to the conn")
	logLevel := flag.String("l", "", "logger level")
	jwtSecretKey = *flag.String("js", "", "jwt secret key")
	jwtTokenDurationStr := flag.String("jt", "", "jwt secret key valid time")

	flag.Parse()

	jwtTokenDurationParsed, err := time.ParseDuration(*jwtTokenDurationStr)
	if err != nil {
		return nil, fmt.Errorf("load: failed to decode jwt duration: %v", err)
	}
	jwtTokenDuration = jwtTokenDurationParsed
	return &Server{
		Address:     *serverAddress,
		DatabaseURI: *databaseURI,
		LogLevel:    *logLevel,
	}, nil
}

type envServerLoader struct{}

func (e *envServerLoader) load() (*Server, error) {
	jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
	jwtTokenDurationParsed, err := time.ParseDuration(os.Getenv("JWT_TOKEN_DURATION"))
	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}
	jwtTokenDuration = jwtTokenDurationParsed
	return &Server{
		Address:     os.Getenv("SERVER_ADDRESS"),
		DatabaseURI: os.Getenv("DATABASE_URI"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
	}, nil
}

func newServerConfigLoader(loaderType string) (serverLoader, error) {
	switch loaderType {
	case "env":
		return &envServerLoader{}, nil
	case "flag":
		return &flagServerLoader{}, nil
	default:
		return nil, fmt.Errorf("unsupported config loader type: %s", loaderType)
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
			return fmt.Errorf("%s field should not be empty", field.Name)
		}
	}
	return nil
}

func NewServer(loaderType string) (*Server, error) {
	loader, err := newServerConfigLoader(loaderType)
	if err != nil {
		return nil, err
	}

	config, err := loader.load()
	if err != nil {
		return nil, err
	}

	err = isConfigFull(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func GetJWTSecretKey() string {
	return jwtSecretKey
}

func GetJWTTokenDuration() time.Duration {
	return jwtTokenDuration
}
