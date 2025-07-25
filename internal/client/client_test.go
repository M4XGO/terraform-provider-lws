package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testDomainName = "example.com"
	testIP4Address = "192.168.1.2"
)

func TestLWSClient_CreateDNSRecord(t *testing.T) {
	tests := []struct {
		name           string
		createResponse string
		getResponse    string
		responseStatus int
		record         *DNSRecord
		expectError    bool
		expectedRecord *DNSRecord
	}{
		{
			name: "successful creation",
			createResponse: `{
				"code": 200, 
				"info": "Added a new line in the DNS Zone", 
				"data": {
					"type": "A",
					"name": "www",
					"value": "192.168.1.1",
					"ttl": 3600
				}
			}`,
			getResponse: `{
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"ttl": 3600
					}
				]
			}`,
			responseStatus: http.StatusOK,
			record: &DNSRecord{
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  testDomainName,
				TTL:   3600,
			},
			expectError: false,
			expectedRecord: &DNSRecord{
				ID:    12345,
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  testDomainName,
				TTL:   3600,
			},
		},
		{
			name: "API error during creation",
			createResponse: `{
				"code": 400, 
				"info": "Invalid zone", 
				"data": null
			}`,
			getResponse:    "",
			responseStatus: http.StatusBadRequest,
			record: &DNSRecord{
				Name:  "invalid",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  testDomainName,
				TTL:   3600,
			},
			expectError:    true,
			expectedRecord: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				if requestCount == 1 {
					// First request should be POST to create the record
					if r.Method != http.MethodPost {
						t.Errorf("Expected POST request for creation, got %s", r.Method)
					}

					// Check if URL contains the zone name
					if !strings.Contains(r.URL.Path, tt.record.Zone) {
						t.Errorf("Expected URL to contain zone name %s", tt.record.Zone)
					}

					w.WriteHeader(tt.responseStatus)
					_, _ = w.Write([]byte(tt.createResponse))
				} else if requestCount == 2 {
					// Second request should be GET to retrieve the ID
					if r.Method != http.MethodGet {
						t.Errorf("Expected GET request for ID retrieval, got %s", r.Method)
					}

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.getResponse))
				}
			}))
			defer server.Close()

			client := NewLWSClient("testlogin", "testkey", server.URL, true)

			record, err := client.CreateDNSRecord(context.Background(), tt.record)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				// For error cases, we expect only one request (the failed POST)
				if requestCount != 1 {
					t.Errorf("Expected 1 request for error case, got %d", requestCount)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if record == nil {
					t.Errorf("Expected record, got nil")
				} else if record.ID != tt.expectedRecord.ID {
					t.Errorf("Expected ID %d, got %d", tt.expectedRecord.ID, record.ID)
				} else if record.Name != tt.expectedRecord.Name {
					t.Errorf("Expected Name %s, got %s", tt.expectedRecord.Name, record.Name)
				} else if record.Type != tt.expectedRecord.Type {
					t.Errorf("Expected Type %s, got %s", tt.expectedRecord.Type, record.Type)
				}
				// For success cases, we expect 2 requests (POST then GET)
				if requestCount != 2 {
					t.Errorf("Expected 2 requests (POST then GET), got %d", requestCount)
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
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"ttl": 3600
					}
				]
			}`,
			responseStatus: http.StatusOK,
			recordID:       "12345",
			expectError:    false,
			expectedRecord: &DNSRecord{
				ID:    12345,
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  "example.com",
				TTL:   3600,
			},
		},
		{
			name:           "record not found",
			responseBody:   `{"code": 404, "info": "Record not found", "data": null}`,
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

				// Check if URL contains the zone name
				if !strings.Contains(r.URL.Path, "example.com") {
					t.Errorf("Expected URL to contain zone name")
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewLWSClient("testlogin", "testkey", server.URL, true)

			record, err := client.GetDNSRecord(context.Background(), "example.com", tt.recordID)

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
						t.Errorf("Expected ID %d, got %d", tt.expectedRecord.ID, record.ID)
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
		// Only expect PUT request now, no GET needed
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request for update, got %s", r.Method)
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"code": 200, 
			"info": "Record updated", 
			"data": {
				"id": 12345, 
				"name": "www", 
				"type": "A", 
				"value": "192.168.1.2", 
				"zone": "example.com", 
				"ttl": 3600
			}
		}`))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	record := &DNSRecord{
		ID:    12345, // ID must be provided for update
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.2",
		Zone:  testDomainName,
		TTL:   3600,
	}

	updatedRecord, err := client.UpdateDNSRecord(context.Background(), record)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if updatedRecord.Value != "192.168.1.2" {
		t.Errorf("Expected updated value '192.168.1.2', got '%s'", updatedRecord.Value)
	}
}

func TestLWSClient_DeleteDNSRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only expect DELETE request now, no GET needed
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 200, "info": "Record deleted", "data": null}`))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	err := client.DeleteDNSRecord(context.Background(), 12345, "example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Test for the legacy DeleteDNSRecordByID method
func TestLWSClient_DeleteDNSRecordByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check if URL contains the zone name and correct endpoint
		if !strings.Contains(r.URL.Path, "example.com") {
			t.Errorf("Expected URL to contain zone name")
		}
		if !strings.Contains(r.URL.Path, "/domain/example.com/zdns") {
			t.Errorf("Expected URL to match DELETE endpoint pattern")
		}

		// Check request body contains ID
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "12345") {
			t.Errorf("Expected request body to contain record ID")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 200, "info": "Record deleted", "data": null}`))
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	err := client.DeleteDNSRecordByID(context.Background(), "12345", "example.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Test for the new findDNSRecordByName method
func TestLWSClient_findDNSRecordByName(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		recordName     string
		recordType     string
		expectError    bool
		expectedRecord *DNSRecord
	}{
		{
			name: "successful find",
			responseBody: `{
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"ttl": 3600
					},
					{
						"id": 12346,
						"name": "mail",
						"type": "CNAME",
						"value": "www.example.com",
						"ttl": 3600
					}
				]
			}`,
			responseStatus: http.StatusOK,
			recordName:     "www",
			recordType:     "A",
			expectError:    false,
			expectedRecord: &DNSRecord{
				ID:    12345,
				Name:  "www",
				Type:  "A",
				Value: "192.168.1.1",
				Zone:  "example.com",
				TTL:   3600,
			},
		},
		{
			name: "record not found",
			responseBody: `{
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"ttl": 3600
					}
				]
			}`,
			responseStatus: http.StatusOK,
			recordName:     "nonexistent",
			recordType:     "A",
			expectError:    true,
			expectedRecord: nil,
		},
		{
			name: "record found but wrong type",
			responseBody: `{
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"ttl": 3600
					}
				]
			}`,
			responseStatus: http.StatusOK,
			recordName:     "www",
			recordType:     "CNAME",
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

				// Check if URL contains the zone name
				if !strings.Contains(r.URL.Path, "example.com") {
					t.Errorf("Expected URL to contain zone name")
				}

				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewLWSClient("testlogin", "testkey", server.URL, true)

			record, err := client.findDNSRecordByName(context.Background(), "example.com", tt.recordName, tt.recordType)

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
						t.Errorf("Expected ID %d, got %d", tt.expectedRecord.ID, record.ID)
					}
					if record.Name != tt.expectedRecord.Name {
						t.Errorf("Expected Name %s, got %s", tt.expectedRecord.Name, record.Name)
					}
					if record.Type != tt.expectedRecord.Type {
						t.Errorf("Expected Type %s, got %s", tt.expectedRecord.Type, record.Type)
					}
					if record.Zone != tt.expectedRecord.Zone {
						t.Errorf("Expected Zone %s, got %s", tt.expectedRecord.Zone, record.Zone)
					}
				}
			}
		})
	}
}

func TestLWSClient_GetDNSZone(t *testing.T) {
	responseBody := `{
		"code": 200,
		"info": "Fetched DNS Zone",
		"data": [
			{
				"id": 1,
				"name": "www",
				"type": "A",
				"value": "192.168.1.1",
				"ttl": 3600
			},
			{
				"id": 2,
				"name": "mail",
				"type": "CNAME",
				"value": "www.example.com",
				"ttl": 3600
			}
		]
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
			_, _ = w.Write([]byte(`{"code": 401, "info": "Unauthorized", "data": null}`))
			return
		}

		if r.Method == "POST" {
			// For POST (create), return creation response without ID
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"code": 200, 
				"info": "Added a new line in the DNS Zone", 
				"data": {
					"name": "test", 
					"type": "A", 
					"value": "1.1.1.1", 
					"ttl": 3600
				}
			}`))
		} else if r.Method == "GET" {
			// For GET (zone retrieval), return zone format with ID
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"code": 200, 
				"info": "Fetched DNS Zone", 
				"data": [
					{
						"id": 123, 
						"name": "test", 
						"type": "A", 
						"value": "1.1.1.1", 
						"ttl": 3600
					}
				]
			}`))
		}
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

func TestLWSClient_UpdateDNSRecord_RequiresID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Should not make any request when ID is missing")
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	record := &DNSRecord{
		ID:    0, // Missing ID
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.2",
		Zone:  testDomainName,
		TTL:   3600,
	}

	_, err := client.UpdateDNSRecord(context.Background(), record)
	if err == nil {
		t.Errorf("Expected error when ID is missing, got nil")
	}

	expectedErrorMsg := "record ID is required for update operation"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestLWSClient_DeleteDNSRecord_RequiresID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Should not make any request when ID is missing")
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	err := client.DeleteDNSRecord(context.Background(), 0, "example.com")
	if err == nil {
		t.Errorf("Expected error when ID is missing, got nil")
	}

	expectedErrorMsg := "record ID is required for delete operation"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}
