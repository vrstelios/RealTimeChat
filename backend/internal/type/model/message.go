package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Message struct {
	Id        bson.ObjectID `bson:"_id"`
	Room      string        `bson:"room"`
	Name      string        `bson:"name"`
	Message   string        `bson:"message"`
	Role      string        `bson:"role"`
	Timestamp time.Time     `bson:"timestamp"`
}
