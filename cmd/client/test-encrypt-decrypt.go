package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/argon2"
	"os"
)

// Функция для генерации ключа шифрования данных и его шифрования с использованием ключа, полученного из пароля.
func generateAndSaveKey(password string) error {
	// Генерация ключа шифрования данных
	dataKey := make([]byte, 32) // AES-256 ключ
	if _, err := rand.Read(dataKey); err != nil {
		return err
	}

	// Генерация ключа разблокировки из пароля
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	unlockKey := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Шифрование ключа шифрования данных с использованием ключа разблокировки
	block, err := aes.NewCipher(unlockKey)
	if err != nil {
		return err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	encryptedDataKey := aesGCM.Seal(nonce, nonce, dataKey, nil)

	// Сохранение зашифрованного ключа и соли в файл
	keyFileContent := append(salt, encryptedDataKey...)
	return os.WriteFile("encryption.key", keyFileContent, 0600)
}

// Функция для считывания и расшифровки ключа шифрования данных из файла с использованием пароля.
func readAndDecryptKey(password string) ([]byte, error) {
	keyFileContent, err := os.ReadFile("encryption.key")
	if err != nil {
		return nil, err
	}

	salt, encryptedDataKey := keyFileContent[:16], keyFileContent[16:]
	unlockKey := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(unlockKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := encryptedDataKey[:nonceSize], encryptedDataKey[nonceSize:]
	dataKey, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return dataKey, nil
}

// Функция для шифрования данных.
func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	encrypted := aesGCM.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// Функция для расшифровки данных.
func decrypt(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func main() {
	password := "секретный пароль"
	if err := generateAndSaveKey(password); err != nil {
		fmt.Println("Ошибка при генерации ключа:", err)
		return
	}

	key, err := readAndDecryptKey(password)
	if err != nil {
		fmt.Println("Ошибка при считывании ключа:", err)
		return
	}

	text := "Тестовое сообщение для шифрования"
	encrypted, err := encrypt([]byte(text), key)
	if err != nil {
		fmt.Println("Ошибка при шифровании:", err)
		return
	}
	fmt.Println("Зашифрованный текст:", hex.EncodeToString(encrypted))

	decrypted, err := decrypt(encrypted, key)
	if err != nil {
		fmt.Println("Ошибка при расшифровке:", err)
		return
	}
	fmt.Println("Расшифрованный текст:", string(decrypted))
}
