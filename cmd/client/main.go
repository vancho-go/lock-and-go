package main

import (
	"bufio"
	"fmt"
	"github.com/vancho-go/lock-and-go/internal/config"
	"log"
	"os"
	"strings"
)

var ServerHost string

func main() {
	loaderType := "flag"
	client, err := config.NewClient(loaderType)
	if err != nil {
		log.Fatalf("error building client configuration: %v", err)
	}

	ServerHost = *client.ServerAddress

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Привет от LockAndGo CLI client. Вбей 'help' чтобы увидеть список доступных команд")

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
			Register(args[1], args[2])

		case "login":
			if len(args) != 3 {
				fmt.Println("Пример: login <username> <password>")
				continue
			}
			Login(args[1], args[2])

		case "import-key":
			if len(args) != 3 {
				fmt.Println("Пример: import-key <key path> <key password>")
				continue
			}
			ImportKey(args[1], args[2])

		case "generate-key":
			if len(args) != 3 {
				fmt.Println("Пример: generate-key <key path> <key password>")
				continue
			}
			GenerateKey(args[1], args[2])

		case "data":
			if len(key) != 0 {
				handleData(scanner)
			} else {
				fmt.Println("Для использования импортируйте ключ или создайте новый")
			}

		default:
			fmt.Println("Неизвестная команда:", args[0])
			fmt.Println("Вбей 'help' чтобы увидеть список доступных команд")
		}
	}
}

func handleData(scanner *bufio.Scanner) {
	fmt.Println("Data handling mode. Type 'help' to see available commands.")

	filename := "data.json"
	dataMap := make(map[string]UserData)

	// Попытка чтения существующих данных из файла
	tempDataMap, err := readDataFromFileSecure(filename, key)
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
			printData(dataMap)
		case "edit":
			editDataFromInput(dataMap)
		case "delete":
			deleteDataFromInput(dataMap)
		case "add":
			userData := createDataFromInput()
			if userData.DataID != "" {
				dataMap[userData.DataID] = userData
				fmt.Println("Record added")
			}
		case "sync":
			if AuthToken == "" {
				fmt.Println("Для использования авторизуйтесь")
				break
			}
			if err := syncDataWithServer(dataMap, key, filename); err != nil {
				fmt.Printf("Ошибка при синхронизации данных: %v\n", err)
			} else {
				fmt.Println("Данные успешно синхронизированы с сервером.")
				// Перечитываем данные после синхронизации, чтобы обновить локальное состояние
				dataMap, err = readDataFromFileSecure(filename, key)
				if err != nil {
					fmt.Println("Ошибка при чтении обновленных данных:", err)
				}
			}
		case "save":
			var userDataSlice []UserData
			for _, userData := range dataMap {
				userDataSlice = append(userDataSlice, userData)
			}
			if err := saveDataToFileSecure(userDataSlice, filename, key); err != nil {
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
