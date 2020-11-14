package view

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	dbengine "github.com/sztyui/webapp/dbengine"
	models "github.com/sztyui/webapp/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

var dbName string = "example_database"
var collectionName string = "articles"

// HomePage is the first page, what you get for /
func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Home page of the articles backend</h1>")
	fmt.Fprintf(w, "<h2>Use this code wherever you want! :)</h2>")
	fmt.Println("Endpoint hit: homePage")
}

// ReturnAllArticles gives you back all of the articles.
func ReturnAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: returnAllArticles")
	w.Header().Set("content-type", "application/json")
	var locArticles []models.Article
	database := dbengine.Client.Database(dbName)
	articleCollection := database.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cursor, err := articleCollection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &locArticles); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(locArticles)
}

// ReturnSingleArticle gives you back one single article
// by ID
func ReturnSingleArticle(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]
	var locArticle models.Article

	db := dbengine.Client.Database(dbName)
	articleCollection := db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := articleCollection.FindOne(ctx, bson.M{"ID": key}).Decode(&locArticle)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `", "error": true }`))
		return
	}
	json.NewEncoder(w).Encode(locArticle)
}

// CreateNewArticle for creating new articles
func CreateNewArticle(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var article models.Article
	json.Unmarshal(reqBody, &article)
	// Inserting to the database
	_ = json.NewDecoder(r.Body).Decode(&article)
	db := dbengine.Client.Database(dbName)
	articleCollection := db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	exists, _ := articleCollection.CountDocuments(ctx, bson.M{"id": article.ID})
	if exists != 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "ID exists: ` + article.ID + `"}`))
		return
	}
	result, _ := articleCollection.InsertOne(ctx, article)
	json.NewEncoder(w).Encode(result)
}

// DeleteArticle for deleting articles
func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["id"]
	db := dbengine.Client.Database(dbName)
	articleCollection := db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	result, _ := articleCollection.DeleteOne(ctx, bson.M{"ID": key})
	json.NewEncoder(w).Encode(result)
}

// UpdateArticle helps updating articles by id
func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateArticle endpoint HIT")
	key := mux.Vars(r)["id"]
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "not well formed token"}`))
		return
	}
	var article, doc models.Article
	json.Unmarshal(reqBody, &article)
	article.ID = key // :D Do not mess with me.

	db := dbengine.Client.Database(dbName)
	articleCollection := db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	filter := bson.M{"id": key}
	update := bson.M{
		"$set": article,
	}
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	result := articleCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		log.Fatal(result.Err())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "internal server error while updating"}`))
		return
	}
	decodeErr := result.Decode(&doc)
	if decodeErr != nil {
		log.Fatal(decodeErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "internal server error while decoding}`))
	}
	json.NewEncoder(w).Encode(doc)
}
