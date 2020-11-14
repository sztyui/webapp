package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	dbengine "github.com/sztyui/webapp/dbengine"
	views "github.com/sztyui/webapp/views"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

// Handling requests
func handleRequests(port string) {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", views.HomePage)
	r.HandleFunc("/all", views.ReturnAllArticles)
	r.HandleFunc("/articles", views.ReturnAllArticles)
	r.HandleFunc("/article/{id}", views.ReturnSingleArticle).Methods("GET")
	r.HandleFunc("/article", views.CreateNewArticle).Methods("POST")
	r.HandleFunc("/article/{id}", views.DeleteArticle).Methods("DELETE")
	r.HandleFunc("/article/{id}", views.UpdateArticle).Methods("PUT")
	log.Fatal(http.ListenAndServe(port, r))
}

// Main
func main() {
	fmt.Println("Rest API v2.0 - Mux Routers")
	username := os.Getenv("MONGODB_USERNAME")
	password := os.Getenv("MONGODB_PASSWORD")
	database := os.Getenv("MONGODB_DBNAME")
	port := os.Getenv("WEBSERVER_PORT")

	if len(username) == 0 || len(password) == 0 {
		log.Fatal("Please set the following environment variables: MONGODB_USERNAME, MONGODB_PASSWORD")
	}

	// Connecting to the good, old mongodb :)
	ctx, _ := context.WithTimeout(context.Background(), 45*time.Second)
	connectionString := fmt.Sprintf("mongodb+srv://%s:%s@%s", username, password, database)
	clientOptions := options.Client().ApplyURI(connectionString)
	dbengine.Client, _ = mongo.Connect(ctx, clientOptions)
	fmt.Println("MongoDB Successfully connected!")
	defer dbengine.Client.Disconnect(ctx)

	// Starting the webserver.
	fmt.Printf("Webserver running on port %v...\n", port)
	handleRequests(port)
}
