package services

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BoardService interface {
	Create(ctx context.Context, ownerID primitive.ObjectID, name string, desc *string, columns []models.BoardColumn, members []primitive.ObjectID) (*models.Board, error)
	ListForUser(ctx context.Context, userID primitive.ObjectID) ([]models.Board, error)
	Get(ctx context.Context, id primitive.ObjectID) (*models.Board, error)
	Update(ctx context.Context, id primitive.ObjectID, name *string, desc *string, columns *[]models.BoardColumn, members *[]primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type boardService struct{}

func NewBoardService() BoardService { return &boardService{} }

func defaultColumns() []models.BoardColumn {
	return []models.BoardColumn{
		{ID: uuid.NewString(), Name: "Planned", Order: 1},
		{ID: uuid.NewString(), Name: "In Progress", Order: 2},
		{ID: uuid.NewString(), Name: "Done", Order: 3},
	}
}

func (s *boardService) Create(ctx context.Context, ownerID primitive.ObjectID, name string, desc *string, columns []models.BoardColumn, members []primitive.ObjectID) (*models.Board, error) {
	if len(columns) == 0 {
		columns = defaultColumns()
	}
	now := time.Now().UTC()
	b := &models.Board{
		ID:          primitive.NewObjectID(),
		OwnerID:     ownerID,
		Name:        name,
		Description: desc,
		Members:     members,
		Columns:     columns,
		IsArchived:  false,
		TimeMeta:    models.TimeMeta{CreatedAt: now, UpdatedAt: now},
	}
	_, err := config.MongoDB.Collection("boards").InsertOne(ctx, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *boardService) ListForUser(ctx context.Context, userID primitive.ObjectID) ([]models.Board, error) {
	cur, err := config.MongoDB.Collection("boards").Find(ctx, bson.M{
		"$or": []bson.M{
			{"ownerId": userID},
			{"members": userID},
		},
	}, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []models.Board
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *boardService) Get(ctx context.Context, id primitive.ObjectID) (*models.Board, error) {
	var b models.Board
	err := config.MongoDB.Collection("boards").FindOne(ctx, bson.M{"_id": id}).Decode(&b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *boardService) Update(ctx context.Context, id primitive.ObjectID, name *string, desc *string, columns *[]models.BoardColumn, members *[]primitive.ObjectID) error {
	set := bson.M{"updatedAt": time.Now().UTC()}
	if name != nil {
		set["name"] = *name
	}
	if desc != nil {
		set["description"] = desc
	}
	if columns != nil {
		set["columns"] = *columns
	}
	if members != nil {
		set["members"] = *members
	}
	_, err := config.MongoDB.Collection("boards").UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}

func (s *boardService) Delete(ctx context.Context, id primitive.ObjectID) error {
	// Hapus board
	if _, err := config.MongoDB.Collection("boards").DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}
	// Hapus semua tasks board (soft-cascade)
	_, _ = config.MongoDB.Collection("tasks").DeleteMany(ctx, bson.M{"boardId": id})
	return nil
}
