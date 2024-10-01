package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Todo structure for JSON
type Todo struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

var client *mongo.Client

func main() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	r := gin.Default()
	r.Static("/public", "./public") // Serve static files from the public folder
	r.StaticFile("/", "./public/index.html")

	r.POST("/todos", createTodo)
	r.GET("/todos", fetchTodos)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	r.Run(":8080")
}

// Create a new todo
func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	todo.ID = primitive.NewObjectID().Hex()
	_, err := client.Database("testdb").Collection("todos").InsertOne(context.TODO(), todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create todo"})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// Fetch all todos
func fetchTodos(c *gin.Context) {
	cursor, err := client.Database("testdb").Collection("todos").Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch todos"})
		return
	}
	defer cursor.Close(context.TODO())

	var todos []Todo
	for cursor.Next(context.TODO()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode todo"})
			return
		}
		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, todos)
}

// Update a todo
func updateTodo(c *gin.Context) {
	id := c.Param("id")
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := client.Database("testdb").Collection("todos").UpdateOne(context.TODO(),
		bson.M{"id": id}, bson.M{"$set": bson.M{"content": todo.Content}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// Delete a todo
func deleteTodo(c *gin.Context) {
	id := c.Param("id")
	_, err := client.Database("testdb").Collection("todos").DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
