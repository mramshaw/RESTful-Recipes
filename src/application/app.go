// Package application is a package encompassing the bulk of the application.
package application

import (
	// native packages
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	// local packages
	"recipes"
	// GitHub packages
	"github.com/julienschmidt/httprouter"
	// Standard SQL Override
	_ "github.com/lib/pq"
)

// App represents the application
type App struct {
	Router *httprouter.Router
	DB     *sql.DB
}

func (a *App) getRecipeEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}
	r := recipes.Recipe{ID: id}
	if err := r.GetRecipe(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Recipe not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) getRecipesEndpoint(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}
	recipes, err := recipes.GetRecipes(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, recipes)
}

func (a *App) createRecipeEndpoint(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var r recipes.Recipe
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := r.CreateRecipe(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, r)
}

func (a *App) modifyRecipeEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}
	var r recipes.Recipe
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&r); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	r.ID = id
	if _, err := r.UpdateRecipe(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, r)
}

func (a *App) deleteRecipeEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}
	r := recipes.Recipe{ID: id}
	if _, err := r.DeleteRecipe(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) addRatingEndpoint(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	recipeID, err := strconv.Atoi(ps.ByName("recipe_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}
	rr := recipes.RecipeRating{RecipeID: recipeID}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&rr); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer req.Body.Close()
	if err := rr.AddRecipeRating(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, rr)
}

func (a *App) searchRecipesEndpoint(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	count, _ := strconv.Atoi(req.FormValue("count"))
	start, _ := strconv.Atoi(req.FormValue("start"))

	var preptime32 float32
	if req.FormValue("preptime") == "" {
		preptime32 = 9999.99 // random large value
	} else {
		preptime64, _ := strconv.ParseFloat(req.FormValue("preptime"), 32)
		preptime32 = float32(preptime64)
	}

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	recipesRated, err := recipes.GetRecipesRated(a.DB, start, count, preptime32)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, recipesRated)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}

func basicAuth(h httprouter.Handle, requiredUser, requiredPassword string) httprouter.Handle {

	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Get the Basic Authentication credentials
		user, password, hasAuth := req.BasicAuth()

		if hasAuth && user == requiredUser && password == requiredPassword {
			// Delegate request to the given handle
			h(w, req, ps)
		} else {
			// Request Basic Authentication otherwise
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

// Initialize sets up the database connection, router, and routes for the app
func (a *App) Initialize(dbHost, dbUser, dbPassword, dbName, authUser, authPassword string) {

	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)

	var err error

	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = httprouter.New()

	a.Router.GET("/v1/recipes", a.getRecipesEndpoint)
	a.Router.POST("/v1/recipes", basicAuth(a.createRecipeEndpoint, authUser, authPassword))
	a.Router.GET("/v1/recipes/:id", a.getRecipeEndpoint)
	a.Router.PUT("/v1/recipes/:id", basicAuth(a.modifyRecipeEndpoint, authUser, authPassword))
	a.Router.PATCH("/v1/recipes/:id", basicAuth(a.modifyRecipeEndpoint, authUser, authPassword))
	a.Router.DELETE("/v1/recipes/:id", basicAuth(a.deleteRecipeEndpoint, authUser, authPassword))
	a.Router.POST("/v1/recipes/:recipe_id/rating", a.addRatingEndpoint)
	a.Router.POST("/v1/search/recipes", a.searchRecipesEndpoint)
}

// Run starts the app and serves on the specified port
func (a *App) Run(port string) {
	log.Print("Now serving recipes ...")
	log.Fatal(http.ListenAndServe(":"+port, a.Router))
}
