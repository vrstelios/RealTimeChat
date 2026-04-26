package database

import (
	"RealTimeChat/backend/internal/type/model"
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"time"
)

func SaveMessage(room, name, message, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := model.Message{
		Id:        bson.NewObjectID(),
		Room:      room,
		Name:      name,
		Message:   message,
		Role:      role,
		Timestamp: time.Now(),
	}

	_, err := Collection("messages").InsertOne(ctx, msg)
	if err != nil {
		log.Println("Failed to save message:", err)
		return err
	}
	return nil
}

func GetMessages(room string) ([]model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"room": room}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := Collection("messages").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []model.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func SaveDocument(room, filename string, chuckCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"room": room}
	update := bson.M{
		"$set": bson.M{
			"room":        room,
			"file":        filename,
			"chuckCount":  chuckCount,
			"lastUpdated": time.Now(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := Collection("documents").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Failed to save document metadata:", err)
		return err
	}
	return nil
}

func GetDocuments(room string) ([]model.Document, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"room": room}
	opts := options.Find().SetSort(bson.D{{Key: "lastUpdated", Value: -1}})

	cursor, err := Collection("documents").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []model.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func EnsureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "room", Value: 1},
			{Key: "timestamp", Value: 1},
		},
	}

	_, err := Collection("messages").Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Println("Failed to create index:", err)
		return
	}
	log.Println("MongoDB indexes created!")
}
