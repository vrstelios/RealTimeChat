package database

import (
	"RealTimeChat/backend/internal/type/model"
	"context"
	"errors"
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

func GetUser(userId bson.ObjectID, email string) ([]model.Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if !userId.IsZero() {
		filter["_id"] = userId
	}
	if email != "" {
		filter["email"] = email
	}

	cursor, err := Collection("users").Find(ctx, filter)
	if err != nil {
		log.Println("Failed to get users:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.Users
	if err := cursor.All(ctx, &users); err != nil {
		log.Println("Failed to decode users:", err)
		return nil, err
	}

	return users, nil
}

func PostUser(user model.Users) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing model.Users
	err := Collection("users").FindOne(ctx, bson.M{"email": user.Email}).Decode(&existing)
	if err == nil {
		return errors.New("user with this email already exists")
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.Println("Failed to check existing user:", err)
		return err
	}

	user.Id = bson.NewObjectID()
	user.CreatedAt = time.Now()

	_, err = Collection("users").InsertOne(ctx, user)
	if err != nil {
		log.Println("Failed to insert user:", err)
		return err
	}

	log.Printf("User created: %s (%s)\n", user.Name, user.Email)
	return nil
}

func PutUser(userId bson.ObjectID, updates model.Users) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setFields := bson.M{}

	if updates.Name != "" {
		setFields["name"] = updates.Name
	}
	if updates.Email != "" {
		var existing model.Users
		err := Collection("users").FindOne(ctx, bson.M{
			"email": updates.Email,
			"_id":   bson.M{"_id": userId},
		}).Decode(&existing)
		if err == nil {
			return errors.New("email already in use by another user")
		}
		setFields["email"] = updates.Email
	}
	if updates.Password != "" {
		setFields["password"] = updates.Password
	}
	if updates.Token != nil {
		setFields["token"] = updates.Token
	}
	if updates.RefreshToken != nil {
		setFields["refreshToken"] = updates.RefreshToken
	}

	if len(setFields) == 0 {
		return errors.New("no fields to update")
	}

	filter := bson.M{"_id": userId}
	update := bson.M{"$set": setFields}

	result, err := Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Failed to update user:", err)
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	//log.Printf("User updated: %s\n", userId)
	return nil
}
