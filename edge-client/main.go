package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
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

func startMqtt(broker, clientId, topic, batchMessageApiUrl string) {

	// Get the MQTT client
	client := getMqttClient(broker, clientId)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Channel to signal sending
	//ticker := time.NewTicker(5 * time.Minute) // Send data every 5 minutes
	ticker := time.NewTicker(5 * time.Second) // Send data every 5 seconds
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			sendJSONBatchRequest(batchMessageApiUrl) // Send data every 5 mins
		}
	}()

	msgRcvd := mqtt.MessageHandler(func(client mqtt.Client, message mqtt.Message) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
		msg := mqttMessage{Topic: message.Topic(), Payload: string(message.Payload())}
		mqttMessages = append(mqttMessages, msg)
	})

	if token := client.Subscribe(topic, 0, msgRcvd); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	// Keep the program running to receive messages
	for {
	}
}

// Function to send a JSON HTTP request
func sendJSONRequest(data mqttMessage) {

	// Convert struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Send HTTP POST request
	resp, err := http.Post("http://localhost:8080/message", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}

// Function to send a JSON HTTP request
func sendJSONBatchRequest(batchMessageApiUrl string) {

	mu.Lock()
	defer mu.Unlock()

	if len(mqttMessages) == 0 {
		//fmt.Println("No new messages to send.")
		return
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(mqttMessages)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Send HTTP POST request
	//resp, err := http.Post("http://localhost:8080/batchmessage", "application/json", bytes.NewBuffer(jsonData))
	resp, err := http.Post(batchMessageApiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	mqttMessages = nil
	fmt.Println("Response Status:", resp.Status)
}

func getMqttClient(broker, clientId string) mqtt.Client {

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
			fmt.Println("Connected to MQTT Broker")
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			fmt.Println("Lost connection to MQTT Broker:", err)
		})

	return mqtt.NewClient(opts)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// MQTT broker address
	broker, exists := os.LookupEnv("MQTT_BROKER_ADDR")
	if !exists {
		log.Fatal("Error: MQTT_BROKER_ADDR environment variable is not set")
	}

	// MQTT client id
	clientId, exists := os.LookupEnv("CLIENT_ID")
	if !exists {
		log.Fatal("Error: CLIENT_ID environment variable is not set")
	}

	// MQTT message topic
	topic, exists := os.LookupEnv("TOPIC")
	if !exists {
		log.Fatal("Error: TOPIC environment variable is not set")
	}

	// cloud api url for batch message
	batchMessageApiUrl, exists := os.LookupEnv("BATCHMESSAGE_API_URL")
	if !exists {
		log.Fatal("Error: BATCHMESSAGE_API_URL environment variable is not set")
	}

	msg := "starting Mqtt client!"
	fmt.Println(msg)
	startMqtt(broker, clientId, topic, batchMessageApiUrl)
}
