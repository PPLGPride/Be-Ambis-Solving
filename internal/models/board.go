package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type BoardColumn struct {
	ID    string `bson:"id" json:"id"`
	Name  string `bson:"name" json:"name"`
	Order int    `bson:"order" json:"order"`
}

type Board struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	OwnerID     primitive.ObjectID   `bson:"ownerId" json:"ownerId"`
	Name        string               `bson:"name" json:"name"`
	Description *string              `bson:"description,omitempty" json:"description,omitempty"`
	Members     []primitive.ObjectID `bson:"members" json:"members"`
	Columns     []BoardColumn        `bson:"columns" json:"columns"`
	IsArchived  bool                 `bson:"isArchived" json:"isArchived"`
	TimeMeta    `bson:",inline"`
}

func (b *Board) CollectionName() string { return "boards" }
