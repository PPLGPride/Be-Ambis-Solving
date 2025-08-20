package config

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func ConnectMongo(ctx context.Context) error {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(Cfg.MongoURI))
	if err != nil {
		return err
	}
	if err := cl.Ping(ctx, nil); err != nil {
		return err
	}
	MongoClient = cl
	MongoDB = cl.Database(Cfg.DBName)
	log.Println("[mongo] connected")
	return ensureIndexes(ctx)
}

func ensureIndexes(ctx context.Context) error {
	// users: email unique
	users := MongoDB.Collection("users")
	_, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})
	if err != nil {
		return err
	}

	// boards: ownerId, members
	boards := MongoDB.Collection("boards")
	if _, err = boards.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "ownerId", Value: 1}}, Options: options.Index().SetName("ix_ownerId")},
		{Keys: bson.D{{Key: "members", Value: 1}}, Options: options.Index().SetName("ix_members")},
	}); err != nil {
		return err
	}

	// tasks: boardId, status, assignees, dueDate, columnId+order
	tasks := MongoDB.Collection("tasks")
	if _, err = tasks.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "boardId", Value: 1}}, Options: options.Index().SetName("ix_boardId")},
		{Keys: bson.D{{Key: "status", Value: 1}}, Options: options.Index().SetName("ix_status")},
		{Keys: bson.D{{Key: "assignees", Value: 1}}, Options: options.Index().SetName("ix_assignees")},
		{Keys: bson.D{{Key: "dueDate", Value: 1}}, Options: options.Index().SetName("ix_dueDate")},
		{Keys: bson.D{{Key: "boardId", Value: 1}, {Key: "columnId", Value: 1}, {Key: "order", Value: 1}},
			Options: options.Index().SetName("ix_board_column_order")},
	}); err != nil {
		return err
	}

	// notes: taskId, boardId, onTimelineAt
	notes := MongoDB.Collection("notes")
	if _, err = notes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "taskId", Value: 1}}, Options: options.Index().SetName("ix_taskId")},
		{Keys: bson.D{{Key: "boardId", Value: 1}}, Options: options.Index().SetName("ix_boardId")},
		{Keys: bson.D{{Key: "onTimelineAt", Value: 1}}, Options: options.Index().SetName("ix_onTimelineAt")},
	}); err != nil {
		return err
	}

	log.Println("[mongo] indexes ensured")
	return nil
}
