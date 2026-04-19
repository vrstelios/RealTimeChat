package database

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
	"time"
)

var DB *mongo.Database

func InitDatabase(uri string) {
	// ctx will be used to set deadline for process, here
	// deadline will of 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	DB = client.Database("realtimechat")
	log.Println("MongoDB connected!")

	EnsureIndexes()
}

// Collection Return the collection from database
func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}
