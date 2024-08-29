package main

import (
	"context"
    "log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func mongoConnection() *mongo.Client{
	// MongoDB connection string
	clientOptions := options.Client().ApplyURI("mongodb+srv://thimmojiruthvik:thimmojiruthvik@cluster0.hg2kz.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")

	// Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

	return client;
}

func main() {

	// Fiber instance
	app := fiber.New()

	// Routes
	app.Get("/", hello)
	app.Get("/api/users", getUsers) 
	app.Post("/api/users", postUsers) 
	app.Delete("/api/users", deleteUser) 

	// Start server
	log.Fatal(app.Listen(":3000"))
}

// Handler
func hello(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}

// Handler for GET /users
func getUsers(c *fiber.Ctx) error {
	client := mongoConnection()
	collection := client.Database("usersDB").Collection("users")

	var ctx context.Context
	results, err := collection.Find(ctx, bson.M{})
	if err != nil {
        return nil
	}

    //reading from the db in an optimal way
    defer results.Close(ctx)

	var users []map[string]interface{}
	for results.Next(ctx) {
        var singleUser map[string]interface{}
        if err = results.Decode(&singleUser); err != nil {
            return nil
        }
        users = append(users, singleUser)
    }
	// Return users as JSON
	return c.JSON(users)
}

// Handler for POST /users
func postUsers(c *fiber.Ctx) error {
    // Connect to MongoDB
    client := mongoConnection()
    collection := client.Database("usersDB").Collection("users")

    // Parse the request body into a struct or map
    var user map[string]interface{}
    if err := c.BodyParser(&user); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Failed to parse request body",
        })
    }

    // Insert the user data into the database
    insertResult, err := collection.InsertOne(context.Background(), user)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to add user to the database",
        })
    }

    // Return a success response with the inserted ID
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "User added to the database",
        "insertedID": insertResult.InsertedID,
    })
}

// Handler for DELETE /users/:id
func deleteUser(c *fiber.Ctx) error {
    // Connect to MongoDB
    client := mongoConnection()
    collection := client.Database("usersDB").Collection("users")

    // Parse the request body to get the user ID
    var requestBody struct {
        ID string `json:"_id"`
    }

    if err := c.BodyParser(&requestBody); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Failed to parse request body",
        })
    }

    // Convert the ID to an ObjectID (required for MongoDB)
    objID, err := primitive.ObjectIDFromHex(requestBody.ID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid user ID format",
        })
    }

    // Define the filter for the document to delete
    filter := bson.M{"_id": objID}

    // Delete the user from the database
    deleteResult, err := collection.DeleteOne(context.Background(), filter)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to delete user from the database",
        })
    }

    // Check if the user was found and deleted
    if deleteResult.DeletedCount == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "message": "User not found",
        })
    }

    // Return a success response
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "User successfully deleted",
    })
}
