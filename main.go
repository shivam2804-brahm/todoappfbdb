package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var ctx context.Context
var client *firestore.Client

const collectionName = "todos"

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

func init() {
	ctx = context.Background()

	opt := option.WithCredentialsFile("C:\\golang coding\\Go workSpace\todo-app\\credential\\todoapp-402607-1c509-firebase-adminsdk-qof1e-53bff3e44c.json") // Replace with the path to your Firebase credentials file

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}
}

func createTodo(c *gin.Context) {
	var t Todo
	if err := c.BindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode request body"})
		return
	}
	if t.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The title is required"})
		return
	}

	t.ID = generateUUID()
	t.CreatedAt = time.Now()

	_, _, err := client.Collection(collectionName).Add(ctx, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding todo to Firestore"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Todo created successfully", "todo": t})
}

func fetchTodo(c *gin.Context) {
	todos := []Todo{}
	iter := client.Collection(collectionName).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching todo items"})
			return
		}

		var t Todo
		doc.DataTo(&t)
		todos = append(todos, t)
	}

	c.JSON(http.StatusOK, gin.H{"data": todos})
}

func updateTodo(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var updatedTodo Todo

	if err := c.BindJSON(&updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode request body"})
		return
	}

	_, err := client.Collection(collectionName).Doc(id).Set(ctx, updatedTodo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating todo in Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully", "todo": updatedTodo})
}

func deleteTodo(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	_, err := client.Collection(collectionName).Doc(id).Delete(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting todo from Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func generateUUID() string {
	// Implement your own logic to generate a unique ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	r := gin.Default()

	r.POST("/todos", createTodo)
	r.GET("/todos", fetchTodo)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	port := "8080"
	r.Run(":" + port)
}
