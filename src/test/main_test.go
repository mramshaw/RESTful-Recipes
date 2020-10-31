package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	// local import
	"application"
)

var app application.App

var authUser, authPassword string

func TestMain(m *testing.M) {
	authUser = os.Getenv("AUTH_USER")
	authPassword = os.Getenv("AUTH_PASSWORD")
	app = application.App{}
	app.Initialize(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		authUser,
		authPassword)
	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func TestEmptyTables(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes", nil)
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	body := response.Body.String()
	assert.Equalf(t, body, "[]", "Expected an empty array. Got %s", body)
}

func TestGetBadRecipe(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes/a", nil)
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetNonExistentRecipe(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes/11", nil)
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	assert.Equalf(t, m["error"], "Recipe not found", "Expected the 'error' key of the response to be set to 'Recipe not found'. Got '%s'", m["error"])
}

func TestCreateRecipeNoCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestCreateRecipeWithCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	assert.Equalf(t, m["name"], "test recipe", "Expected recipe name to be 'test recipe'. Got '%v'", m["name"])

	assert.Equalf(t, m["preptime"], 0.1, "Expected recipe prep time to be '0.1'. Got '%v'", m["preptime"])

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["difficulty"], 2.0, "Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])

	assert.Equalf(t, m["vegetarian"], true, "Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["id"], 1.0, "Expected recipe ID to be '1'. Got '%v'", m["id"])
}

func TestCreateDuplicateServerWithCredentials(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	// Now check duplicate

	req, err = http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on 2nd http.NewRequest: %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusConflict, response.Code)
}

func TestGetRecipe(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestGetRecipes(t *testing.T) {
	clearTables()
	addRecipes(3)

	req, err := http.NewRequest("GET", "/v1/recipes?count=25&start=-1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest: %s", err)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)

	assert.Equalf(t, len(mm), 3, "Expected '3' recipes. Got '%v'", len(mm))
}

func TestUpdatePutRecipeNoCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PUT", "/v1/recipes/1", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (PUT): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestUpdatePutRecipeWithCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PUT", "/v1/recipes/1", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (PUT): %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	assert.Equalf(t, m["id"], originalRecipe["id"], "Expected the id to remain the same (%v). Got %v", originalRecipe["id"], m["id"])

	assert.NotEqualf(t, m["name"], originalRecipe["name"], "Expected the name to change from '%v' to '%v'. Got '%v'", originalRecipe["name"], m["name"], m["name"])
	assert.NotEqualf(t, m["preptime"], originalRecipe["preptime"], "Expected the prep time to change from '%v' to '%v'. Got '%v'", originalRecipe["preptime"], m["preptime"], m["preptime"])
	assert.NotEqualf(t, m["difficulty"], originalRecipe["difficulty"], "Expected the difficulty to change from '%v' to '%v'. Got '%v'", originalRecipe["difficulty"], m["difficulty"], m["difficulty"])
	assert.NotEqualf(t, m["vegetarian"], originalRecipe["vegetarian"], "Expected the vegetarian to change from '%v' to '%v'. Got '%v'", originalRecipe["vegetarian"], m["vegetarian"], m["vegetarian"])
}

func TestUpdatePatchRecipeNoCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PATCH", "/v1/recipes/1", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (PATCH): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestUpdatePatchRecipeWithCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PATCH", "/v1/recipes/1", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (PATCH): %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	assert.Equalf(t, m["id"], originalRecipe["id"], "Expected the id to remain the same (%v). Got %v", originalRecipe["id"], m["id"])

	assert.NotEqualf(t, m["name"], originalRecipe["name"], "Expected the name to change from '%v' to '%v'. Got '%v'", originalRecipe["name"], m["name"], m["name"])
	assert.NotEqualf(t, m["preptime"], originalRecipe["preptime"], "Expected the prep time to change from '%v' to '%v'. Got '%v'", originalRecipe["preptime"], m["preptime"], m["preptime"])
	assert.NotEqualf(t, m["difficulty"], originalRecipe["difficulty"], "Expected the difficulty to change from '%v' to '%v'. Got '%v'", originalRecipe["difficulty"], m["difficulty"], m["difficulty"])
	assert.NotEqualf(t, m["vegetarian"], originalRecipe["vegetarian"], "Expected the vegetarian to change from '%v' to '%v'. Got '%v'", originalRecipe["vegetarian"], m["vegetarian"], m["vegetarian"])
}

func TestDeleteRecipeNoCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (DELETE): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestDeleteRecipeWithCredentials(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (GET): %s", err)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (DELETE): %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("GET", "/v1/recipes/1", nil)
	assert.Nilf(t, err, "Error on http.NewRequest (Second GET): %s", err)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	assert.Equalf(t, expected, actual, "Expected response code %d - Got %d", expected, actual)
}

func ensureTablesExist() {
	if _, err := app.DB.Exec(recipesTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := app.DB.Exec(ratingsTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	app.DB.Exec("DELETE FROM recipes")
	app.DB.Exec("ALTER SEQUENCE recipes_id_seq RESTART WITH 1")
	app.DB.Exec("DELETE FROM recipe_ratings")
	app.DB.Exec("ALTER SEQUENCE recipe_ratings_rating_id_seq RESTART WITH 1")
}

func TestAddRating(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (1st POST): %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (2nd POST): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["rating_id"], 1.0, "Expected rating ID to be '1'. Got '%v'", m["rating_id"])
}

func TestSearch(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (1st POST): %s", err)
	req.SetBasicAuth(authUser, authPassword)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (2nd POST): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["rating_id"], 1.0, "Expected rating ID to be '1'. Got '%v'", m["rating_id"])

	payload = []byte(`{"rating":2}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	assert.Nilf(t, err, "Error on http.NewRequest (3rd POST): %s", err)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["rating_id"], 2.0, "Expected rating ID to be '2'. Got '%v'", m["rating_id"])

	var bb bytes.Buffer
	mw := multipart.NewWriter(&bb)
	mw.WriteField("count", "1")
	mw.WriteField("start", "0")
	mw.WriteField("preptime", "50.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/recipes", &bb)
	assert.Nilf(t, err, "Error on http.NewRequest (4th POST): %s", err)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)
	// only want the first one
	m = mm[0]

	assert.Equalf(t, m["name"], "test recipe", "Expected recipe name to be 'test recipe'. Got '%v'", m["name"])

	assert.Equalf(t, m["preptime"], 0.1, "Expected recipe prep time to be '0.1'. Got '%v'", m["preptime"])

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["difficulty"], 2.0, "Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])

	assert.Equalf(t, m["vegetarian"], true, "Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])

	// the avg_rating is compared to 2.5 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["avg_rating"], 2.5, "Expected average recipe rating to be '2.5'. Got '%v'", m["avg_rating"])

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["id"], 1.0, "Expected recipe ID to be '1'. Got '%v'", m["id"])

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "15")
	mw.WriteField("start", "-5")
	mw.WriteField("preptime", "50.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/recipes", &bb)
	assert.Nilf(t, err, "Error on http.NewRequest (5th POST): %s", err)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)
	// only want the first one
	m = mm[0]

	assert.Equalf(t, m["name"], "test recipe", "Expected recipe name to be 'test recipe'. Got '%v'", m["name"])

	assert.Equalf(t, m["preptime"], 0.1, "Expected recipe prep time to be '0.1'. Got '%v'", m["preptime"])

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["difficulty"], 2.0, "Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])

	assert.Equalf(t, m["vegetarian"], true, "Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])

	// the avg_rating is compared to 2.5 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["avg_rating"], 2.5, "Expected average recipe rating to be '2.5'. Got '%v'", m["avg_rating"])

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	assert.Equalf(t, m["id"], 1.0, "Expected recipe ID to be '1'. Got '%v'", m["id"])

	addRecipes(12)

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/recipes", &bb)
	assert.Nilf(t, err, "Error on http.NewRequest (6th POST): %s", err)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	assert.Equalf(t, len(mm), 10, "Expected '10' recipes. Got '%v'", len(mm))

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.WriteField("preptime", "30.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/search/recipes", &bb)
	assert.Nilf(t, err, "Error on http.NewRequest (7th POST): %s", err)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	assert.Equalf(t, len(mm), 2, "Expected '2' recipes. Got '%v'", len(mm))
}

func BenchmarkCreateRateAndDeleteRecipe(b *testing.B) {
	clearTables()

	for i := 1; i < 501; i++ {
		// Create recipe
		payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

		req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
		assert.Nilf(b, err, "Error on http.NewRequest: %s", err)
		response := executeRequest(req)

		assert.Equalf(b, response.Code, http.StatusCreated, "Expected response code %d - Got %d", http.StatusCreated, response.Code)

		// Rate recipe
		payload = []byte(`{"rating":3}`)

		req, err = http.NewRequest("POST", "/v1/recipes/"+strconv.Itoa(i)+"/rating", bytes.NewBuffer(payload))
		assert.Nilf(b, err, "Error on http.NewRequest (rating): %s", err)
		response = executeRequest(req)

		assert.Equalf(b, response.Code, http.StatusCreated, "Expected response code (rating) %d - Got %d", http.StatusCreated, response.Code)

		// Delete recipe
		req, err = http.NewRequest("DELETE", "/v1/recipes/"+strconv.Itoa(i), nil)
		assert.Nilf(b, err, "Error on http.NewRequest (DELETE): %s", err)
		response = executeRequest(req)

		assert.Equalf(b, response.Code, http.StatusOK, "Expected response code %d - Got %d", http.StatusOK, response.Code)

		// Query recipe
		req, err = http.NewRequest("GET", "/v1/recipes/"+strconv.Itoa(i), nil)
		assert.Nilf(b, err, "Error on http.NewRequest (GET): %s", err)
		response = executeRequest(req)

		assert.Equalf(b, response.Code, http.StatusNotFound, "Expected response code %d - Got %d", http.StatusNotFound, response.Code)
	}
}

func addRecipes(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		app.DB.Exec("INSERT INTO recipes(name, preptime, difficulty, vegetarian) VALUES($1, $2, $3, $4)",
			"Recipe "+strconv.Itoa(i), (i+1.0)*10, i%3+1, true)
	}
}

func addRecipeRatings(recipe int, count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		addRecipeRating(recipe, i%5+1)
	}
}

func addRecipeRating(recipe int, rating int) {
	app.DB.Exec("INSERT INTO recipe_ratings(recipe_id, rating) VALUES($1, $2)", recipe, rating)
}

const recipesTableCreationQuery = `CREATE TABLE IF NOT EXISTS recipes
(
	id BIGSERIAL,
	name TEXT NOT NULL UNIQUE,
	preptime FLOAT(4) NOT NULL DEFAULT 0.0,
	difficulty NUMERIC(1) NOT NULL CHECK (difficulty > 0) CHECK (difficulty < 4) DEFAULT 0,
	vegetarian BOOLEAN NOT NULL DEFAULT false,
	CONSTRAINT recipes_pkey PRIMARY KEY (id)
)`

const ratingsTableCreationQuery = `CREATE TABLE IF NOT EXISTS recipe_ratings
(
	recipe_id BIGINT REFERENCES recipes(id) ON DELETE CASCADE,
	rating_id BIGSERIAL,
	rating SMALLINT NOT NULL CHECK (rating > 0) CHECK (rating < 6) DEFAULT 0,
	PRIMARY KEY (recipe_id, rating_id)
)`
