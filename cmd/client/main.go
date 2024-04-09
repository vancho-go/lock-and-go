package main

import (
	"bufio"
	"fmt"
	"github.com/vancho-go/lock-and-go/cmd/client/crypto"
	"github.com/vancho-go/lock-and-go/cmd/client/data"
	"github.com/vancho-go/lock-and-go/cmd/client/handlers"
	"github.com/vancho-go/lock-and-go/internal/config"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildInfo()

	var AuthToken string

	loaderType := "flag"
	client, err := config.NewClient(loaderType)
	if err != nil {
		log.Fatalf("error building client_macOS configuration: %v", err)
	}

	httpClient := &http.Client{}
	authClient := handlers.NewAuthClient(httpClient, *client.ServerAddress)

	keyManager := crypto.NewKeyManager()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Привет от LockAndGo CLI client_macOS. Вбей 'help' чтобы увидеть список доступных команд")

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		args := strings.Fields(input)

		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "exit":
			fmt.Println("Завершение работы ...")
			return
		case "help":
			fmt.Println("Доступные команды:")
			fmt.Println("  register <username> <password> - Зарегистрировать нового пользователя")
			fmt.Println("  login <username> <password> - Войти")
			fmt.Println("  import-key <key path> <key password> - Импортировать свой ключ шифрования")
			fmt.Println("  generate-key <key path> <key password> - Сгенерировать новый ключ шифрования")
			fmt.Println("  data - Войти в режим работы с данными")
			fmt.Println("  exit - Выйти")
		case "register":
			if len(args) != 3 {
				fmt.Println("Пример: register <username> <password>")
				continue
			}
			err = authClient.Register(args[1], args[2])
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Registration successful.")
			}

		case "login":
			if len(args) != 3 {
				fmt.Println("Пример: login <username> <password>")
				continue
			}
			AuthToken, err = authClient.Login(args[1], args[2])
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Login successful.")
			}

		case "import-key":
			if len(args) != 3 {
				fmt.Println("Пример: import-key <key path> <key password>")
				continue
			}
			if err = keyManager.ImportKey(args[1], args[2]); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Encryption key imported successfully.")
			}

		case "generate-key":
			if len(args) != 3 {
				fmt.Println("Пример: generate-key <key path> <key password>")
				continue
			}
			if err = keyManager.GenerateKey(args[1], args[2]); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Encryption key generated and saved successfully.")
			}

		case "data":
			if len(keyManager.Key) != 0 {
				handleData(scanner, AuthToken, authClient, keyManager)
			} else {
				fmt.Println("Для использования импортируйте ключ или создайте новый")
			}

		default:
			fmt.Println("Неизвестная команда:", args[0])
			fmt.Println("Вбей 'help' чтобы увидеть список доступных команд")
		}
	}
}

func handleData(scanner *bufio.Scanner, authToken string, ac *handlers.AuthClient, km *crypto.KeyManager) {
	fmt.Println("Data handling mode. Type 'help' to see available commands.")

	filename := "data.json"
	dataMap := make(map[string]data.UserData)

	// Попытка чтения существующих данных из файла
	tempDataMap, err := data.ReadDataFromFileSecure(filename, km)
	if err != nil {
		fmt.Println("Failed to read the file, we start with an empty database.")
	} else {
		dataMap = tempDataMap
	}

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		args := strings.Fields(input)

		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "help":
			fmt.Println("Доступные команды:")
			fmt.Println("  add - Добавить новую запись")
			fmt.Println("  edit - Редактировать запись")
			fmt.Println("  delete - Удалить запись")
			fmt.Println("  show - Показать все записи")
			fmt.Println("  sync - Синхронизировать записи с сервером")
			fmt.Println("  save - Сохранить внесенные изменения")
			fmt.Println("  back - Вернуться в главное меню")
		case "show":
			data.PrintData(dataMap)
		case "edit":
			data.EditDataFromInput(dataMap)
		case "delete":
			data.DeleteDataFromInput(dataMap)
		case "add":
			userData := data.CreateDataFromInput()
			if userData.DataID != "" {
				dataMap[userData.DataID] = userData
				fmt.Println("Record added")
			}
		case "sync":
			if authToken == "" {
				fmt.Println("Для использования авторизуйтесь")
				break
			}
			if err := ac.SyncDataWithServer(dataMap, filename, authToken, km); err != nil {
				fmt.Printf("Ошибка при синхронизации данных: %v\n", err)
			} else {
				fmt.Println("Данные успешно синхронизированы с сервером.")
				// Перечитываем данные после синхронизации, чтобы обновить локальное состояние
				dataMap, err = data.ReadDataFromFileSecure(filename, km)
				if err != nil {
					fmt.Println("Ошибка при чтении обновленных данных:", err)
				}
			}
		case "save":
			var userDataSlice []data.UserData
			for _, userData := range dataMap {
				userDataSlice = append(userDataSlice, userData)
			}
			if err := data.SaveDataToFileSecure(userDataSlice, filename, km); err != nil {
				fmt.Println("Ошибка при сохранении данных:", err)
			} else {
				fmt.Println("Данные успешно сохранены.")
			}
		case "back":
			fmt.Println("Возвращаемся в главное меню ...")
			return
		default:
			fmt.Println("Неизвестная команда:", args[0])
		}
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
