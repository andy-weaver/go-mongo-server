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

func getItem(ctx context.Context, w http.ResponseWriter, r *http.Request, dbName, collectionName string) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    var item Item
    collection := client.Database(dbName).Collection(collectionName)
    if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(item)
}

func getAllItems(ctx context.Context, w http.ResponseWriter, r *http.Request, dbName string, collectionName string) {
    w.Header().Set("Content-Type", "application/json")
    
    var items []Item
    collection := client.Database(dbName).Collection(collectionName)
    
    // Finding all documents in the collection
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)
    
    // Iterating through the cursor and decoding each document into the Item struct
    for cursor.Next(ctx) {
        var item Item
        if err := cursor.Decode(&item); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        items = append(items, item)
    }
    
    // Check if there was an error during iteration
    if err := cursor.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Encode the slice of items as JSON and send it in the response
    json.NewEncoder(w).Encode(items)
}

func updateItem(ctx context.Context, w http.ResponseWriter, r *http.Request, dbName string, collectionName string) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var item Item
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    collection := client.Database(dbName).Collection(collectionName)
    filter := bson.M{"_id": id}
    update := bson.M{"$set": item}

    result, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}

func deleteItem(ctx context.Context, w http.ResponseWriter, r *http.Request, dbName string, collectionName string) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    collection := client.Database(dbName).Collection(collectionName)
    filter := bson.M{"_id": id}

    result, err := collection.DeleteOne(ctx, filter)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}

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
