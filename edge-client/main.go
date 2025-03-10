package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

type mqttMessage struct {
	Topic   string // topic
	Payload string // payload
}

var (
	mu           sync.Mutex
	mqttMessages []mqttMessage // Buffer to store messages
)

func startMqttClient(broker, clientId, topic, batchMessageApiUrl string, client mqtt.Client, ticker *time.Ticker, stopCh chan struct{}) error {

	if strings.TrimSpace(broker) == "" {
		return errors.New("Error: broker is empty or contains only spaces")
	}

	if strings.TrimSpace(clientId) == "" {
		return errors.New("Error: client id is empty or contains only spaces")
	}

	if strings.TrimSpace(topic) == "" {
		return errors.New("Error: topic is empty or contains only spaces")
	}

	if strings.TrimSpace(batchMessageApiUrl) == "" {
		return errors.New("Error: batch message api url is empty or contains only spaces")
	}

	// Connect the client
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// Process received mqtt message
	msgRcvd := mqtt.MessageHandler(func(client mqtt.Client, message mqtt.Message) {
		mu.Lock()
		defer mu.Unlock()
		log.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
		msg := mqttMessage{Topic: message.Topic(), Payload: string(message.Payload())}
		mqttMessages = append(mqttMessages, msg)
	})

	// Subscribe to the topic
	if token := client.Subscribe(topic, 0, msgRcvd); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// Goroutine to send periodic messages
	go func() {
		for {
			select {
			case <-ticker.C:
				err := sendJsonBatchRequest(batchMessageApiUrl)
				if err != nil {
					log.Println("Failed to send json batch request:", err)
					// Don't return; continue trying on the next tick
				}
			case <-stopCh: // Stop signal received
				log.Println("Stopping MQTT client...")
				ticker.Stop()
				client.Disconnect(250) // Gracefully disconnect MQTT client
				return
			}
		}
	}()

	return nil
}

// Function to send a JSON HTTP request
func sendJsonBatchRequest(batchMessageApiUrl string) error {

	if strings.TrimSpace(batchMessageApiUrl) == "" {
		return errors.New("Error: batch message api url is empty or contains only spaces")
	}

	if len(mqttMessages) == 0 {
		return nil // No messages to send, not an error
	}

	mu.Lock()
	defer mu.Unlock()

	// Convert struct to JSON
	jsonData, err := json.Marshal(mqttMessages)
	if err != nil {
		return errors.New("Error marshaling JSON")
	}

	// Send HTTP POST request
	resp, err := http.Post(batchMessageApiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.New("Error sending request")
	}
	defer resp.Body.Close()

	mqttMessages = nil
	log.Println("Response Status:", resp.Status)

	return nil
}

func getMqttClient(broker, clientId string) (mqtt.Client, error) {

	if strings.TrimSpace(broker) == "" {
		return nil, errors.New("Error: broker is empty or contains only spaces")
	}

	if strings.TrimSpace(clientId) == "" {
		return nil, errors.New("Error: client id is empty or contains only spaces")
	}

	// Configure persistent client options
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientId).
		SetCleanSession(false). // Enable persistent session
		SetKeepAlive(60).       // Keep connection alive
		SetAutoReconnect(true). // Automatically reconnect if disconnected
		SetConnectRetry(true).  // Retry connection if it fails
		SetConnectRetryInterval(5 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Println("Connected to MQTT Broker")
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Println("Lost connection to MQTT Broker:", err)
		})

	return mqtt.NewClient(opts), nil
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

func getEnvironmentVariables() (string, string, string, string, error) {
	// load env vars
	err := godotenv.Load()
	if err != nil {
		return "", "", "", "", fmt.Errorf("Error loading .env file: %w", err)
	}

	// MQTT broker address
	broker, err := getEnvVar("MQTT_BROKER_ADDR")
	if err != nil {
		return "", "", "", "", err
	}

	// MQTT client id
	clientId, err := getEnvVar("CLIENT_ID")
	if err != nil {
		return "", "", "", "", err
	}

	// MQTT message topic
	topic, err := getEnvVar("TOPIC")
	if err != nil {
		return "", "", "", "", err
	}

	// cloud api url for batch message
	batchMessageApiUrl, err := getEnvVar("BATCHMESSAGE_API_URL")
	if err != nil {
		return "", "", "", "", err
	}

	return broker, clientId, topic, batchMessageApiUrl, nil
}

func main() {
	// Load env vars
	broker, clientId, topic, batchMessageApiUrl, err := getEnvironmentVariables()
	if err != nil {
		log.Println("Failed to load environment variables:", err)
		return
	}

	// Initialize MQTT client
	client, err := getMqttClient(broker, clientId)
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}

	log.Println("MQTT client initialized successfully")

	// Create a ticker for periodic execution
	//ticker := time.NewTicker(5 * time.Minute) // Send data every 5 minutes
	ticker := time.NewTicker(15 * time.Second) // Send data every 15 seconds
	defer ticker.Stop()

	// Stop channel to signal shutdown
	stopCh := make(chan struct{})

	msg := "starting Mqtt client!"
	log.Println(msg)

	// Start the MQTT client in a goroutine
	go func() {
		err := startMqttClient(broker, clientId, topic, batchMessageApiUrl, client, ticker, stopCh)
		if err != nil {
			log.Fatalf("Error starting MQTT client: %v", err)
		}
	}()

	// Handle OS interrupt signals (CTRL+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
	<-sigCh
	log.Println("Shutdown signal received")
	close(stopCh)               // Notify startMqttClient to stop
	time.Sleep(1 * time.Second) // Give some time to clean up
	log.Println("Application exiting")
}
