package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testIP4Address = "192.168.1.2"
	testDomainName = "example.com"
)

func TestLWSClient_CreateDNSRecord(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		record         *DNSRecord
		expectError    bool
	}{
		{
			name:           "successful creation",
			responseBody:   `{"success": true, "data": {"id": "12345", "name": "www", "type": "A", "value": "192.168.1.1", "zone": "example.com", "ttl": 3600}, "message": "Record created"}`,
			responseStatus: http.StatusOK,
			record: &DNSRecord{
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  "example.com",
				TTL:   3600,
			},
			expectError: false,
		},
		{
			name:           "API error",
			responseBody:   `{"success": false, "error": "Invalid zone"}`,
			responseStatus: http.StatusBadRequest,
			record: &DNSRecord{
				Name:  "invalid",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  "invalid.com",
				TTL:   3600,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify HTTP method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify headers
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type: application/json")
				}
				if r.Header.Get("X-Auth-Login") != "testlogin" {
					t.Errorf("Expected X-Auth-Login header")
				}
				if r.Header.Get("X-Auth-Pass") != "testkey" {
					t.Errorf("Expected X-Auth-Pass header")
				}
				if r.Header.Get("X-Test-Mode") != "true" {
					t.Errorf("Expected X-Test-Mode header")
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create client with test server URL
			client := NewLWSClient("testlogin", "testkey", server.URL, true)

			// Test CreateDNSRecord
			record, err := client.CreateDNSRecord(context.Background(), tt.record)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if record == nil {
					t.Errorf("Expected record, got nil")
				} else if record.ID != "12345" {
					t.Errorf("Expected ID 12345, got %s", record.ID)
				}
			}
		})
	}
}

func TestLWSClient_GetDNSRecord(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		recordID       string
		expectError    bool
		expectedRecord *DNSRecord
	}{
		{
			name: "successful get",
			responseBody: `{
				"success": true,
				"data": {
					"id": "12345",
					"name": "www",
					"type": "A",
					"value": "192.168.1.1",
					"zone": "example.com",
					"ttl": 3600
				}
			}`,
			responseStatus: http.StatusOK,
			recordID:       "12345",
			expectError:    false,
			expectedRecord: &DNSRecord{
				ID:    "12345",
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  "example.com",
				TTL:   3600,
			},
		},
		{
			name:           "record not found",
			responseBody:   `{"success": false, "error": "Record not found"}`,
			responseStatus: http.StatusNotFound,
			recordID:       "nonexistent",
			expectError:    true,
			expectedRecord: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Check if URL contains the record ID
				if !strings.Contains(r.URL.Path, tt.recordID) {
					t.Errorf("Expected URL to contain record ID %s", tt.recordID)
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewLWSClient("testlogin", "testkey", server.URL, true)

			record, err := client.GetDNSRecord(context.Background(), tt.recordID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if record == nil {
					t.Errorf("Expected record, got nil")
				} else {
					if record.ID != tt.expectedRecord.ID {
						t.Errorf("Expected ID %s, got %s", tt.expectedRecord.ID, record.ID)
					}
					if record.Name != tt.expectedRecord.Name {
						t.Errorf("Expected Name %s, got %s", tt.expectedRecord.Name, record.Name)
					}
					if record.Type != tt.expectedRecord.Type {
						t.Errorf("Expected Type %s, got %s", tt.expectedRecord.Type, record.Type)
					}
				}
			}
		})
	}
}

func TestLWSClient_UpdateDNSRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "data": {"id": "12345", "name": "www", "type": "A", "value": "` + testIP4Address + `", "zone": "` + testDomainName + `", "ttl": 3600}, "message": "Record updated"}`))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	record := &DNSRecord{
		ID:    "12345",
		Name:  "www",
		Type:  "A",
		Value: testIP4Address, // Updated value
		Zone:  testDomainName,
		TTL:   3600,
	}

	updatedRecord, err := client.UpdateDNSRecord(context.Background(), record)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if updatedRecord == nil {
		t.Errorf("Expected updated record, got nil")
	} else if updatedRecord.Value != testIP4Address {
		t.Errorf("Expected updated value %s, got %s", testIP4Address, updatedRecord.Value)
	}
}

func TestLWSClient_DeleteDNSRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "message": "Record deleted"}`))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	err := client.DeleteDNSRecord(context.Background(), "12345")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLWSClient_GetDNSZone(t *testing.T) {
	responseBody := `{
		"success": true,
		"data": {
			"name": "example.com",
			"records": [
				{
					"id": "1",
					"name": "www",
					"type": "A",
					"value": "192.168.1.1",
					"zone": "example.com",
					"ttl": 3600
				},
				{
					"id": "2",
					"name": "mail",
					"type": "CNAME",
					"value": "www.example.com",
					"zone": "example.com",
					"ttl": 3600
				}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "example.com") {
			t.Errorf("Expected URL to contain zone name")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	zone, err := client.GetDNSZone(context.Background(), "example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if zone == nil {
		t.Errorf("Expected zone, got nil")
	} else {
		if zone.Name != testDomainName {
			t.Errorf("Expected zone name %s, got %s", testDomainName, zone.Name)
		}
		if len(zone.Records) != 2 {
			t.Errorf("Expected 2 records, got %d", len(zone.Records))
		}
	}
}

func TestLWSClient_Authentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login := r.Header.Get("X-Auth-Login")
		apiKey := r.Header.Get("X-Auth-Pass")

		if login != "correctlogin" || apiKey != "correctkey" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"success": false, "error": "Unauthorized"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "data": {"id": "123", "name": "test", "type": "A", "value": "1.1.1.1", "zone": "test.com", "ttl": 3600}}`))
	}))
	defer server.Close()

	// Test with correct credentials
	client := NewLWSClient("correctlogin", "correctkey", server.URL, false)
	record := &DNSRecord{Name: "test", Type: "A", Value: "1.1.1.1", Zone: "test.com", TTL: 3600}
	_, err := client.CreateDNSRecord(context.Background(), record)
	if err != nil {
		t.Errorf("Expected success with correct credentials, got error: %v", err)
	}

	// Test with incorrect credentials
	client = NewLWSClient("wronglogin", "wrongkey", server.URL, false)
	_, err = client.CreateDNSRecord(context.Background(), record)
	if err == nil {
		t.Errorf("Expected error with incorrect credentials, got success")
	}
}
 