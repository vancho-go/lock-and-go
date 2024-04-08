package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"os"
)

var key []byte

// Функция для генерации ключа шифрования данных и его шифрования с использованием ключа, полученного из пароля.
func generateAndSaveKey(password, path string) error {
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
	return os.WriteFile(path, keyFileContent, 0600)
}

// Функция для считывания и расшифровки ключа шифрования данных из файла с использованием пароля.
func readAndDecryptKey(password, keyPath string) ([]byte, error) {
	keyFileContent, err := os.ReadFile(keyPath)
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
	if len(encrypted) < nonceSize {
		return []byte{}, errors.New("error decrypting data file, it may be empty")
	}
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func ImportKey(keyPath, keyPass string) {
	decryptedKey, err := readAndDecryptKey(keyPass, keyPath)
	if err != nil {
		fmt.Println("Failed to decrypt your encryption key:", err)
		return
	}
	key = decryptedKey
	fmt.Println("Encryption key imported successfully.")
}

func GenerateKey(keyPath, keyPass string) {
	if err := generateAndSaveKey(keyPass, keyPath); err != nil {
		fmt.Println("Failed to generate and save encryption key:", err)
		return
	}
	fmt.Println("Encryption key generated and saved successfully.")
}
