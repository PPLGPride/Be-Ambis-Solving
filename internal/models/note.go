package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	BoardID      *primitive.ObjectID `bson:"boardId,omitempty" json:"boardId,omitempty"`
	TaskID       *primitive.ObjectID `bson:"taskId,omitempty" json:"taskId,omitempty"`
	AuthorID     primitive.ObjectID  `bson:"authorId" json:"authorId"`
	Content      string              `bson:"content" json:"content"`
	Pinned       bool                `bson:"pinned" json:"pinned"`
	OnTimelineAt *time.Time          `bson:"onTimelineAt,omitempty" json:"onTimelineAt,omitempty"`
	TimeMeta     `bson:",inline"`
}

func (n *Note) CollectionName() string { return "notes" }
