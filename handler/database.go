package handler

import (
	"context"
	"gbGATEWAY/config"
	"gbGATEWAY/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DataBaseHandler struct {
	Mongo  config.MongoDB
	Logger *utils.Logger
}

// if operation find no document it will return ErrNoDocuments error.
func (db *DataBaseHandler) IsUserRegistered(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result := db.Mongo.Users.FindOne(
		context.TODO(),
		bson.M{"_id": _id},
	)
	if result.Err() != nil {
		return err
	}
	return nil
}
