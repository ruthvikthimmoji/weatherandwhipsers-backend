package main

import (
	"context"
    "log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
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
	app.Get("/api/users", getUsers) // Add the GET /users route
	app.Post("/api/users", postUsers) // Add the POST /users route

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
