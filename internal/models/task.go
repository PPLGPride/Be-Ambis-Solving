package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskStatus string
type TaskPriority string

const (
	StatusPlanned    TaskStatus = "planned"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"

	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
	PriorityUrgent TaskPriority = "urgent"
)

type Attachment struct {
	Name string `bson:"name" json:"name"`
	URL  string `bson:"url" json:"url"`
	Size int64  `bson:"size" json:"size"`
}

type Task struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	BoardID       primitive.ObjectID   `bson:"boardId" json:"boardId"`
	Title         string               `bson:"title" json:"title"`
	Description   *string              `bson:"description,omitempty" json:"description,omitempty"`
	Status        TaskStatus           `bson:"status" json:"status"`
	ColumnID      string               `bson:"columnId" json:"columnId"`
	Priority      TaskPriority         `bson:"priority" json:"priority"`
	Assignees     []primitive.ObjectID `bson:"assignees" json:"assignees"`
	StartDate     *time.Time           `bson:"startDate,omitempty" json:"startDate,omitempty"`
	DueDate       *time.Time           `bson:"dueDate,omitempty" json:"dueDate,omitempty"`
	EstimateHours *int                 `bson:"estimateHours,omitempty" json:"estimateHours,omitempty"`
	Tags          []string             `bson:"tags,omitempty" json:"tags,omitempty"`
	Attachments   []Attachment         `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Order         *int                 `bson:"order,omitempty" json:"order,omitempty"`
	CreatedBy     primitive.ObjectID   `bson:"createdBy" json:"createdBy"`
	UpdatedBy     primitive.ObjectID   `bson:"updatedBy" json:"updatedBy"`
	TimeMeta      `bson:",inline"`
}

func (t *Task) CollectionName() string { return "tasks" }
