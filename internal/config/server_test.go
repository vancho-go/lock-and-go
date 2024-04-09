package config

import (
	"reflect"
	"testing"
	"time"
)

func TestGetJWTConfig(t *testing.T) {
	// GetJWTConfig всегда возвращает указатель на один экземпляр JWTConfig,
	// ожидаем, что два последовательных вызова вернут один и тот же объект.
	tests := []struct {
		name string
		want *JWTConfig
	}{
		{
			name: "Singleton instance",
			want: GetJWTConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetJWTConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJWTConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJWTSecretKey(t *testing.T) {
	expectedKey := "testSecret"
	GetJWTConfig().SecretKey = expectedKey
	if got := GetJWTSecretKey(); got != expectedKey {
		t.Errorf("GetJWTSecretKey() = %v, want %v", got, expectedKey)
	}
}

func TestGetJWTTokenDuration(t *testing.T) {
	expectedDuration := 5 * time.Minute
	GetJWTConfig().TokenDuration = expectedDuration
	if got := GetJWTTokenDuration(); got != expectedDuration {
		t.Errorf("GetJWTTokenDuration() = %v, want %v", got, expectedDuration)
	}
}

func Test_initializeJWTConfig(t *testing.T) {
	tests := []struct {
		name             string
		secretKey        string
		tokenDurationStr string
		wantErr          bool
	}{
		{
			name:             "Valid duration",
			secretKey:        "mySecret",
			tokenDurationStr: "2h",
			wantErr:          false,
		},
		{
			name:             "Invalid duration",
			secretKey:        "mySecret",
			tokenDurationStr: "invalid",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := initializeJWTConfig(tt.secretKey, tt.tokenDurationStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("initializeJWTConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isConfigFull(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}

	tests := []struct {
		name    string
		config  any
		wantErr bool
	}{
		{
			name:    "All fields are filled",
			config:  &TestStruct{Field1: "value", Field2: 1},
			wantErr: false,
		},
		{
			name:    "Missing Field1",
			config:  &TestStruct{Field2: 1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isConfigFull(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("isConfigFull() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newServerConfigLoader(t *testing.T) {
	type args struct {
		loaderType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "env loader",
			args: args{
				loaderType: "env",
			},
			wantErr: false,
		},
		{
			name: "flag loader",
			args: args{
				loaderType: "flag",
			},
			wantErr: false,
		},
		{
			name: "unsupported loader",
			args: args{
				loaderType: "unsupported",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newServerConfigLoader(tt.args.loaderType)
			if (err != nil) != tt.wantErr {
				t.Errorf("newServerConfigLoader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && got != nil {
				t.Errorf("newServerConfigLoader() expected no loader to be returned on error, got %v", got)
			}
		})
	}
}
