package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ✅ Test cases
func TestGetDatabaseConnection(t *testing.T) {
	tests := []struct {
		name          string
		dbUser        string
		dbPass        string
		dbHost        string
		dbName        string
		expectedError error
	}{
		//{"Valid Inputs", "user", "pass", "127.0.0.1:3306", "dbname", nil},
		{"Empty Db User", "", "pass", "127.0.0.1:3306", "dbname", errors.New("Error: db user is empty or contains only spaces")},
		{"Empty Db Pass", "user", "", "127.0.0.1:3306", "dbname", errors.New("Error: db pass is empty or contains only spaces")},
		{"Empty Db Host", "user", "pass", "", "dbname", errors.New("Error: db host is empty or contains only spaces")},
		{"Empty Db Name", "user", "pass", "127.0.0.1:3306", "", errors.New("Error: db name is empty or contains only spaces")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := getDatabaseConnection(tt.dbUser, tt.dbPass, tt.dbHost, tt.dbName)

			if tt.expectedError != nil {
				assert.Nil(t, db)                                   // Ensure db is nil when there's an error
				assert.EqualError(t, err, tt.expectedError.Error()) // Compare error messages
			} else {
				assert.NotNil(t, db) // Ensure db is not nil when no error
				assert.NoError(t, err)
			}
		})
	}
}
