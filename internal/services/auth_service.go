package services

import (
	"context"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (token string, userID string, err error)
}

type authService struct {
	users UserService
}

func NewAuthService(users UserService) AuthService {
	return &authService{users: users}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	if !utils.CheckPassword(u.PasswordHash, password) {
		return "", "", ErrInvalidCredentials
	}
	token, err := utils.GenerateJWT(u.ID.Hex())
	if err != nil {
		return "", "", err
	}
	return token, u.ID.Hex(), nil
}
