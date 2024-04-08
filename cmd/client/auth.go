package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	api "github.com/vancho-go/lock-and-go/api/http"
	"io"
	"net/http"
)

var AuthToken string

func Register(username, password string) {
	data := api.RegisterUserRequest{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return
	}

	resp, err := http.Post((ServerHost + "/register"), "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to register: %s\n", body)
		return
	}

	fmt.Println("Registration successful.")
}

func Login(username, password string) {
	data := api.AuthenticateUserRequest{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return
	}

	resp, err := http.Post(ServerHost+"/login", "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to login: %s\n", body)
		return
	}

	// Предполагаем, что сервер устанавливает cookie с именем AuthToken
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "AuthToken" {
			AuthToken = cookie.Value
			break
		}
	}

	fmt.Println("Login successful.")
}
