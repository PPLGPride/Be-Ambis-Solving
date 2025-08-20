package models

import "time"

type TimeMeta struct {
	CreatedAt time.Time `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt,omitempty" json:"updatedAt"`
}
