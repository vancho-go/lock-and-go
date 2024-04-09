package auth

import (
	"context"
	"fmt"
	"github.com/vancho-go/lock-and-go/internal/model"
	"github.com/vancho-go/lock-and-go/internal/repository/storage/psql"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
	"github.com/vancho-go/lock-and-go/pkg/customerrors"
	"golang.org/x/crypto/bcrypt"
)

// UserAuthService сервис для работы с авторизацией пользователей.
type UserAuthService struct {
	repo psql.UserRepository
	jwt  jwt.Manager
}

// NewUserAuthService конструктор UserAuthService.
func NewUserAuthService(repo psql.UserRepository, jwt jwt.Manager) *UserAuthService {
	return &UserAuthService{repo: repo, jwt: jwt}
}

// hashPassword генерирует хэш пароля.
func (s *UserAuthService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generating hash from password error: %w", err)
	}
	return string(hashedPassword), nil
}

// isPasswordEqualsToHashedPassword сравнивает пароль с его хешем.
func (s *UserAuthService) isPasswordEqualsToHashedPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Register метод, обрабатывающий запрос на регистрацию пользователя.
func (s *UserAuthService) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return err
	}
	user := &model.UserHashed{Username: username, PasswordHash: hashedPassword}
	return s.repo.CreateUser(ctx, user)
}

// Authenticate метод, обрабатывающий запрос на аутентификацию пользователя.
func (s *UserAuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if !s.isPasswordEqualsToHashedPassword(password, user.PasswordHash) {
		return "", customerrors.ErrWrongPassword
	}

	return s.jwt.GenerateToken(user.ID)
}
