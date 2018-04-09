package dao

import (
	"log"

	. "github.com/dmallory/books/models"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type BookDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "books"
)

func (dao *BookDAO) Connect() {
	session, err := mgo.Dial(dao.Server)
	if err != nil {
		log.Fatal(err)
	}
	db = session.DB(dao.Database)
}

// Find list of items
func (dao *BookDAO) FindAll() ([]Book, error) {
	var items []Book
	err := db.C(COLLECTION).Find(bson.M{}).All(&items)
	return items, err
}

// Find a item by its id
func (dao *BookDAO) FindById(id string) (Book, error) {
	var item Book
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&item)
	return item, err
}

// Insert an item
func (dao *BookDAO) Insert(item Book) error {
	err := db.C(COLLECTION).Insert(&item)
	return err
}

// Delete an existing item
func (dao *BookDAO) DeleteById(id string) error {
	err := db.C(COLLECTION).RemoveId(bson.ObjectIdHex(id))
	return err
}

// Clear collection
func (dao *BookDAO) Clear() error {
	_, err := db.C(COLLECTION).RemoveAll(nil)
	return err
}

// Update an existing item
func (dao *BookDAO) Update(item Book) error {
	err := db.C(COLLECTION).UpdateId(item.ID, &item)
	return err
}
