package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test sendJsonBatchRequest function
func TestSendJsonBatchRequest(t *testing.T) {
	tests := []struct {
		name               string
		batchMessageApiUrl string
		messages           []mqttMessage
		mockServer         bool
		expectedErr        error
	}{
		{
			name:               "Valid Request",
			batchMessageApiUrl: "/valid",
			messages:           []mqttMessage{{Topic: "test", Payload: "message"}},
			mockServer:         true,
			expectedErr:        nil,
		},
		{
			name:               "Empty API URL",
			batchMessageApiUrl: "  ",
			messages:           []mqttMessage{{Topic: "test", Payload: "message"}},
			mockServer:         false,
			expectedErr:        errors.New("Error: batch message api url is empty or contains only spaces"),
		},
		{
			name:               "No Messages to Send",
			batchMessageApiUrl: "/valid",
			messages:           []mqttMessage{},
			mockServer:         true,
			expectedErr:        nil,
		},
		{
			name:               "Invalid API URL",
			batchMessageApiUrl: "http://invalid-url",
			messages:           []mqttMessage{{Topic: "test", Payload: "message"}},
			mockServer:         false,
			expectedErr:        fmt.Errorf("Error sending request"),
		},
	}

	// Mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test messages
			mqttMessages = tt.messages

			// Use mock server URL if needed
			url := tt.batchMessageApiUrl
			if tt.mockServer {
				url = mockServer.URL + tt.batchMessageApiUrl
			}

			err := sendJsonBatchRequest(url)

			// Compare expected vs actual error
			if tt.expectedErr == nil && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			} else if tt.expectedErr != nil && err == nil {
				t.Errorf("Expected error %v, but got nil", tt.expectedErr)
			} else if tt.expectedErr != nil && err != nil && !strings.Contains(err.Error(), tt.expectedErr.Error()) {
				t.Errorf("Expected error containing %q, but got %q", tt.expectedErr.Error(), err.Error())
			}
		})
	}
}
