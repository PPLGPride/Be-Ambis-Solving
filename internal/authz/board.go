package authz

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsMemberOrOwner(ctx context.Context, boardID, userID primitive.ObjectID) (bool, error) {
	var doc struct {
		OwnerID primitive.ObjectID `bson:"ownerId"`
	}
	if err := config.MongoDB.Collection("boards").FindOne(ctx, bson.M{"_id": boardID}).Decode(&doc); err != nil {
		return false, err
	}
	if doc.OwnerID == userID {
		return true, nil
	}
	n, err := config.MongoDB.Collection("boards").CountDocuments(ctx, bson.M{"_id": boardID, "members": userID})
	return n > 0, err
}

func BoardIDFromTask(ctx context.Context, taskID primitive.ObjectID) (primitive.ObjectID, error) {
	var t struct {
		BoardID primitive.ObjectID `bson:"boardId"`
	}
	err := config.MongoDB.Collection("tasks").FindOne(ctx, bson.M{"_id": taskID}).Decode(&t)
	return t.BoardID, err
}

func WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 4*time.Second)
}
