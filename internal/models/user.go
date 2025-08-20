package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Name         string             `bson:"name" json:"name"`
	AvatarURL    *string            `bson:"avatarUrl,omitempty" json:"avatarUrl,omitempty"`
	Role         UserRole           `bson:"role" json:"role"`
	IsActive     bool               `bson:"isActive" json:"isActive"`
	TimeMeta     `bson:",inline"`
}

func (u *User) CollectionName() string { return "users" }
