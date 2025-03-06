package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ✅ Test cases
func TestGetMqttClient(t *testing.T) {
	tests := []struct {
		name          string
		broker        string
		clientId      string
		expectedError error
	}{
		{"Valid Inputs", "tcp://broker:1883", "client1", nil},
		{"Empty Broker", "", "client1", errors.New("Error: broker is empty or contains only spaces")},
		{"Empty Client ID", "tcp://broker:1883", "", errors.New("Error: client id is empty or contains only spaces")},
		{"Both Empty", "", "", errors.New("Error: broker is empty or contains only spaces")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := getMqttClient(tt.broker, tt.clientId)

			if tt.expectedError != nil {
				assert.Nil(t, client)                               // Ensure client is nil when there's an error
				assert.EqualError(t, err, tt.expectedError.Error()) // Compare error messages
			} else {
				assert.NotNil(t, client) // Ensure client is not nil when no error
				assert.NoError(t, err)
			}
		})
	}
}
