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

type UserService struct {
	repo psql.UserRepository
	jwt  jwt.Manager
}

func NewUserAuthService(repo psql.UserRepository, jwt jwt.Manager) *UserService {
	return &UserService{repo: repo, jwt: jwt}
}

func (s *UserService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashPassword: generating hash from password error: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *UserService) isPasswordEqualsToHashedPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *UserService) Register(ctx context.Context, username, password string) error {
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	user := &model.UserHashed{Username: username, PasswordHash: hashedPassword}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("authenticate: %w", err)
	}

	if !s.isPasswordEqualsToHashedPassword(password, user.PasswordHash) {
		return "", customerrors.ErrWrongPassword
	}

	return s.jwt.GenerateToken(user.ID)
}
