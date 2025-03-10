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
			return "", "", "", "", "", fmt.Errorf("Error loading .env file: %w", err)
		}

		// This ensures the function relies only on environment variables provided by the OS or test setup.
	*/

	t.Run("Valid Environment Variables", func(t *testing.T) {
		// Set valid environment variables
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		// Call function
		serverAddr, dbUser, dbPass, dbHost, dbName, err := getEnvironmentVariables()

		// Validate results
		assert.NoError(t, err)
		assert.Equal(t, "127.0.0.1:8080", serverAddr)
		assert.Equal(t, "uid", dbUser)
		assert.Equal(t, "pwd", dbPass)
		assert.Equal(t, "127.0.0.1:3306", dbHost)
		assert.Equal(t, "mydb", dbName)
	})

	t.Run("Missing SERVER_ADDR", func(t *testing.T) {
		unsetEnv("SERVER_ADDR")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: SERVER_ADDR environment variable is not set")
	})

	t.Run("SERVER_ADDR is empty", func(t *testing.T) {
		setEnv("SERVER_ADDR", "   ")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: SERVER_ADDR is empty or contains only spaces")
	})

	t.Run("Missing DB_USER", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		unsetEnv("DB_USER")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_USER environment variable is not set")
	})

	t.Run("DB_USER is empty", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "   ")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_USER is empty or contains only spaces")
	})

	t.Run("Missing DB_PASSWORD", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		unsetEnv("DB_PASSWORD")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_PASSWORD environment variable is not set")
	})

	t.Run("DB_PASSWORD is empty", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "   ")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_PASSWORD is empty or contains only spaces")
	})

	t.Run("Missing DB_HOST_PORT", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		unsetEnv("DB_HOST_PORT")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_HOST_PORT environment variable is not set")
	})

	t.Run("DB_HOST_PORT is empty", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "   ")
		setEnv("DB_NAME", "mydb")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_HOST_PORT is empty or contains only spaces")
	})

	t.Run("Missing DB_NAME", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		unsetEnv("DB_NAME")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_NAME environment variable is not set")
	})

	t.Run("DB_NAME is empty", func(t *testing.T) {
		setEnv("SERVER_ADDR", "127.0.0.1:8080")
		setEnv("DB_USER", "uid")
		setEnv("DB_PASSWORD", "pwd")
		setEnv("DB_HOST_PORT", "127.0.0.1:3306")
		setEnv("DB_NAME", "   ")

		_, _, _, _, _, err := getEnvironmentVariables()
		assert.Error(t, err)
		assert.EqualError(t, err, "Error: DB_NAME is empty or contains only spaces")
	})
}
