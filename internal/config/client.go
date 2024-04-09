package config

import (
	"flag"
	"fmt"
	"os"
)

// Client конфигурация для клиента.
type Client struct {
	// ServerAddress адрес сервера, на который должен обращаться клиент
	// для синхронизации данных.
	ServerAddress *string
}

// FlagSetter методы для работы с флагами.
type FlagSetter interface {
	// StringVar считывает string флаги.
	StringVar(p *string, name string, value string, usage string)
	// Parse парсит считанные флаги.
	Parse()
}

// EnvReader методы для работы с переменными окружения.
type EnvReader interface {
	// GetEnv считывает env переменные.
	GetEnv(key string) string
}

// RealFlagSetter имплементация FlagSetter.
type RealFlagSetter struct{}

// StringVar считывает string флаги.
func (r RealFlagSetter) StringVar(p *string, name string, value string, usage string) {
	flag.StringVar(p, name, value, usage)
}

// Parse парсит считанные флаги.
func (r RealFlagSetter) Parse() {
	flag.Parse()
}

// RealEnvReader имплементация EnvReader.
type RealEnvReader struct{}

// GetEnv считывает env переменные.
func (r RealEnvReader) GetEnv(key string) string {
	return os.Getenv(key)
}

// clientLoader методы, которые должны быть реализованы
// различными загрузчиками конфигураций клиента.
type clientLoader interface {
	// load собирает необходимые переменные.
	load() (*Client, error)
}

// flagClientLoader имплементация clientLoader.
type flagClientLoader struct {
	flagSetter FlagSetter
}

// load собирает необходимые переменные.
func (f *flagClientLoader) load() (*Client, error) {
	var serverAddress string
	f.flagSetter.StringVar(&serverAddress, "s", "http://localhost:8080", "address:port of the Server")
	f.flagSetter.Parse()

	return &Client{
		ServerAddress: &serverAddress,
	}, nil
}

// envClientLoader имплементация clientLoader.
type envClientLoader struct {
	envReader EnvReader
}

// load собирает необходимые переменные.
func (e *envClientLoader) load() (*Client, error) {
	serverAddress := e.envReader.GetEnv("SERVER_ADDRESS_FOR_CLIENT")

	return &Client{
		ServerAddress: &serverAddress,
	}, nil
}

// newClientConfigLoader переключает загрузчики в зависимости параметра.
func newClientConfigLoader(loaderType string) (clientLoader, error) {
	switch loaderType {
	case "env":
		return &envClientLoader{envReader: RealEnvReader{}}, nil
	case "flag":
		return &flagClientLoader{flagSetter: RealFlagSetter{}}, nil
	default:
		return nil, fmt.Errorf("unsupported config loader type: %s", loaderType)
	}
}

// NewClient генерирует конфигурацию клиента.
func NewClient(loaderType string) (*Client, error) {
	loader, err := newClientConfigLoader(loaderType)
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
