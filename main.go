package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

type Client struct {
	ID   primitive.ObjectID `bson:"_id", omitempty`
	Name string             `bson:"name"`
}

type ClientResponse struct {
	id   string
	name string
}

func main() {
	// setdb environment
	os.Setenv("DB_CONNECTION", "mongodb+srv://gosandbox:dnegc6ruM8RVXMYr@cluster0.cdepk.gcp.mongodb.net/go-sandbox?retryWrites=true&w=majority")
	fmt.Printf("hello")
	dbConnection()
	router := gin.Default()
	// config := cors.DefaultConfig()
	// router.Use(cors.New(config))
	router.POST("/api/client", addClient)
	router.GET("/api/clients", getClients)
	router.GET("/api/client/:id", getClient)
	router.PUT("/api/client/:id", editClient)
	router.DELETE("api/client/:id", deleteClient)
	router.Run()
}

func resultError(ginContext *gin.Context, err error) bool {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ginContext.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return true
		}
		ginContext.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return true
	}
	return false

}

func dbConnection() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbConnect := os.Getenv("DB_CONNECTION")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbConnect))
	if err != nil {
		log.Fatal(err)
		return
	}
	db = client.Database("go-sandbox")
}

func addClient(ginContext *gin.Context) {
	var clientForm struct {
		Name string `form:"name" binding:"required"`
	}
	err := ginContext.ShouldBind(&clientForm)
	if err != nil {
		ginContext.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}
	collection := db.Collection("clients")

	result, err := collection.InsertOne(context.TODO(), bson.M{"name": clientForm.Name})
	if err != nil {
		ginContext.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ginContext.JSON(200, result)

}

func getClients(ginContext *gin.Context) {
	collection := db.Collection("clients")
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		ginContext.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var result []Client
	if err = cur.All(ginContext, &result); err != nil {
		ginContext.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	ginContext.JSON(200, result)
}

func getClient(ginContext *gin.Context) {
	id := ginContext.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ginContext.JSON(400, "Invalid ID")
	}
	collection := db.Collection("clients")
	var result Client
	filter := bson.M{"_id": objectId}
	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if !resultError(ginContext, err) {
		ginContext.JSON(200, result)
	}
}

func editClient(ginContext *gin.Context) {
	id := ginContext.Param("id")
	var clientForm struct {
		Name string `form:"name" binding:"required"`
	}
	err := ginContext.ShouldBind(&clientForm)
	if err != nil {
		ginContext.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ginContext.JSON(400, gin.H{"status": "error", "message": "Invalid Id"})
		return
	}
	collection := db.Collection("clients")
	var result Client
	filter := bson.M{"_id": objectId}
	err = collection.FindOneAndUpdate(context.Background(), filter, bson.M{"$push": bson.M{"name": clientForm.Name}}).Decode(&result)
	if !resultError(ginContext, err) {
		ginContext.JSON(200, result)
	}
}

func deleteClient(ginContext *gin.Context) {
	id := ginContext.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ginContext.JSON(400, "Invalid ID")
		return
	}
	collection := db.Collection("clients")
	var result Client
	filter := bson.M{"_id": objectId}
	err = collection.FindOneAndDelete(context.Background(), filter).Decode(&result)
	if !resultError(ginContext, err) {
		ginContext.JSON(200, result)
	}
}
