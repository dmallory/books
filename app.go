package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	. "github.com/dmallory/books/config"
	. "github.com/dmallory/books/dao"
	. "github.com/dmallory/books/models"
	"github.com/gorilla/mux"
)

var config = Config{}
var books = BookDAO{}

// Get all books
func AllBooks(writer http.ResponseWriter, request *http.Request) {
	books, err := books.FindAll()
	if err != nil {
		returnInternalError(writer, err)
		return
	}
	returnData(writer, books)
}

// Get a single book by ID
func FindBook(writer http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	id := params["id"]
	if msg := validateId(id); msg != "" {
		returnInvalidId(writer, msg)
		return
	}
	book, err := books.FindById(id)
	if err != nil {
		returnNotFound(writer, id)
		return
	}
	returnData(writer, book)
}

// Create a new book
func CreateBook(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	var book Book
	if err := json.NewDecoder(request.Body).Decode(&book); err != nil {
		returnInvalidJson(writer)
		return
	}
	if msg := validateBook(book); msg != "" {
		returnInvalidData(writer, msg)
		return
	}
	book.ID = bson.NewObjectId()
	if err := books.Insert(book); err != nil {
		returnInternalError(writer, err)
		return
	}
	returnCreated(writer, book)
}

// Update an existing book
func UpdateBook(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	var book Book
	if err := json.NewDecoder(request.Body).Decode(&book); err != nil {
		returnInvalidJson(writer)
		return
	}
	if err := books.Update(book); err != nil {
		returnInternalError(writer, err)
		return
	}
	returnSuccess(writer)
}

// Delete an existing book
func DeleteBook(writer http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	id := params["id"]
	if msg := validateId(id); msg != "" {
		returnInvalidId(writer, msg)
		return
	}
	err := books.DeleteById(id)
	if err != nil {
		returnNotFound(writer, id)
		return
	}
	returnSuccess(writer)
}

// Validation helpers

func validateId(id string) string {
	if len(id) != 24 {
		return id
	}
	return ""
}

// Validate book (required fields, reasonable strings, range limits, etc. - beyond valid JSON)
func validateBook(book Book) string {
	var msgs []string
	if msg := checkRequiredString("Title", book.Title); msg != "" {
		msgs = append(msgs, msg)
	}
	if msg := checkRequiredString("Author", book.Author); msg != "" {
		msgs = append(msgs, msg)
	}
	if msg := checkRequiredString("Publisher", book.Publisher); msg != "" {
		msgs = append(msgs, msg)
	}
	if msg := checkValidDate("Publish Date", book.PublishDate); msg != "" {
		msgs = append(msgs, msg)
	}
	if msg := checkRangedInt("Rating", book.Rating, 1, 3); msg != "" {
		msgs = append(msgs, msg)
	}
	if msg := checkEnumeratedString("Status", book.Status, []string{"CheckedIn", "CheckedOut"}); msg != "" {
		msgs = append(msgs, msg)
	}
	if len(msgs) > 0 {
		return strings.Join(msgs, "; ")
	}
	return ""
}

func checkRequiredString(attribute string, value string) string {
	if strings.TrimSpace(value) == "" {
		return "Value must be present and not empty: " + attribute
	}
	return ""
}

func checkRangedInt(attribute string, value int, lower int, upper int) string {
	if value < lower || value > upper {
		return "Value must be in specified range: " + attribute + " (" + strconv.Itoa(lower) + "-" + strconv.Itoa(upper) + ")"
	}
	return ""
}

func checkValidDate(attribute string, value time.Time) string {
	if value.Unix() == 0 {
		return "Value must be a valid date (ISO 8601): " + attribute
	}
	return ""
}

func checkEnumeratedString(attribute string, value string, values []string) string {
	if !contains(value, values) {
		return "Value must be in specified set: " + attribute + " (" + strings.Join(values, ",") + ")"
	}
	return ""
}

// Seriously?
func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Response helpers

// Invalid JSON received
func returnInvalidJson(writer http.ResponseWriter) {
	returnError(writer, http.StatusBadRequest, "Invalid JSON")
}

// Invalid ID
func returnInvalidId(writer http.ResponseWriter, id string) {
	returnError(writer, http.StatusBadRequest, "Invalid ID: "+id)
}

// Invalid data
func returnInvalidData(writer http.ResponseWriter, message string) {
	returnError(writer, http.StatusBadRequest, "Invalid data: "+message)
}

// ID not found
func returnNotFound(writer http.ResponseWriter, id string) {
	returnError(writer, http.StatusBadRequest, "ID not found: "+id)
}

// Internal error (database, etc.)
func returnInternalError(writer http.ResponseWriter, err error) {
	returnError(writer, http.StatusInternalServerError, err.Error())
}

// Error with code and message
func returnError(writer http.ResponseWriter, code int, message string) {
	returnJson(writer, code, map[string]string{
		"error": message,
	})
}

// Created
func returnCreated(writer http.ResponseWriter, data interface{}) {
	returnJson(writer, http.StatusCreated, data)
}

// Data
func returnData(writer http.ResponseWriter, data interface{}) {
	returnJson(writer, http.StatusOK, data)
}

// Success
func returnSuccess(writer http.ResponseWriter) {
	returnJson(writer, http.StatusOK, map[string]string{
		"result": "success",
	})
}

// JSON response
func returnJson(writer http.ResponseWriter, code int, data interface{}) {
	response, _ := json.Marshal(data)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	if string(response) == "null" {
		// Don't send "null", leave response empty
	} else {
		writer.Write(response)
	}
}

// Creation of books database, for server or testing
func Books() BookDAO {
	var books BookDAO
	books.Server = "localhost"
	books.Database = "books"
	books.Connect()
	return books
}

// App setup
func init() {

	// Load config
	config.Read()

	// Set up and connect database
	books = Books()
}

// Main app entry point
func main() {

	// Create main router for REST API
	router := Router()

	// Start server
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("Failed to start server", err)
	}
}

// Creation of router, for server or testing
func Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/books", AllBooks).Methods("GET")
	router.HandleFunc("/books", CreateBook).Methods("POST")
	router.HandleFunc("/books", UpdateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", DeleteBook).Methods("DELETE")
	router.HandleFunc("/books/{id}", FindBook).Methods("GET")
	return router
}
