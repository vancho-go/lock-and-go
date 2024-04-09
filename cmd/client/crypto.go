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

// KeyManager менеджер ключей шифрования.
type KeyManager struct {
	Key []byte
}

// NewKeyManager конструктор KeyManager.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// ImportKey импортирует существующий ключ пользователя.
func (km *KeyManager) ImportKey(keyPath, keyPass string) error {
	decryptedKey, err := readAndDecryptKey(keyPass, keyPath)
	if err != nil {
		return fmt.Errorf("failed to decrypt your encryption key: %w", err)
	}
	km.Key = decryptedKey
	return nil
}

// GenerateKey генерирует новый ключ шифрования для пользователя.
func (km *KeyManager) GenerateKey(keyPath, keyPass string) error {
	if err := generateAndSaveKey(keyPass, keyPath); err != nil {
		return fmt.Errorf("failed to generate and save encryption key: %w", err)
	}
	return nil
}

// Encrypt шифрует данные пользователя.
func (km *KeyManager) Encrypt(data []byte) ([]byte, error) {
	return encrypt(data, km.Key)
}

// Decrypt расшифровывает данные пользователя.
func (km *KeyManager) Decrypt(encrypted []byte) ([]byte, error) {
	return decrypt(encrypted, km.Key)
}

// generateAndSaveKey функция для генерации ключа шифрования данных и его шифрования с использованием ключа, полученного из пароля.
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

// readAndDecryptKey функция для считывания и расшифровки ключа шифрования данных из файла с использованием пароля.
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

// encrypt функция для шифрования данных.
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

// decrypt функция для расшифровки данных.
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
