package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Users struct {
	Id           bson.ObjectID `bson:"_id"`
	Name         string        `bson:"name"`
	Email        string        `bson:"email"`
	Password     string        `bson:"password"`
	Token        *string       `bson:"token,omitempty"`
	RefreshToken *string       `bson:"refreshToken,omitempty"`
	CreatedAt    time.Time     `bson:"createdAt"`
}

type Login struct {
	HashedPassword string
	SessionToken   string
	Role           string
}
