package utils

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MustObjectID(hex string) (primitive.ObjectID, error) {
	if hex == "" {
		return primitive.NilObjectID, errors.New("empty id")
	}
	oid, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return oid, nil
}
