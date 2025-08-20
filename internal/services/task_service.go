package services

import (
	"context"
	"errors"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskService interface {
	ListByBoard(ctx context.Context, boardID primitive.ObjectID) ([]models.Task, error)
	Create(ctx context.Context, boardID, userID primitive.ObjectID, title string, desc *string, columnId string, status *models.TaskStatus, due *time.Time, assignees []primitive.ObjectID) (*models.Task, error)
	Get(ctx context.Context, id primitive.ObjectID) (*models.Task, error)
	Update(ctx context.Context, id primitive.ObjectID, patch bson.M, updater primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Move(ctx context.Context, id primitive.ObjectID, toColumn string, toPos int) error
}

type taskService struct{}

func NewTaskService() TaskService { return &taskService{} }

func (s *taskService) ListByBoard(ctx context.Context, boardID primitive.ObjectID) ([]models.Task, error) {
	cur, err := config.MongoDB.Collection("tasks").Find(ctx, bson.M{"boardId": boardID},
		options.Find().SetSort(bson.D{{Key: "columnId", Value: 1}, {Key: "order", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []models.Task
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *taskService) maxOrder(ctx context.Context, boardID primitive.ObjectID, columnId string) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "order", Value: -1}})
	var t models.Task
	err := config.MongoDB.Collection("tasks").FindOne(ctx, bson.M{"boardId": boardID, "columnId": columnId}, opts).Decode(&t)
	if err != nil {
		return 0, nil // tidak apa-apa, berarti kosong
	}
	if t.Order == nil {
		v := 0
		return v, nil
	}
	return *t.Order, nil
}

// map status default berdasarkan nama kolom
func defaultStatusForColumn(name string) models.TaskStatus {
	switch name {
	case "In Progress":
		return models.StatusInProgress
	case "Done":
		return models.StatusDone
	default:
		return models.StatusPlanned
	}
}

func (s *taskService) verifyColumn(ctx context.Context, boardID primitive.ObjectID, columnId string) (string, error) {
	// return column name untuk bantu status default
	var b models.Board
	if err := config.MongoDB.Collection("boards").FindOne(ctx, bson.M{"_id": boardID, "columns.id": columnId}).Decode(&b); err != nil {
		return "", errors.New("column not found in board")
	}
	for _, c := range b.Columns {
		if c.ID == columnId {
			return c.Name, nil
		}
	}
	return "", errors.New("column not found")
}

func (s *taskService) Create(ctx context.Context, boardID, userID primitive.ObjectID, title string, desc *string, columnId string, status *models.TaskStatus, due *time.Time, assignees []primitive.ObjectID) (*models.Task, error) {
	colName, err := s.verifyColumn(ctx, boardID, columnId)
	if err != nil {
		return nil, err
	}

	max, _ := s.maxOrder(ctx, boardID, columnId)
	next := max + 1
	st := models.StatusPlanned
	if status != nil {
		st = *status
	} else {
		st = defaultStatusForColumn(colName)
	}
	now := time.Now().UTC()
	t := &models.Task{
		ID:          primitive.NewObjectID(),
		BoardID:     boardID,
		Title:       title,
		Description: desc,
		Status:      st,
		ColumnID:    columnId,
		Priority:    models.PriorityMedium,
		Assignees:   assignees,
		DueDate:     due,
		Order:       &next,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		TimeMeta:    models.TimeMeta{CreatedAt: now, UpdatedAt: now},
	}
	_, err = config.MongoDB.Collection("tasks").InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *taskService) Get(ctx context.Context, id primitive.ObjectID) (*models.Task, error) {
	var t models.Task
	if err := config.MongoDB.Collection("tasks").FindOne(ctx, bson.M{"_id": id}).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *taskService) Update(ctx context.Context, id primitive.ObjectID, patch bson.M, updater primitive.ObjectID) error {
	patch["updatedAt"] = time.Now().UTC()
	patch["updatedBy"] = updater
	_, err := config.MongoDB.Collection("tasks").UpdateByID(ctx, id, bson.M{"$set": patch})
	return err
}

func (s *taskService) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := config.MongoDB.Collection("tasks").DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Move: geser order di kolom sumber/tujuan (best-effort, tanpa transaksi)
func (s *taskService) Move(ctx context.Context, id primitive.ObjectID, toColumn string, toPos int) error {
	// Ambil task lama
	task, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Validasi kolom tujuan
	if _, err := s.verifyColumn(ctx, task.BoardID, toColumn); err != nil {
		return err
	}

	srcCol := task.ColumnID
	var srcOrder int
	if task.Order != nil {
		srcOrder = *task.Order
	}

	// 1) Jika pindah kolom: longgar-kan kolom tujuan untuk toPos
	if srcCol != toColumn {
		_, _ = config.MongoDB.Collection("tasks").UpdateMany(ctx,
			bson.M{"boardId": task.BoardID, "columnId": toColumn, "order": bson.M{"$gte": toPos}},
			bson.M{"$inc": bson.M{"order": 1}},
		)
		// rapikan kolom sumber (tutup gap)
		_, _ = config.MongoDB.Collection("tasks").UpdateMany(ctx,
			bson.M{"boardId": task.BoardID, "columnId": srcCol, "order": bson.M{"$gt": srcOrder}},
			bson.M{"$inc": bson.M{"order": -1}},
		)
	} else {
		// tetap di kolom sama: geser ruang untuk toPos lalu tutup gap dari posisi lama
		if toPos < srcOrder {
			// geser turun semua item yang >= toPos dan < srcOrder
			_, _ = config.MongoDB.Collection("tasks").UpdateMany(ctx,
				bson.M{"boardId": task.BoardID, "columnId": srcCol, "order": bson.M{"$gte": toPos, "$lt": srcOrder}},
				bson.M{"$inc": bson.M{"order": 1}},
			)
		} else if toPos > srcOrder {
			// geser naik semua item yang <= toPos dan > srcOrder
			_, _ = config.MongoDB.Collection("tasks").UpdateMany(ctx,
				bson.M{"boardId": task.BoardID, "columnId": srcCol, "order": bson.M{"$lte": toPos, "$gt": srcOrder}},
				bson.M{"$inc": bson.M{"order": -1}},
			)
		}
	}

	// 2) Update task tujuan
	st := task.Status
	if srcCol != toColumn {
		// set status default sesuai kolom tujuan
		// (opsional: frontend juga bisa kirim status)
		// ambil nama kolom tujuan dulu:
		var b models.Board
		if err := config.MongoDB.Collection("boards").
			FindOne(ctx, bson.M{"_id": task.BoardID, "columns.id": toColumn}).Decode(&b); err == nil {
			for _, c := range b.Columns {
				if c.ID == toColumn {
					st = defaultStatusForColumn(c.Name)
					break
				}
			}
		}
	}

	_, err = config.MongoDB.Collection("tasks").UpdateByID(ctx, id, bson.M{
		"$set": bson.M{
			"columnId":  toColumn,
			"order":     toPos,
			"status":    st,
			"updatedAt": time.Now().UTC(),
		},
	})
	return err
}
