package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Item struct {
    ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Name        string             `json:"name,omitempty" bson:"name,omitempty"`
    Description string             `json:"description,omitempty" bson:"description,omitempty"`
}

func connectMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }
    // Check the connection
    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, err
    }
    return client, nil
}

func createItem(ctx context.Context, w http.ResponseWriter, r *http.Request, dbName string, collectionName string) {
    w.Header().Set("Content-Type", "application/json")
    var item Item
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    collection := client.Database(dbName).Collection(collectionName)
    result, err := collection.InsertOne(ctx, item)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

// Similar modifications for getItem, getAllItems, updateItem, deleteItem using ctx for operations

func setupRouter(client *mongo.Client) *mux.Router {
    router := mux.NewRouter()

    // Wrapping the handler function to use context derived from the HTTP request
    router.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()
        createItem(ctx, w, r, "mydatabase", "items")
    }).Methods("POST")

    // Setup routes for getItem, getAllItems, updateItem, deleteItem similarly

    return router
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = connectMongoDB(ctx, "mongodb://localhost:27017")
    if err != nil {
        log.Fatal(err)
    }

    router := setupRouter(client)
    log.Fatal(http.ListenAndServe(":3000", router))
}
