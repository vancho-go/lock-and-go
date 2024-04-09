package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/vancho-go/lock-and-go/internal/model"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

// DataInterface определяет методы, которые должны быть реализованы типами данных
type DataInterface interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// LoginPasswordData представляет данные логина и пароля.
type LoginPasswordData struct {
	Login    string
	Password string
	MetaInfo string
}

// Serialize сериализатор LoginPasswordData.
func (lpd *LoginPasswordData) Serialize() ([]byte, error) {
	return json.Marshal(lpd)
}

// Deserialize десериализатор LoginPasswordData.
func (lpd *LoginPasswordData) Deserialize(data []byte) error {
	return json.Unmarshal(data, lpd)
}

// TextData текстовый тип хранимых данных.
type TextData struct {
	Text     string
	MetaInfo string
}

// Serialize сериализатор TextData.
func (td *TextData) Serialize() ([]byte, error) {
	return json.Marshal(td)
}

// Deserialize десериализатор TextData.
func (td *TextData) Deserialize(data []byte) error {
	return json.Unmarshal(data, td)
}

// BankCardData тип данных для банковских карт.
type BankCardData struct {
	CardNumber string
	ExpiryDate string
	CVV        string
	MetaInfo   string
}

// Serialize сериализатор BankCardData.
func (bcd *BankCardData) Serialize() ([]byte, error) {
	return json.Marshal(bcd)
}

// Deserialize десериализатор BankCardData.
func (bcd *BankCardData) Deserialize(data []byte) error {
	return json.Unmarshal(data, bcd)
}

// BinaryData тип для бинарных данных.
type BinaryData struct {
	Data     []byte
	MetaInfo string
}

// Serialize сериализатор BinaryData.
func (bd *BinaryData) Serialize() ([]byte, error) {
	// Преобразование бинарных данных в Base64 строку для сериализации
	encodedData := base64.StdEncoding.EncodeToString(bd.Data)
	// Создание копии с Base64 строкой для сериализации
	bdCopy := &BinaryData{
		Data:     []byte(encodedData),
		MetaInfo: bd.MetaInfo,
	}
	return json.Marshal(bdCopy)
}

// Deserialize десериализатор BinaryData.
func (bd *BinaryData) Deserialize(data []byte) error {
	var bdCopy BinaryData
	if err := json.Unmarshal(data, &bdCopy); err != nil {
		return err
	}
	// Декодирование Base64 строки обратно в бинарные данные
	decodedData, err := base64.StdEncoding.DecodeString(string(bdCopy.Data))
	if err != nil {
		return err
	}
	bd.Data = decodedData
	bd.MetaInfo = bdCopy.MetaInfo
	return nil
}

// UserData содержит данные пользователя, включая динамические данные
type UserData struct {
	DataID     string          `json:"data_id"`
	RawData    json.RawMessage `json:"data"`
	DataType   string          `json:"data_type"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ModifiedAt time.Time       `json:"modified_at"`
}

// saveDataToFileSecure сохраняет данные пользователя в локальный файл.
func saveDataToFileSecure(data []UserData, filename string, km *KeyManager) error {
	// Клонирование данных для безопасного шифрования
	dataCopy := make([]UserData, len(data))
	for i, d := range data {
		encryptedData, err := km.Encrypt(d.RawData)
		if err != nil {
			return err
		}
		base64EncryptedData := base64.StdEncoding.EncodeToString(encryptedData)
		// Преобразование зашифрованных данных в строку JSON
		dataCopy[i] = UserData{
			DataID:     d.DataID,
			DataType:   d.DataType,
			Status:     d.Status,
			CreatedAt:  d.CreatedAt,
			ModifiedAt: d.ModifiedAt,
			// Оборачиваем Base64-строку в кавычки для корректной сериализации в JSON
			RawData: json.RawMessage(fmt.Sprintf("%q", base64EncryptedData)),
		}
	}

	jsonData, err := json.MarshalIndent(dataCopy, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// readDataFromFileSecure считывает данные пользователя из локального файла.
func readDataFromFileSecure(filename string, km *KeyManager) (map[string]UserData, error) {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dataList []UserData
	if err := json.Unmarshal(jsonData, &dataList); err != nil {
		return nil, err
	}

	for i, d := range dataList {
		// Преобразование из JSON-строки в обычную строку
		base64EncryptedData := d.RawData[1 : len(d.RawData)-1] // Удаление кавычек из JSON-строки
		encryptedData, err := base64.StdEncoding.DecodeString(string(base64EncryptedData))
		if err != nil {
			return nil, err
		}
		decryptedData, err := km.Decrypt(encryptedData)
		if err != nil {
			return nil, err
		}
		dataList[i].RawData = decryptedData
	}

	dataMap := make(map[string]UserData)
	for _, userData := range dataList {
		dataMap[userData.DataID] = userData
	}

	return dataMap, nil
}

// printData выводит данные пользователя из локального файла.
func printData(dataMap map[string]UserData) {
	for id, userData := range dataMap {
		fmt.Printf("Data ID: %s\n", id)
		fmt.Printf("Data Type: %s\n", userData.DataType)
		fmt.Printf("Status: %s\n", userData.Status)
		fmt.Printf("Created At: %s\n", userData.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Modified At: %s\n", userData.ModifiedAt.Format(time.RFC3339))

		// Десериализация RawData для вывода содержимого
		var tempData interface{}
		if err := json.Unmarshal(userData.RawData, &tempData); err != nil {
			fmt.Println("Error unmarshalling RawData:", err)
			continue
		}
		formattedData, err := json.MarshalIndent(tempData, "", "    ")
		if err != nil {
			fmt.Println("Error formatting data:", err)
			continue
		}
		fmt.Printf("Data Content:\n%s\n\n", string(formattedData))
	}
}

// createDataFromInput создает новые данные разных типов.
func createDataFromInput() UserData {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Выберите тип данных для добавления:")
	fmt.Println("1: LoginPasswordData")
	fmt.Println("2: TextData")
	fmt.Println("3: BankCardData")
	fmt.Println("4: BinaryData")

	dataType, _ := reader.ReadString('\n')
	dataType = strings.TrimSpace(dataType)

	var data DataInterface
	var dataID string

	switch dataType {
	case "1":
		fmt.Println("Введите login:")
		login, _ := reader.ReadString('\n')
		fmt.Println("Введите password:")
		password, _ := reader.ReadString('\n')
		fmt.Println("Введите MetaInfo:")
		metaInfo, _ := reader.ReadString('\n')

		data = &LoginPasswordData{
			Login:    strings.TrimSpace(login),
			Password: strings.TrimSpace(password),
			MetaInfo: strings.TrimSpace(metaInfo),
		}
	case "2":
		fmt.Println("Введите Text:")
		text, _ := reader.ReadString('\n')
		fmt.Println("Введите MetaInfo:")
		metaInfo, _ := reader.ReadString('\n')

		data = &TextData{
			Text:     strings.TrimSpace(text),
			MetaInfo: strings.TrimSpace(metaInfo),
		}
	case "3":
		fmt.Println("Введите CardNumber:")
		cardNumber, _ := reader.ReadString('\n')
		fmt.Println("Введите ExpiryDate:")
		expiryDate, _ := reader.ReadString('\n')
		fmt.Println("Введите CVV:")
		cvv, _ := reader.ReadString('\n')
		fmt.Println("Введите MetaInfo:")
		metaInfo, _ := reader.ReadString('\n')

		data = &BankCardData{
			CardNumber: strings.TrimSpace(cardNumber),
			ExpiryDate: strings.TrimSpace(expiryDate),
			CVV:        strings.TrimSpace(cvv),
			MetaInfo:   strings.TrimSpace(metaInfo),
		}
	case "4":
		fmt.Println("Введите Data (будет закодировано в Base64):")
		rawData, _ := reader.ReadString('\n')
		fmt.Println("Введите MetaInfo:")
		metaInfo, _ := reader.ReadString('\n')

		encodedData := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(rawData)))

		data = &BinaryData{
			Data:     []byte(encodedData),
			MetaInfo: strings.TrimSpace(metaInfo),
		}
	default:
		fmt.Println("Неверный выбор, возвращаем пустую запись.")
		return UserData{}
	}

	dataID = uuid.NewString() // Генерируем уникальный ID для новых данных
	raw, _ := data.Serialize()

	currentTime := time.Now()
	return UserData{
		DataID:     dataID,
		RawData:    raw,
		DataType:   reflect.TypeOf(data).Elem().Name(),
		Status:     "created",
		CreatedAt:  currentTime,
		ModifiedAt: currentTime,
	}
}

// editDataFromInput позволяет отредактировать локльные пользовательские данные.
func editDataFromInput(dataMap map[string]UserData) {
	fmt.Println("Введите DataID для редактирования:")
	reader := bufio.NewReader(os.Stdin)
	dataID, _ := reader.ReadString('\n')
	dataID = strings.TrimSpace(dataID)

	userData, exists := dataMap[dataID]
	if !exists {
		fmt.Println("Запись не найдена.")
		return
	}

	switch userData.DataType {
	case "LoginPasswordData":
		var data LoginPasswordData
		json.Unmarshal(userData.RawData, &data)

		fmt.Println("Текущий login:", data.Login)
		fmt.Println("Введите новый login:")
		login, _ := reader.ReadString('\n')
		data.Login = strings.TrimSpace(login)

		fmt.Println("Текущий password:", data.Password)
		fmt.Println("Введите новый password:")
		password, _ := reader.ReadString('\n')
		data.Password = strings.TrimSpace(password)
		userData.RawData, _ = json.Marshal(data)
	case "TextData":
		var data TextData
		if err := json.Unmarshal(userData.RawData, &data); err != nil {
			fmt.Println("Ошибка при десериализации данных:", err)
			return
		}
		fmt.Println("Введите новый Text: ")
		data.Text, _ = reader.ReadString('\n')
		data.Text = strings.TrimSpace(data.Text)
		fmt.Printf("Текущий MetaInfo: %s\nВведите новый MetaInfo: ", data.MetaInfo)
		data.MetaInfo, _ = reader.ReadString('\n')
		data.MetaInfo = strings.TrimSpace(data.MetaInfo)
		userData.RawData, _ = json.Marshal(data)
	case "BankCardData":
		var data BankCardData
		if err := json.Unmarshal(userData.RawData, &data); err != nil {
			fmt.Println("Ошибка при десериализации данных:", err)
			return
		}
		fmt.Printf("Текущий CardNumber: %s\nВведите новый CardNumber: ", data.CardNumber)
		data.CardNumber, _ = reader.ReadString('\n')
		data.CardNumber = strings.TrimSpace(data.CardNumber)
		fmt.Printf("Текущий ExpiryDate: %s\nВведите новый ExpiryDate: ", data.ExpiryDate)
		data.ExpiryDate, _ = reader.ReadString('\n')
		data.ExpiryDate = strings.TrimSpace(data.ExpiryDate)
		fmt.Printf("Текущий CVV: %s\nВведите новый CVV: ", data.CVV)
		data.CVV, _ = reader.ReadString('\n')
		data.CVV = strings.TrimSpace(data.CVV)
		fmt.Printf("Текущий MetaInfo: %s\nВведите новый MetaInfo: ", data.MetaInfo)
		data.MetaInfo, _ = reader.ReadString('\n')
		data.MetaInfo = strings.TrimSpace(data.MetaInfo)
		userData.RawData, _ = json.Marshal(data)
	case "BinaryData":
		var data BinaryData
		if err := json.Unmarshal(userData.RawData, &data); err != nil {
			fmt.Println("Ошибка при десериализации данных:", err)
			return
		}
		fmt.Println("Введите новые данные в формате строки:")
		rawData, _ := reader.ReadString('\n')
		rawData = strings.TrimSpace(rawData)
		data.Data = []byte(base64.StdEncoding.EncodeToString([]byte(rawData)))
		fmt.Printf("Текущий MetaInfo: %s\nВведите новый MetaInfo: ", data.MetaInfo)
		data.MetaInfo, _ = reader.ReadString('\n')
		data.MetaInfo = strings.TrimSpace(data.MetaInfo)
		userData.RawData, _ = json.Marshal(data)
	}

	userData.ModifiedAt = time.Now()
	userData.Status = "modified"
	dataMap[dataID] = userData
	fmt.Println("Данные успешно обновлены.")
}

// deleteDataFromInput удаляет пользовательские данные.
func deleteDataFromInput(dataMap map[string]UserData) {
	fmt.Println("Введите DataID для удаления:")
	reader := bufio.NewReader(os.Stdin)
	dataID, _ := reader.ReadString('\n')
	dataID = strings.TrimSpace(dataID)

	userData, exists := dataMap[dataID]
	if !exists {
		fmt.Println("Запись не найдена.")
		return
	}

	userData.ModifiedAt = time.Now()
	userData.Status = "deleted"

	dataMap[dataID] = userData
	fmt.Println("Запись помечена как удалённая.")
}

// syncDataWithServer синхронизирует локальную версию пользовательских данных с сервером.
func (ac *AuthClient) syncDataWithServer(dataMap map[string]UserData, filename, authToken string, km *KeyManager) error {
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
	newDataMap := make(map[string]UserData)
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

		newDataMap[sd.DataID] = UserData{
			DataID:     sd.DataID,
			RawData:    decryptedData,
			DataType:   sd.DataType,
			Status:     sd.Status,
			CreatedAt:  sd.CreatedAt,
			ModifiedAt: sd.ModifiedAt,
		}
	}

	// Сохранение обновленных данных в файл с шифрованием
	newDataList := make([]UserData, 0, len(newDataMap))
	for _, v := range newDataMap {
		newDataList = append(newDataList, v)
	}

	return saveDataToFileSecure(newDataList, filename, km)
}
