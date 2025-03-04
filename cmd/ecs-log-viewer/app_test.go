package main

import (
	"testing"
)

func TestAppOptionValidate(t *testing.T) {
	tests := []struct {
		name        string
		appOption   AppOption
		expectError bool
	}{
		{
			name: "Valid simple format with one field",
			appOption: AppOption{
				format: "simple",
				fields: []string{"@message"},
			},
			expectError: false,
		},
		{
			name: "Invalid simple format with multiple fields",
			appOption: AppOption{
				format: "simple",
				fields: []string{"@message", "@timestamp"},
			},
			expectError: true,
		},
		{
			name: "Valid JSON format with multiple fields",
			appOption: AppOption{
				format: "json",
				fields: []string{"@message", "@timestamp"},
			},
			expectError: false,
		},
		{
			name: "Valid CSV format with multiple fields",
			appOption: AppOption{
				format: "csv",
				fields: []string{"@message", "@timestamp"},
			},
			expectError: false,
		},
		{
			name: "Invalid format",
			appOption: AppOption{
				format: "invalid",
				fields: []string{"@message"},
			},
			expectError: true,
		},
		{
			name: "Valid tail without web",
			appOption: AppOption{
				tail:   true,
				web:    false,
				format: "simple",
				fields: []string{"@message"},
			},
			expectError: false,
		},
		{
			name: "Invalid tail with web",
			appOption: AppOption{
				tail: true,
				web:  true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.appOption.validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
