package services

import (
	"context"
	"errors"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, name, email, password string) (*models.User, error)
}

type userService struct{}

func NewUserService() UserService { return &userService{} }

func (s *userService) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := config.MongoDB.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *userService) Create(ctx context.Context, name, email, password string) (*models.User, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		Role:         models.RoleUser,
		IsActive:     true,
		TimeMeta:     models.TimeMeta{CreatedAt: now, UpdatedAt: now},
	}
	_, err = config.MongoDB.Collection("users").InsertOne(ctx, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")
