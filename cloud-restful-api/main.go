package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gammazero/workerpool"
	"github.com/go-sql-driver/mysql"
)

type mqttMessage struct {
	Topic   string `json:"topic"`   // topic
	Payload string `json:"payload"` // payload
}

// Worker pool to control concurrency
var wp = workerpool.New(10) // Limit to 10 concurrent workers

const batchSize = 10 // Insert in batches of 10

// greeting for default page.
func greeting(c *gin.Context) {
	c.String(http.StatusOK, "Welcome, glad to have you here!")
}

// postMqttMessage adds mqtt message from JSON received in the request body.
func postMqttMessage(c *gin.Context) {
	var msg mqttMessage

	// Call BindJSON to bind the received JSON to
	// msg.
	if err := c.BindJSON(&msg); err != nil {
		return
	}

	// Save the new mqtt Message.
	log.Println("new message:", msg)

	c.JSON(http.StatusCreated, gin.H{"status": "Message queued for processing"})
}

// postMqttBatchMessage adds mqtt message from JSON received in the request body.
func postMqttBatchMessage(c *gin.Context, db *sql.DB) {
	var msgs []mqttMessage

	// Call BindJSON to bind the received JSON to
	// msgs.
	if err := c.BindJSON(&msgs); err != nil {
		return
	}

	log.Println("new message:", msgs)
	if len(msgs) > 0 {
		// Use worker pool to handle DB inserts
		wp.Submit(func() {
			// Save the new mqtt messages.
			addMessages(msgs, db)
		})
	}

	c.JSON(http.StatusCreated, gin.H{"status": "Messages queued for processing"})
}

func postMqttBatchMessageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postMqttBatchMessage(c, db)
	}
}

// insertBatch inserts a batch of messages into the database
func insertBatch(batch []mqttMessage, db *sql.DB) {
	if len(batch) == 0 {
		return
	}

	// Use a single transaction for efficiency
	tx, err := db.Begin()
	if err != nil {
		log.Println("Transaction error:", err)
		return
	}

	stmt, err := tx.Prepare("insert into iot_messages (topic, payload) values (?, ?)")
	if err != nil {
		log.Println("Prepare statement error:", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, msg := range batch {
		_, err := stmt.Exec(msg.Topic, msg.Payload)
		if err != nil {
			log.Println("Batch insert error:", err)
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("Transaction commit error:", err)
		return
	}

	log.Printf("Inserted batch of %d messages\n", len(batch))
}

// addMessages adds the specified messages to the database
func addMessages(msgs []mqttMessage, db *sql.DB) {
	// This function processes messages in batches of 10 instead of inserting them one-by-one.
	// This improves performance by reducing the number of database calls.

	/*
		msgs = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14]

		1st iteration: i=0 → batch = [1-10]
		2nd iteration: i=10 → batch = [11-14]

	*/

	if len(msgs) == 0 {
		return
	}

	// sync.WaitGroup ensures the function waits for all goroutines to finish before returning.
	var wg sync.WaitGroup

	// Iterates through the msgs slice in chunks of batchSize (10 messages at a time).
	// Handles the last batch, which may contain fewer than 10 messages.
	for i := 0; i < len(msgs); i += batchSize {
		end := i + batchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		wg.Add(1) // increments the counter before launching a new goroutine

		// Creates a new goroutine for each batch to insert messages asynchronously.
		go func(batch []mqttMessage) {
			defer wg.Done() // defer wg.Done() ensures the counter is decremented when the goroutine finishes
			insertBatch(batch, db)
		}(msgs[i:end]) // Passes the batch slice (msgs[i:end]) to insertBatch for database insertion.
	}

	wg.Wait() //Wait for All Goroutines to Finish
}

// getDatabaseConnection returns the database connection
func getDatabaseConnection(dbUser, dbPass, dbHost, dbName string) (*sql.DB, error) {
	if strings.TrimSpace(dbUser) == "" {
		return nil, errors.New("Error: db user is empty or contains only spaces")
	}

	if strings.TrimSpace(dbPass) == "" {
		return nil, errors.New("Error: db pass is empty or contains only spaces")
	}

	if strings.TrimSpace(dbHost) == "" {
		return nil, errors.New("Error: db host is empty or contains only spaces")
	}

	if strings.TrimSpace(dbName) == "" {
		return nil, errors.New("Error: db name is empty or contains only spaces")
	}

	// Capture connection properties.
	cfg := mysql.Config{
		User:                 dbUser,
		Passwd:               dbPass,
		Net:                  "tcp",
		Addr:                 dbHost,
		DBName:               dbName,
		AllowNativePasswords: true, // Enable native password authentication
	}

	// Get a database handle.
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	// Set connection pooling settings
	db.SetMaxOpenConns(25)                 // Max simultaneous open connections
	db.SetMaxIdleConns(10)                 // Max idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections after 5 min

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully!")
	return db, nil
}

// Helper function to get and validate an environment variable
func getEnvVar(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("Error: %s environment variable is not set", key)
	}
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "", fmt.Errorf("Error: %s is empty or contains only spaces", key)
	}
	return trimmedValue, nil
}

func getEnvironmentVariables() (string, string, string, string, string, error) {

	// Load env vars
	err := godotenv.Load()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("Error loading .env file: %w", err)
	}

	serverAddr, err := getEnvVar("SERVER_ADDR")
	if err != nil {
		return "", "", "", "", "", err
	}

	dbUser, err := getEnvVar("DB_USER")
	if err != nil {
		return "", "", "", "", "", err
	}

	dbPass, err := getEnvVar("DB_PASSWORD")
	if err != nil {
		return "", "", "", "", "", err
	}

	dbHost, err := getEnvVar("DB_HOST_PORT")
	if err != nil {
		return "", "", "", "", "", err
	}

	dbName, err := getEnvVar("DB_NAME")
	if err != nil {
		return "", "", "", "", "", err
	}

	return serverAddr, dbUser, dbPass, dbHost, dbName, nil
}

func main() {

	// get env vars
	serverAddr, dbUser, dbPass, dbHost, dbName, err := getEnvironmentVariables()
	if err != nil {
		log.Fatal("Failed to load environment variables:", err)
	}

	// Initialize database
	db, err := getDatabaseConnection(dbUser, dbPass, dbHost, dbName)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	router := gin.Default()
	router.GET("/", greeting)
	router.POST("/message", postMqttMessage)
	router.POST("/batchmessage", postMqttBatchMessageHandler(db))

	router.Run(serverAddr)
}
