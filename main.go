package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	ID     string `bson:"_id,omitempty" json:"_id,omitempty"`
	Sender string `bson:"sender" json:"sender"`
	Value  string `bson:"value" json:"value"`
}

var mongoClient *mongo.Client

type Env struct {
	mongodbPort string
}

var envDev = Env{
	mongodbPort: "27010",
}

var envProd = Env{
	mongodbPort: "27017",
}

var profile = os.Args[1]

func main() {
	var env Env
	if profile == "local" {
		env = envDev
	} else {
		env = envProd
	}

	// Set MongoDB client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:" + env.mongodbPort)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	mongoClient = client

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	port := "65000"
	if len(os.Args) >= 3 {
		port = os.Args[2]
	}
	http.Handle("/chat/", http.StripPrefix("/chat/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/api/messages", messagesHandler)
	certFile := "cert.pem"
	keyFile := "key.pem"
	log.Println("Starting server on https://localhost:" + port)
	err = http.ListenAndServeTLS(":"+port, certFile, keyFile, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getMessagesApi(w, r)
		return
	} else if r.Method == http.MethodPost {
		addMessageApi(w, r)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func getMessagesApi(w http.ResponseWriter, _ *http.Request) {

	messages := findAll()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
	}
}

func addMessageApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addMessage(msg)
	w.WriteHeader(http.StatusCreated)
}

func addMessage(message Message) *mongo.InsertOneResult {
	// Specify the database and collection
	collection := mongoClient.Database("bruhChat").Collection("messages")

	// Insert the new user into the collection
	insertResult, err := collection.InsertOne(context.TODO(), message)
	if err != nil {
		log.Fatal(err)
	}

	return insertResult
}

func findAll() []Message {
	// Specify the database and collection
	collection := mongoClient.Database("bruhChat").Collection("messages")

	findOptions := options.Find()
	findOptions.SetLimit(3)
	findOptions.SetSort(bson.D{{"_id", -1}}) // Sort by _id in descending order

	cursor, err := collection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	defer cursor.Close(context.TODO())

	// Iterate through the cursor and decode each document
	var results []Message
	for cursor.Next(context.TODO()) {
		var message Message
		err := cursor.Decode(&message)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, message)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}
