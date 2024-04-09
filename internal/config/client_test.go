package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		loaderType string
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "env loader success",
			args: args{
				loaderType: "env",
			},
			wantErr: false,
		},
		{
			name: "flag loader success",
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
			_, err := NewClient(tt.args.loaderType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRealEnvReader_GetEnv(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "get existing env variable",
			args: args{
				key: "PATH",
			},
			want: os.Getenv("PATH"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RealEnvReader{}
			if got := r.GetEnv(tt.args.key); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockEnvReader struct {
	mock.Mock
}

func (m *MockEnvReader) GetEnv(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func Test_envClientLoader_load(t *testing.T) {
	mockEnvReader := new(MockEnvReader)
	expectedServerAddress := "http://mockserver:8080"
	mockEnvReader.On("GetEnv", "SERVER_ADDRESS_FOR_CLIENT").Return(expectedServerAddress)

	loader := &envClientLoader{
		envReader: mockEnvReader,
	}

	clientConfig, err := loader.load()

	assert.NoError(t, err)
	assert.Equal(t, expectedServerAddress, *clientConfig.ServerAddress)
	mockEnvReader.AssertExpectations(t)
}

type MockFlagSetter struct {
	mock.Mock
}

func (m *MockFlagSetter) StringVar(p *string, name string, value string, usage string) {
	m.Called(p, name, value, usage)
}

func (m *MockFlagSetter) Parse() {
	m.Called()
}

func Test_flagClientLoader_load(t *testing.T) {
	mockFlagSetter := new(MockFlagSetter)
	expectedServerAddress := "http://mockserver:8080"

	mockFlagSetter.On("StringVar", mock.AnythingOfType("*string"), "s", "http://localhost:8080", "address:port of the Server").Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = expectedServerAddress
	}).Once()

	mockFlagSetter.On("Parse").Once()

	loader := &flagClientLoader{
		flagSetter: mockFlagSetter,
	}

	clientConfig, err := loader.load()

	assert.NoError(t, err)
	assert.Equal(t, expectedServerAddress, *clientConfig.ServerAddress)
	mockFlagSetter.AssertExpectations(t)
}

func Test_newClientConfigLoader(t *testing.T) {
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
			got, err := newClientConfigLoader(tt.args.loaderType)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClientConfigLoader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && got != nil {
				t.Errorf("newClientConfigLoader() expected no loader to be returned on error, got %v", got)
			}
		})
	}
}
