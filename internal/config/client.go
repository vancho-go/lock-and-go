package config

import (
	"flag"
	"fmt"
	"os"
)

type Client struct {
	ServerAddress *string
}

type clientLoader interface {
	load() (*Client, error)
}

type flagClientLoader struct{}

func (f *flagClientLoader) load() (*Client, error) {
	serverAddress := flag.String("s", "http://localhost:8080", "address:port of the Server")
	flag.Parse()

	return &Client{

		ServerAddress: serverAddress,
	}, nil
}

type envClientLoader struct{}

func (e *envClientLoader) load() (*Client, error) {
	serverAddress := os.Getenv("SERVER_ADDRESS_FOR_CLIENT")

	return &Client{
		ServerAddress: &serverAddress,
	}, nil
}

func newClientConfigLoader(loaderType string) (clientLoader, error) {
	switch loaderType {
	case "env":
		return &envClientLoader{}, nil
	case "flag":
		return &flagClientLoader{}, nil
	default:
		return nil, fmt.Errorf("unsupported config loader type: %s", loaderType)
	}
}

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
