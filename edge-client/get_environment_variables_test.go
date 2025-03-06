package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to unset an environment variable
func unsetEnv(key string) {
	_ = os.Unsetenv(key)
}

// Helper function to set an environment variable
func setEnv(key, value string) {
	_ = os.Setenv(key, value)
}

func TestGetEnvironmentVariables(t *testing.T) {
	// Before testing this function, comment out below code in getEnvironmentVariables
	/*
		err := godotenv.Load()
		if err != nil {
			return "", "", "", "", fmt.Errorf("Error loading .env file: %w", err)
		}

		// This ensures the function relies only on environment variables provided by the OS or test setup.
	*/

	t.Run("Valid Environment Variables", func(t *testing.T) {
		// Set valid environment variables
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		// Call function
		broker, clientId, topic, batchMessageApiUrl, err := getEnvironmentVariables()

		// Validate results
		assert.NoError(t, err)
		assert.Equal(t, "mqtt://test-broker", broker)
		assert.Equal(t, "test-client", clientId)
		assert.Equal(t, "test-topic", topic)
		assert.Equal(t, "https://api.example.com", batchMessageApiUrl)
	})

	t.Run("Missing MQTT_BROKER_ADDR", func(t *testing.T) {
		unsetEnv("MQTT_BROKER_ADDR")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: MQTT_BROKER_ADDR environment variable is not set")
	})

	t.Run("MQTT_BROKER_ADDR is empty", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "   ")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: MQTT_BROKER_ADDR is empty or contains only spaces")
	})

	t.Run("Missing CLIENT_ID", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		unsetEnv("CLIENT_ID")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: CLIENT_ID environment variable is not set")
	})

	t.Run("CLIENT_ID is empty", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "   ")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: CLIENT_ID is empty or contains only spaces")
	})

	t.Run("Missing TOPIC", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "test-client")
		unsetEnv("TOPIC")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: TOPIC environment variable is not set")
	})

	t.Run("TOPIC is empty", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "   ")
		setEnv("BATCHMESSAGE_API_URL", "https://api.example.com")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: TOPIC is empty or contains only spaces")
	})

	t.Run("Missing BATCHMESSAGE_API_URL", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "test-topic")
		unsetEnv("BATCHMESSAGE_API_URL")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: BATCHMESSAGE_API_URL environment variable is not set")
	})

	t.Run("BATCHMESSAGE_API_URL is empty", func(t *testing.T) {
		setEnv("MQTT_BROKER_ADDR", "mqtt://test-broker")
		setEnv("CLIENT_ID", "test-client")
		setEnv("TOPIC", "test-topic")
		setEnv("BATCHMESSAGE_API_URL", "   ")

		_, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: BATCHMESSAGE_API_URL is empty or contains only spaces")
	})
}
