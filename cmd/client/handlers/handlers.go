package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/vancho-go/lock-and-go/cmd/client/crypto"
	"github.com/vancho-go/lock-and-go/cmd/client/data"
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
	user := model.User{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshalling user: %v", err)
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
	user := model.User{
		Username: username,
		Password: password,
	}
	dataBytes, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("error marshalling user: %v", err)
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
		return "", fmt.Errorf("login successful, but no handlers token received")
	}

	return authToken, nil
}

// SyncDataWithServer синхронизирует локальную версию пользовательских данных с сервером.
func (ac *AuthClient) SyncDataWithServer(dataMap map[string]data.UserData, filename, authToken string, km *crypto.KeyManager) error {
	// Подготовка данных к отправке
	var toSync []model.UserData
	for _, ud := range dataMap {
		if ud.Status != "synced" {
			// Шифрование данных
			encryptedData, err := km.Encrypt(ud.RawData)
			if err != nil {
				return fmt.Errorf("error encrypting data: %v", err)
			}
			base64Data := base64.StdEncoding.EncodeToString(encryptedData)

			toSync = append(toSync, model.UserData{
				DataID:     ud.DataID,
				Data:       base64Data,
				DataType:   ud.DataType,
				Status:     ud.Status,
				CreatedAt:  ud.CreatedAt,
				ModifiedAt: ud.ModifiedAt,
			})
		}
	}

	// Отправка собранных данных на сервер одним POST-запросом
	if len(toSync) > 0 {
		dataBytes, err := json.Marshal(toSync)
		if err != nil {
			return fmt.Errorf("error marshalling data: %v", err)
		}

		req, err := http.NewRequest("POST", ac.serverHost+"/data/sync", bytes.NewBuffer(dataBytes))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", "AuthToken="+authToken) // Добавляем куки

		response, err := ac.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned non-OK status: %d", response.StatusCode)
		}
	}

	// Получение актуальных данных с сервера
	req, err := http.NewRequest("GET", ac.serverHost+"/data", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Cookie", "AuthToken="+authToken) // Добавляем куки

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var serverData []model.UserData
	if err := json.Unmarshal(body, &serverData); err != nil {
		return fmt.Errorf("error unmarshalling server data: %v", err)
	}

	// Обновление локальных данных и сохранение в файл
	newDataMap := make(map[string]data.UserData)
	for _, sd := range serverData {
		if sd.Data == "" {
			// Обработка случая, когда строка данных пуста
			fmt.Println("Предупреждение: попытка декодировать пустую строку.")
			continue // Пропускаем текущую итерацию цикла
		}

		decodedData, errDec := base64.StdEncoding.DecodeString(sd.Data)
		if errDec != nil {
			// Обработка ошибки некорректного формата Base64
			return fmt.Errorf("error decoding data from base64: %v", errDec)
		}

		decryptedData, errDecr := km.Decrypt(decodedData)
		if errDecr != nil {
			return fmt.Errorf("error decrypting data: %v", errDecr)
		}

		// Обновление статуса на "synced"
		sd.Status = "synced"

		newDataMap[sd.DataID] = data.UserData{
			DataID:     sd.DataID,
			RawData:    decryptedData,
			DataType:   sd.DataType,
			Status:     sd.Status,
			CreatedAt:  sd.CreatedAt,
			ModifiedAt: sd.ModifiedAt,
		}
	}

	// Сохранение обновленных данных в файл с шифрованием
	newDataList := make([]data.UserData, 0, len(newDataMap))
	for _, v := range newDataMap {
		newDataList = append(newDataList, v)
	}

	return data.SaveDataToFileSecure(newDataList, filename, km)
}
