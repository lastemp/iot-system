package main

import (
	"errors"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

// Mock Token
type mockToken struct {
	err  error
	done chan struct{}
}

// Implement mqtt.Token interface
func (m *mockToken) Wait() bool                     { return true }
func (m *mockToken) WaitTimeout(time.Duration) bool { return true }
func (m *mockToken) Done() <-chan struct{} {
	close(m.done) // Ensure channel is closed
	return m.done
}
func (m *mockToken) Error() error { return m.err }

// Mock MQTT Client
type mockMqttClient struct {
	connectError   error
	subscribeError error
}

// SubscribeMultiple implements mqtt.Client.
func (m *mockMqttClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	panic("unimplemented")
}

// Implement mqtt.Client interface (only required methods for this test)
func (m *mockMqttClient) Connect() mqtt.Token {
	return &mockToken{err: m.connectError, done: make(chan struct{})}
}

func (m *mockMqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return &mockToken{err: m.subscribeError, done: make(chan struct{})}
}

func (m *mockMqttClient) Disconnect(quiesce uint) {}

// ✅ **Implement missing method**
func (m *mockMqttClient) AddRoute(topic string, callback mqtt.MessageHandler) {}

// Other methods required by mqtt.Client (empty implementations for testing)
func (m *mockMqttClient) IsConnected() bool      { return true }
func (m *mockMqttClient) IsConnectionOpen() bool { return true }
func (m *mockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return &mockToken{done: make(chan struct{})}
}
func (m *mockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockToken{done: make(chan struct{})}
}
func (m *mockMqttClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}

/*
func TestStartMqttClient(t *testing.T) {
	mockClient := &mockMqttClient{} // ✅ Now valid as mqtt.Client

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stopCh := make(chan struct{})
	defer close(stopCh)

	// Call function under test
	err := startMqttClient("tcp://localhost:1883", "testClient", "test/topic", "http://localhost/api", mockClient, ticker, stopCh)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Simulate a stop signal
	stopCh <- struct{}{}

	// Allow some time for graceful shutdown
	time.Sleep(500 * time.Millisecond)

	t.Log("Test completed successfully")
}
*/

// ✅ Test cases
func TestStartMqttClient(t *testing.T) {
	tests := []struct {
		name               string
		broker             string
		clientId           string
		topic              string
		batchMessageApiUrl string
		mockClient         *mockMqttClient
		expectedError      error
	}{
		{"Valid Inputs", "tcp://broker:1883", "client1", "topic1", "http://api.com", &mockMqttClient{}, nil},
		{"Empty Broker", "", "client1", "topic1", "http://api.com", &mockMqttClient{}, errors.New("Error: broker is empty or contains only spaces")},
		{"Empty Client ID", "tcp://broker:1883", "", "topic1", "http://api.com", &mockMqttClient{}, errors.New("Error: client id is empty or contains only spaces")},
		{"Empty Topic", "tcp://broker:1883", "client1", "", "http://api.com", &mockMqttClient{}, errors.New("Error: topic is empty or contains only spaces")},
		{"Empty API URL", "tcp://broker:1883", "client1", "topic1", "", &mockMqttClient{}, errors.New("Error: batch message api url is empty or contains only spaces")},
		{"MQTT Connection Failure", "tcp://broker:1883", "client1", "topic1", "http://api.com", &mockMqttClient{connectError: errors.New("connection failed")}, errors.New("connection failed")},
		{"MQTT Subscription Failure", "tcp://broker:1883", "client1", "topic1", "http://api.com", &mockMqttClient{subscribeError: errors.New("subscription failed")}, errors.New("subscription failed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			stopCh := make(chan struct{})

			err := startMqttClient(tt.broker, tt.clientId, tt.topic, tt.batchMessageApiUrl, tt.mockClient, ticker, stopCh)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			close(stopCh) // Stop the goroutine
		})
	}
}
