package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	HOST = "http://localhost:3000"
)

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}

func setup() {
	// Clear out database
	books.Clear()
}

func teardown() {
	// Clear out database
	books.Clear()
}

// Test through basic CRUD workflow with valid data
func TestValidCRUD(t *testing.T) {
	router := Router()

	// Get all books when initially empty
	request, _ := http.NewRequest("GET", "/books", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")
	assert.Equal(t, response.Body.String(), "", "No books expected")

	// Create a book
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "A Wrinkle In Time",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       1,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 201, response.Code, "CREATED expected")

	// Check for expected value in each field regardless of order and ID for subsequent requests
	var book = map[string]interface{}{}
	value(response.Body.Bytes(), &book)
	checkBook(t, book)
	id := book["id"].(string)
	assert.NotNil(t, id, "Nil ID")
	assert.NotEqual(t, "", id, "Empty ID")

	// Get all books with saved book
	request, _ = http.NewRequest("GET", "/books", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")
	assert.NotEqual(t, response.Body.String(), "", "Books empty after save")

	// Check for same book that was saved including ID
	data := make([]map[string]interface{}, 1, 1)
	value(response.Body.Bytes(), &data)
	book = data[0]
	checkBook(t, book)
	assert.Equal(t, id, book["id"])

	// Get same book by ID
	request, _ = http.NewRequest("GET", "/books/"+id, nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")
	assert.NotEqual(t, response.Body.String(), "", "Book empty after save")

	// Update some values
	request, _ = http.NewRequest("PUT", "/books", buffer(map[string]interface{}{
		"id":           id,
		"title":        "A Wrinkle In Time 2",
		"author":       "Madeleine L'Engle 2",
		"publisher":    "Farrar, Straus & Giroux 2",
		"publish_date": "1962-01-02T00:00:00-07:00",
		"rating":       2,
		"status":       "CheckedOut",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")

	// Get same book by ID
	request, _ = http.NewRequest("GET", "/books/"+id, nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")
	assert.NotEqual(t, response.Body.String(), "", "Book empty after save")

	// Check for new values
	value(response.Body.Bytes(), &book)
	checkUpdatedBook(t, book)
	assert.Equal(t, id, book["id"])

	// Delete book
	request, _ = http.NewRequest("DELETE", "/books/"+id, nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")

	// Get all books that should be empty again
	request, _ = http.NewRequest("GET", "/books", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK expected")
	assert.Equal(t, response.Body.String(), "", "No books expected")
}

// Test data validation
func TestInvalidData(t *testing.T) {
	router := Router()

	// Create a book with empty title
	request, _ := http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       1,
		"status":       "CheckedIn",
	}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	var data = map[string]interface{}{}
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be present and not empty: Title", data["error"], "Wrong error message")

	// Create a book with missing title
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       1,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be present and not empty: Title", data["error"], "Wrong error message")

	// Create a book with multiple fields missing
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "",
		"author":       "",
		"publisher":    "",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       1,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be present and not empty: Title; Value must be present and not empty: Author; Value must be present and not empty: Publisher", data["error"], "Wrong error message")

	// Create a book with invalid date
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "A Wrinkle In Time",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "asdf",
		"rating":       1,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid JSON", data["error"], "Wrong error message")

	// Create a book with rating below range
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "A Wrinkle In Time",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       0,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be in specified range: Rating (1-3)", data["error"], "Wrong error message")

	// Create a book with rating above range
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "A Wrinkle In Time",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       4,
		"status":       "CheckedIn",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be in specified range: Rating (1-3)", data["error"], "Wrong error message")

	// Create a book with invalid status
	request, _ = http.NewRequest("POST", "/books", buffer(map[string]interface{}{
		"title":        "A Wrinkle In Time",
		"author":       "Madeleine L'Engle",
		"publisher":    "Farrar, Straus & Giroux",
		"publish_date": "1962-01-01T00:00:00-07:00",
		"rating":       1,
		"status":       "CheckedAway",
	}))
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid data: Value must be in specified set: Status (CheckedIn,CheckedOut)", data["error"], "Wrong error message")
}

func TestInvalidRequests(t *testing.T) {
	router := Router()

	// Get book that doesn't exist
	request, _ := http.NewRequest("GET", "/books/5acb40295843ef00e69e28d2", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	var data = map[string]interface{}{}
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "ID not found: 5acb40295843ef00e69e28d2", data["error"], "Wrong error message")

	// Get book with invalid BSON ID
	request, _ = http.NewRequest("GET", "/books/asdf", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "Invalid ID: asdf", data["error"], "Wrong error message")

	// Delete book that doesn't exist
	request, _ = http.NewRequest("DELETE", "/books/5acb40295843ef00e69e28d2", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "ERROR expected")
	value(response.Body.Bytes(), &data)
	assert.Equal(t, "ID not found: 5acb40295843ef00e69e28d2", data["error"], "Wrong error message")
}

// Value helpers
func checkBook(t *testing.T, book map[string]interface{}) {
	assert.Equal(t, "A Wrinkle In Time", book["title"], "Wrong title")
	assert.Equal(t, "Madeleine L'Engle", book["author"], "Wrong autor")
	assert.Equal(t, "Farrar, Straus & Giroux", book["publisher"], "Wrong publisher")
	assert.Equal(t, "1962-01-01T00:00:00-07:00", book["publish_date"], "Wrong date")
	assert.Equal(t, float64(1), book["rating"], "Wrong rating")
	assert.Equal(t, "CheckedIn", book["status"], "Wrong status")
}

func checkUpdatedBook(t *testing.T, book map[string]interface{}) {
	assert.Equal(t, "A Wrinkle In Time 2", book["title"], "Wrong title")
	assert.Equal(t, "Madeleine L'Engle 2", book["author"], "Wrong autor")
	assert.Equal(t, "Farrar, Straus & Giroux 2", book["publisher"], "Wrong publisher")
	assert.Equal(t, "1962-01-02T00:00:00-07:00", book["publish_date"], "Wrong date")
	assert.Equal(t, float64(2), book["rating"], "Wrong rating")
	assert.Equal(t, "CheckedOut", book["status"], "Wrong status")
}

// Data helpers

func buffer(data map[string]interface{}) *bytes.Buffer {
	json, _ := json.Marshal(data)
	return bytes.NewBuffer(json)
}

func value(bytes []byte, data interface{}) {
	json.Unmarshal(bytes, data)
}
