package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	log.Println("new message:", msg)
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

	log.Println("new message:", msgs)
	if len(msgs) > 0 {
		// Save the new mqtt messages.
		go addMessages(msgs, db)
	}
	c.IndentedJSON(http.StatusCreated, "")
}

func postMqttBatchMessageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postMqttBatchMessage(c, db)
	}
}

// addMessages adds the specified messages to the database
func addMessages(msgs []mqttMessage, db *sql.DB) {

	if len(msgs) == 0 {
		return
	}

	// Prepare an INSERT statement
	stmt, err := db.Prepare("insert into iot_messages (topic, payload) values (?, ?)")
	if err != nil {
		log.Println("insert error:", err.Error())
		return
	}
	defer stmt.Close()

	// Insert each message into the database
	for _, msg := range msgs {
		_, err := stmt.Exec(msg.Topic, msg.Payload)
		if err != nil {
			log.Println("batch error:", err.Error())
			return
		}
	}
}

// getDatabaseConnection returns the database connection
func getDatabaseConnection(dbUser, dbPass, dbHost, dbName string) (*sql.DB, error) {
	if strings.TrimSpace(dbUser) == "" {
		log.Fatal("Error: db user is empty or contains only spaces")
	}

	if strings.TrimSpace(dbPass) == "" {
		log.Fatal("Error: db pass is empty or contains only spaces")
	}

	if strings.TrimSpace(dbHost) == "" {
		log.Fatal("Error: db host is empty or contains only spaces")
	}

	if strings.TrimSpace(dbName) == "" {
		log.Fatal("Error: db name is empty or contains only spaces")
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
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	log.Println("Connected!")
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
		log.Println("Failed to load environment variables:", err)
		return
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
