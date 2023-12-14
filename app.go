package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var collection *mongo.Collection

func initMongoDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = client.Database("monggo").Collection("users")
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		id := r.URL.Query().Get("id")
		if id != "" {
			idInt := 0
			fmt.Sscanf(id, "%d", &idInt)

			var user User
			err := collection.FindOne(context.Background(), bson.M{"id": idInt}).Decode(&user)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			json.NewEncoder(w).Encode(user)
		} else {
			cur, err := collection.Find(context.Background(), bson.D{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cur.Close(context.Background())

			var users []User
			for cur.Next(context.Background()) {
				var user User
				err := cur.Decode(&user)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				users = append(users, user)
			}

			if err := cur.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(users)
		}
	case "POST":
		w.Header().Set("Content-Type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var newUser User
		err = json.Unmarshal(body, &newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newUser.ID = primitive.NewObjectID().Hex()

		_, err = collection.InsertOne(context.Background(), newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(newUser)
	}
}

func main() {
	initMongoDB()

	http.HandleFunc("/users", handleUsers)

	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", nil)
}
