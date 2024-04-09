package data

import (
	"reflect"
	"testing"
)

func TestBankCardData_Serialize_Deserialize(t *testing.T) {
	original := &BankCardData{
		CardNumber: "1234567890123456",
		ExpiryDate: "06/23",
		CVV:        "123",
		MetaInfo:   "Test Meta Info",
	}
	serialized, err := original.Serialize()
	if err != nil {
		t.Errorf("Serialize() error = %v", err)
	}

	deserialized := &BankCardData{}
	err = deserialized.Deserialize(serialized)
	if err != nil {
		t.Errorf("Deserialize() error = %v", err)
	}

	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Original and deserialized objects are not equal. Original: %v, Deserialized: %v", original, deserialized)
	}
}

func TestTextData_Serialize_Deserialize(t *testing.T) {
	original := &TextData{
		Text:     "Sample text data",
		MetaInfo: "Sample meta info",
	}
	serialized, err := original.Serialize()
	if err != nil {
		t.Errorf("Serialize() error = %v", err)
	}

	deserialized := &TextData{}
	err = deserialized.Deserialize(serialized)
	if err != nil {
		t.Errorf("Deserialize() error = %v", err)
	}

	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Original and deserialized objects are not equal. Original: %v, Deserialized: %v", original, deserialized)
	}
}

func TestBinaryData_Serialize_Deserialize(t *testing.T) {
	originalData := "This is a test"
	original := &BinaryData{
		Data:     []byte(originalData),
		MetaInfo: "Binary data meta info",
	}
	serialized, err := original.Serialize()
	if err != nil {
		t.Errorf("Serialize() error = %v", err)
	}

	deserialized := &BinaryData{}
	err = deserialized.Deserialize(serialized)
	if err != nil {
		t.Errorf("Deserialize() error = %v", err)
	}

	if string(deserialized.Data) != originalData || deserialized.MetaInfo != original.MetaInfo {
		t.Errorf("Original and deserialized objects are not equal. Original: %v, Deserialized: %v", original, deserialized)
	}
}

func TestLoginPasswordData_Serialize_Deserialize(t *testing.T) {
	original := &LoginPasswordData{
		Login:    "user@example.com",
		Password: "securepassword123",
		MetaInfo: "Test account",
	}
	serialized, err := original.Serialize()
	if err != nil {
		t.Errorf("Serialize() error = %v", err)
	}

	deserialized := &LoginPasswordData{}
	err = deserialized.Deserialize(serialized)
	if err != nil {
		t.Errorf("Deserialize() error = %v", err)
	}

	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Original and deserialized objects are not equal. Original: %v, Deserialized: %v", original, deserialized)
	}
}
