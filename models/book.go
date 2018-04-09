package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Book struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Author      string        `bson:"author" json:"author"`
	Publisher   string        `bson:"publisher" json:"publisher"`
	PublishDate time.Time     `bson:"publish_date" json:"publish_date"`
	Rating      int           `bson:"rating" json:"rating"`
	Status      string        `bson:"status" json:"status"`
}
