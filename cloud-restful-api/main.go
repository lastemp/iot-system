package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/go-sql-driver/mysql"
)

type mqttMessage struct {
	Topic   string `json:"topic"`   // topic
	Payload string `json:"payload"` // payload
}

// greeting for default page.
func greeting(c *gin.Context) {
	c.String(http.StatusOK, "")
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
	fmt.Println("new message:", msg)
	c.IndentedJSON(http.StatusCreated, "")
}

// postMqttBatchMessage adds mqtt message from JSON received in the request body.
func postMqttBatchMessage(c *gin.Context, db *sql.DB) {
	var msgs []mqttMessage

	// Call BindJSON to bind the received JSON to
	// msgs.
	if err := c.BindJSON(&msgs); err != nil {
		return
	}

	// Save the new mqtt Message.
	fmt.Println("new message:", msgs)
	go addMessages(msgs, db)
	c.IndentedJSON(http.StatusCreated, "")
}

func postMqttBatchMessageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postMqttBatchMessage(c, db)
	}
}

// addMessages adds the specified messages to the database
func addMessages(msgs []mqttMessage, db *sql.DB) { //(int64, error)

	// Prepare an INSERT statement
	stmt, err := db.Prepare("INSERT INTO iot_messages (topic, payload) VALUES (?, ?)")
	if err != nil {
		fmt.Println("insert error:", err.Error())
		return
	}
	defer stmt.Close()

	// Insert each message into the database
	for _, msg := range msgs {
		_, err := stmt.Exec(msg.Topic, msg.Payload)
		if err != nil {
			fmt.Println("batch error:", err.Error())
			return
		}
	}
}

// getDatabaseConnection returns the database connection
func getDatabaseConnection(dbUser, dbPass, dbHost, dbName string) (*sql.DB, error) {
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
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected!")
	return db, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serverAddr, exists := os.LookupEnv("SERVER_ADDR")
	if !exists {
		log.Fatal("Error: SERVER_ADDR environment variable is not set")
	}

	dbUser, exists := os.LookupEnv("DB_USER")
	if !exists {
		log.Fatal("Error: DB_USER environment variable is not set")
	}

	dbPass, exists := os.LookupEnv("DB_PASSWORD")
	if !exists {
		log.Fatal("Error: DB_PASSWORD environment variable is not set")
	}

	dbHost, exists := os.LookupEnv("DB_HOST_PORT")
	if !exists {
		log.Fatal("Error: DB_HOST_PORT environment variable is not set")
	}

	dbName, exists := os.LookupEnv("DB_NAME")
	if !exists {
		log.Fatal("Error: DB_NAME environment variable is not set")
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
