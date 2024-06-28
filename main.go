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

	"gopkg.in/yaml.v2"
)

type Message struct {
	ID     string `bson:"_id,omitempty" json:"_id,omitempty"`
	Sender string `bson:"sender" json:"sender"`
	Value  string `bson:"value" json:"value"`
}

var mongoClient *mongo.Client

type Env struct {
	MongodbPort string `yaml:"mongodbPort"`
	Port        string `yaml:"port"`
	BackendUrl  string `yaml:"backendUrl"`
}

func main() {
	yamlFile, err := os.ReadFile("profile.yml")
	if err != nil {
		log.Fatalf("failed to read YAML file: %v", err)
	}
	var env Env
	err = yaml.UnmarshalStrict(yamlFile, &env)
	if err != nil {
		log.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Set MongoDB client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:" + env.MongodbPort)

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

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/api/messages", messagesHandler)
	//certFile := "cert.pem"
	//keyFile := "key.pem"
	//err = http.ListenAndServeTLS(":"+port, certFile, keyFile, nil)
	log.Println("Starting server on https://localhost:" + env.Port)
	err = http.ListenAndServe(":"+env.Port, nil)
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
	findOptions.SetLimit(8)
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
