package services

import (
	"context"
	"errors"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteService interface {
	Create(ctx context.Context, authorID primitive.ObjectID, content string, boardID *primitive.ObjectID, taskID *primitive.ObjectID, onAt *time.Time, pinned bool) (*models.Note, error)
	ListByBoard(ctx context.Context, boardID primitive.ObjectID) ([]models.Note, error)
	ListByTask(ctx context.Context, taskID primitive.ObjectID) ([]models.Note, error)
	Update(ctx context.Context, id primitive.ObjectID, patch bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type noteService struct{}

func NewNoteService() NoteService { return &noteService{} }

func (s *noteService) taskBoardID(ctx context.Context, taskID primitive.ObjectID) (*primitive.ObjectID, error) {
	var t models.Task
	if err := config.MongoDB.Collection("tasks").FindOne(ctx, bson.M{"_id": taskID}).Decode(&t); err != nil {
		return nil, err
	}
	bid := t.BoardID
	return &bid, nil
}

func (s *noteService) Create(ctx context.Context, authorID primitive.ObjectID, content string, boardID *primitive.ObjectID, taskID *primitive.ObjectID, onAt *time.Time, pinned bool) (*models.Note, error) {
	if content == "" {
		return nil, errors.New("content required")
	}
	// Jika tak ada boardID tapi ada taskID → turunkan boardID dari task
	if boardID == nil && taskID != nil {
		if b, err := s.taskBoardID(ctx, *taskID); err == nil {
			boardID = b
		}
	}
	now := time.Now().UTC()
	n := &models.Note{
		ID:           primitive.NewObjectID(),
		BoardID:      boardID,
		TaskID:       taskID,
		AuthorID:     authorID,
		Content:      content,
		Pinned:       pinned,
		OnTimelineAt: onAt,
		TimeMeta:     models.TimeMeta{CreatedAt: now, UpdatedAt: now},
	}
	if _, err := config.MongoDB.Collection("notes").InsertOne(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *noteService) ListByBoard(ctx context.Context, boardID primitive.ObjectID) ([]models.Note, error) {
	cur, err := config.MongoDB.Collection("notes").Find(ctx, bson.M{"boardId": boardID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []models.Note
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *noteService) ListByTask(ctx context.Context, taskID primitive.ObjectID) ([]models.Note, error) {
	cur, err := config.MongoDB.Collection("notes").Find(ctx, bson.M{"taskId": taskID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []models.Note
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *noteService) Update(ctx context.Context, id primitive.ObjectID, patch bson.M) error {
	if len(patch) == 0 {
		return nil
	}
	patch["updatedAt"] = time.Now().UTC()
	_, err := config.MongoDB.Collection("notes").UpdateByID(ctx, id, bson.M{"$set": patch})
	return err
}

func (s *noteService) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := config.MongoDB.Collection("notes").DeleteOne(ctx, bson.M{"_id": id})
	return err
}
