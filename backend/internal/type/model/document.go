package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type Document struct {
	Id          bson.ObjectID `bson:"_id"`
	Room        string        `bson:"room"`
	File        string        `bson:"file"`
	ChuckCount  int           `bson:"chuckCount"`
	LastUpdated time.Time     `json:"LastUpdated" bson:"lastUpdated"`
}
