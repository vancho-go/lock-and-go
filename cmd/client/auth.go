package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vancho-go/lock-and-go/internal/model"
	"io"
	"net/http"
)

// HttpClient методы для http клиента
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AuthClient клиент для аутентификации.
type AuthClient struct {
	httpClient HttpClient
	serverHost string
}

// NewAuthClient конструктор AuthClient.
func NewAuthClient(httpClient HttpClient, serverHost string) *AuthClient {
	return &AuthClient{
		httpClient: httpClient,
		serverHost: serverHost,
	}
}

// Register метод регистрации пользователя.
func (ac *AuthClient) Register(username, password string) error {
	data := model.User{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}
	req, err := http.NewRequest("POST", ac.serverHost+"/register", bytes.NewBuffer(dataBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to register: %s", body)
	}
	return nil
}

// Login метод авторизации пользователя.
func (ac *AuthClient) Login(username, password string) (string, error) {
	data := model.User{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshalling data: %v", err)
	}

	req, err := http.NewRequest("POST", ac.serverHost+"/login", bytes.NewBuffer(dataBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to login: %s", body)
	}

	var authToken string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "AuthToken" {
			authToken = cookie.Value
			break
		}
	}

	if authToken == "" {
		return "", fmt.Errorf("login successful, but no auth token received")
	}

	return authToken, nil
}
