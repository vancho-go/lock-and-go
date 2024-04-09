package config

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// JWTConfig структура для хранения конфигурации JWT.
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// jwtConfigInstance хранит единственный экземпляр конфигурации JWT для всего приложения.
var jwtConfigInstance *JWTConfig

// once используется для реализации паттерна "одиночка".
var once sync.Once

// GetJWTConfig функция для получения единственного экземпляра JWTConfig.
func GetJWTConfig() *JWTConfig {
	once.Do(func() {
		jwtConfigInstance = &JWTConfig{}
	})
	return jwtConfigInstance
}

// initializeJWTConfig инициализирует jwt конфиг.
func initializeJWTConfig(secretKey string, tokenDurationStr string) error {
	jwtTokenDuration, err := time.ParseDuration(tokenDurationStr)
	if err != nil {
		return fmt.Errorf("failed to decode jwt duration: %v", err)
	}

	jwtConfig := GetJWTConfig()
	jwtConfig.SecretKey = secretKey
	jwtConfig.TokenDuration = jwtTokenDuration

	return nil
}

// Server конфигурация для сервера.
type Server struct {
	// Address адрес сервера.
	Address string
	// DatabaseURI строка для подключения к БД.
	DatabaseURI string
	// LogLevel уровень логирования.
	LogLevel string
}

// serverLoader методы, которые должны быть реализованы
// различными загрузчиками конфигураций сервера.
type serverLoader interface {
	load() (*Server, error)
}

// flagServerLoader имплементация serverLoader.
type flagServerLoader struct {
	flagSetter FlagSetter
}

// load собирает необходимые переменные.
func (f *flagServerLoader) load() (*Server, error) {
	var serverAddress, databaseURI, logLevel, jwtSecretKey, jwtTokenDurationStr string

	f.flagSetter.StringVar(&serverAddress, "a", "", "address:port to run Server")
	f.flagSetter.StringVar(&databaseURI, "d", "", "connection string for driver to establish connection to the conn")
	f.flagSetter.StringVar(&logLevel, "l", "", "logger level")
	f.flagSetter.StringVar(&jwtSecretKey, "js", "", "jwt secret key")
	f.flagSetter.StringVar(&jwtTokenDurationStr, "jt", "", "jwt secret key valid time")

	f.flagSetter.Parse()

	if err := initializeJWTConfig(jwtSecretKey, jwtTokenDurationStr); err != nil {
		return nil, err
	}

	return &Server{
		Address:     serverAddress,
		DatabaseURI: databaseURI,
		LogLevel:    logLevel,
	}, nil
}

// envServerLoader имплементация serverLoader.
type envServerLoader struct {
	envReader EnvReader
}

// load собирает необходимые переменные.
func (e *envServerLoader) load() (*Server, error) {
	jwtSecretKey := e.envReader.GetEnv("JWT_SECRET_KEY")
	jwtTokenDuration := e.envReader.GetEnv("JWT_TOKEN_DURATION")

	if err := initializeJWTConfig(jwtSecretKey, jwtTokenDuration); err != nil {
		return nil, err
	}

	return &Server{
		Address:     e.envReader.GetEnv("SERVER_ADDRESS"),
		DatabaseURI: e.envReader.GetEnv("DATABASE_URI"),
		LogLevel:    e.envReader.GetEnv("LOG_LEVEL"),
	}, nil
}

// newServerConfigLoader переключает загрузчики в зависимости параметра.
func newServerConfigLoader(loaderType string) (serverLoader, error) {
	switch loaderType {
	case "env":
		return &envServerLoader{envReader: RealEnvReader{}}, nil
	case "flag":
		return &flagServerLoader{flagSetter: RealFlagSetter{}}, nil
	default:
		return nil, fmt.Errorf("unsupported config loader type: %s", loaderType)
	}
}

// isConfigFull проверяет полноту конфигурации.
func isConfigFull(config any) error {
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
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

// NewServer генерирует конфигурацию сервера.
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

// GetJWTSecretKey возвращает JWT секретный ключ из единственного экземпляра JWTConfig.
func GetJWTSecretKey() string {
	return GetJWTConfig().SecretKey
}

// GetJWTTokenDuration возвращает продолжительность действия JWT токена из единственного экземпляра JWTConfig.
func GetJWTTokenDuration() time.Duration {
	return GetJWTConfig().TokenDuration
}
