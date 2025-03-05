package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

/*
func startMqttClient(broker, clientId, topic, batchMessageApiUrl string) {

	if strings.TrimSpace(broker) == "" {
		log.Fatal("Error: broker is empty or contains only spaces")
	}

	if strings.TrimSpace(clientId) == "" {
		log.Fatal("Error: client id is empty or contains only spaces")
	}

	if strings.TrimSpace(topic) == "" {
		log.Fatal("Error: topic is empty or contains only spaces")
	}

	if strings.TrimSpace(batchMessageApiUrl) == "" {
		log.Fatal("Error: batch message api url is empty or contains only spaces")
	}

	// Get the MQTT client
	client := getMqttClient(broker, clientId)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Channel to signal sending
	//ticker := time.NewTicker(5 * time.Minute) // Send data every 5 minutes
	ticker := time.NewTicker(15 * time.Second) // Send data every 15 seconds
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			sendJsonBatchRequest(batchMessageApiUrl) // Send data periodically
		}
	}()

	msgRcvd := mqtt.MessageHandler(func(client mqtt.Client, message mqtt.Message) {
		mu.Lock()
		defer mu.Unlock()
		//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
		log.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
		msg := mqttMessage{Topic: message.Topic(), Payload: string(message.Payload())}
		mqttMessages = append(mqttMessages, msg)
	})

	if token := client.Subscribe(topic, 0, msgRcvd); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// Keep the program running to receive messages
	for {
	}
}
*/

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
				sendJsonBatchRequest(batchMessageApiUrl)
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
func sendJsonBatchRequest(batchMessageApiUrl string) {

	if strings.TrimSpace(batchMessageApiUrl) == "" {
		log.Fatal("Error: batch message api url is empty or contains only spaces")
	}

	if len(mqttMessages) == 0 {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Convert struct to JSON
	jsonData, err := json.Marshal(mqttMessages)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// Send HTTP POST request
	resp, err := http.Post(batchMessageApiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	mqttMessages = nil
	log.Println("Response Status:", resp.Status)
}

func getMqttClient(broker, clientId string) mqtt.Client {

	if strings.TrimSpace(broker) == "" {
		log.Fatal("Error: broker is empty or contains only spaces")
	}

	if strings.TrimSpace(clientId) == "" {
		log.Fatal("Error: client id is empty or contains only spaces")
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

	return mqtt.NewClient(opts)
}

func getEnvironmentVariables() (string, string, string, string) {
	// get env vars
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// MQTT broker address
	broker, exists := os.LookupEnv("MQTT_BROKER_ADDR")
	if !exists {
		log.Fatal("Error: MQTT_BROKER_ADDR environment variable is not set")
	}
	if strings.TrimSpace(broker) == "" {
		log.Fatal("Error: MQTT_BROKER_ADDR is empty or contains only spaces")
	}
	broker = strings.TrimSpace(broker)

	// MQTT client id
	clientId, exists := os.LookupEnv("CLIENT_ID")
	if !exists {
		log.Fatal("Error: CLIENT_ID environment variable is not set")
	}
	if strings.TrimSpace(clientId) == "" {
		log.Fatal("Error: CLIENT_ID is empty or contains only spaces")
	}
	clientId = strings.TrimSpace(clientId)

	// MQTT message topic
	topic, exists := os.LookupEnv("TOPIC")
	if !exists {
		log.Fatal("Error: TOPIC environment variable is not set")
	}
	if strings.TrimSpace(topic) == "" {
		log.Fatal("Error: TOPIC is empty or contains only spaces")
	}
	topic = strings.TrimSpace(topic)

	// cloud api url for batch message
	batchMessageApiUrl, exists := os.LookupEnv("BATCHMESSAGE_API_URL")
	if !exists {
		log.Fatal("Error: BATCHMESSAGE_API_URL environment variable is not set")
	}
	if strings.TrimSpace(batchMessageApiUrl) == "" {
		log.Fatal("Error: BATCHMESSAGE_API_URL is empty or contains only spaces")
	}
	batchMessageApiUrl = strings.TrimSpace(batchMessageApiUrl)

	return broker, clientId, topic, batchMessageApiUrl
}

func main() {
	// get env vars
	broker, clientId, topic, batchMessageApiUrl := getEnvironmentVariables()

	// Initialize MQTT client
	client := getMqttClient(broker, clientId)

	// Create a ticker for periodic execution
	//ticker := time.NewTicker(5 * time.Minute) // Send data every 5 minutes
	ticker := time.NewTicker(15 * time.Second) // Send data every 15 seconds
	defer ticker.Stop()

	// Stop channel to signal shutdown
	stopCh := make(chan struct{})

	//startMqttClient(broker, clientId, topic, batchMessageApiUrl)

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
