package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHttpClient реализует интерфейс HttpClient для тестирования
type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestAuthClient_Register(t *testing.T) {
	mockHTTPClient := new(MockHttpClient)
	authClient := NewAuthClient(mockHTTPClient, "http://example.com")

	mockResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	err := authClient.Register("testuser", "testpassword")

	assert.NoError(t, err)

	mockHTTPClient.AssertExpectations(t)
}

func TestAuthClient_Register_UserExists(t *testing.T) {
	mockHTTPClient := new(MockHttpClient)
	authClient := NewAuthClient(mockHTTPClient, "http://example.com")

	mockResponse := &http.Response{
		StatusCode: http.StatusConflict,
		Body:       io.NopCloser(bytes.NewBufferString("User already exists")),
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	err := authClient.Register("existinguser", "password")

	assert.Error(t, err)

	mockHTTPClient.AssertExpectations(t)
}

func TestAuthClient_Login_Success(t *testing.T) {
	mockHTTPClient := new(MockHttpClient)
	authClient := NewAuthClient(mockHTTPClient, "http://example.com")

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header: http.Header{
			"Set-Cookie": []string{"AuthToken=validtoken; Path=/; HttpOnly"},
		},
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	token, err := authClient.Login("validuser", "validpassword")

	assert.NoError(t, err)
	assert.Equal(t, "validtoken", token)

	mockHTTPClient.AssertExpectations(t)
}

func TestAuthClient_Login_WrongCredentials(t *testing.T) {
	mockHTTPClient := new(MockHttpClient)
	authClient := NewAuthClient(mockHTTPClient, "http://example.com")

	mockResponse := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       ioutil.NopCloser(bytes.NewBufferString("Invalid credentials")),
	}

	mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

	token, err := authClient.Login("user", "wrongpassword")

	assert.Error(t, err)
	assert.Empty(t, token)
	mockHTTPClient.AssertExpectations(t)
}
