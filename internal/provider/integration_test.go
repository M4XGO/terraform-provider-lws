package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	testRecordEndpoint = "/dns/record/rec_12345"
)

func TestProvider_CompleteWorkflow(t *testing.T) {
	// Create a mock LWS API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on endpoint
		switch {
		case r.URL.Path == "/dns/record" && r.Method == http.MethodPost:
			// Create DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": {
					"id": "rec_12345",
					"name": "www",
					"type": "A",
					"value": "192.168.1.1",
					"zone": "example.com",
					"ttl": 3600
				}
			}`))

		case r.URL.Path == testRecordEndpoint && r.Method == http.MethodGet:
			// Get DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": {
					"id": "rec_12345",
					"name": "www",
					"type": "A",
					"value": "192.168.1.1",
					"zone": "example.com",
					"ttl": 3600
				}
			}`))

		case r.URL.Path == testRecordEndpoint && r.Method == http.MethodPut:
			// Update DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": {
					"id": "rec_12345",
					"name": "www",
					"type": "A",
					"value": "192.168.1.2",
					"zone": "example.com",
					"ttl": 3600
				}
			}`))

		case r.URL.Path == testRecordEndpoint && r.Method == http.MethodDelete:
			// Delete DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"message": "Record deleted"
			}`))

		case r.URL.Path == "/dns/zone/example.com" && r.Method == http.MethodGet:
			// Get DNS zone
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": {
					"name": "example.com",
					"records": [
						{
							"id": "rec_12345",
							"name": "www",
							"type": "A",
							"value": "192.168.1.1",
							"zone": "example.com",
							"ttl": 3600
						}
					]
				}
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success": false, "error": "Endpoint not found"}`))
		}
	}))
	defer server.Close()

	// Create LWS client with mock server
	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	// Test 1: Create DNS record
	record := &DNSRecord{
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.1",
		Zone:  "example.com",
		TTL:   3600,
	}

	createdRecord, err := client.CreateDNSRecord(context.Background(), record)
	if err != nil {
		t.Fatalf("Failed to create DNS record: %v", err)
	}

	if createdRecord.ID != "rec_12345" {
		t.Errorf("Expected record ID 'rec_12345', got '%s'", createdRecord.ID)
	}

	// Test 2: Get DNS record
	fetchedRecord, err := client.GetDNSRecord(context.Background(), "rec_12345")
	if err != nil {
		t.Fatalf("Failed to get DNS record: %v", err)
	}

	if fetchedRecord.Name != "www" {
		t.Errorf("Expected record name 'www', got '%s'", fetchedRecord.Name)
	}

	// Test 3: Update DNS record
	fetchedRecord.Value = "192.168.1.2"
	updatedRecord, err := client.UpdateDNSRecord(context.Background(), fetchedRecord)
	if err != nil {
		t.Fatalf("Failed to update DNS record: %v", err)
	}

	if updatedRecord.Value != "192.168.1.2" {
		t.Errorf("Expected updated value '192.168.1.2', got '%s'", updatedRecord.Value)
	}

	// Test 4: Get DNS zone
	zone, err := client.GetDNSZone(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("Failed to get DNS zone: %v", err)
	}

	if zone.Name != "example.com" {
		t.Errorf("Expected zone name 'example.com', got '%s'", zone.Name)
	}

	if len(zone.Records) != 1 {
		t.Errorf("Expected 1 record in zone, got %d", len(zone.Records))
	}

	// Test 5: Delete DNS record
	err = client.DeleteDNSRecord(context.Background(), "rec_12345")
	if err != nil {
		t.Fatalf("Failed to delete DNS record: %v", err)
	}
}

func TestProvider_ErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/dns/record":
			// Simulate API error
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"success": false, "error": "Invalid zone name"}`))

		case "/dns/record/nonexistent":
			// Simulate not found
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success": false, "error": "Record not found"}`))

		case "/dns/zone/nonexistent.com":
			// Simulate zone not found
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"success": false, "error": "Zone not found"}`))

		default:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"success": false, "error": "Internal server error"}`))
		}
	}))
	defer server.Close()

	client := NewLWSClient("testlogin", "testkey", server.URL, true)

	// Test error on create
	record := &DNSRecord{
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.1",
		Zone:  "invalid.com",
		TTL:   3600,
	}

	_, err := client.CreateDNSRecord(context.Background(), record)
	if err == nil {
		t.Error("Expected error when creating record with invalid zone")
	}

	// Test error on get
	_, err = client.GetDNSRecord(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent record")
	}

	// Test error on zone get
	_, err = client.GetDNSZone(context.Background(), "nonexistent.com")
	if err == nil {
		t.Error("Expected error when getting nonexistent zone")
	}
}

func TestProvider_Authentication(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		apiKey      string
		expectError bool
	}{
		{
			name:        "valid credentials",
			login:       "validlogin",
			apiKey:      "validkey",
			expectError: false,
		},
		{
			name:        "invalid login",
			login:       "invalidlogin",
			apiKey:      "validkey",
			expectError: true,
		},
		{
			name:        "invalid api key",
			login:       "validlogin",
			apiKey:      "invalidkey",
			expectError: true,
		},
		{
			name:        "empty credentials",
			login:       "",
			apiKey:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server that validates credentials
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				login := r.Header.Get("X-Auth-Login")
				apiKey := r.Header.Get("X-Auth-Pass")

				if login == "validlogin" && apiKey == "validkey" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"success": true, "data": {"id": "test"}}`))
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"success": false, "error": "Unauthorized"}`))
				}
			}))
			defer server.Close()

			client := NewLWSClient(tt.login, tt.apiKey, server.URL, true)

			record := &DNSRecord{
				Name:  "test",
				Type:  "A",
				Value: "1.1.1.1",
				Zone:  "test.com",
				TTL:   3600,
			}

			_, err := client.CreateDNSRecord(context.Background(), record)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error with credentials %s/%s", tt.login, tt.apiKey)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error with valid credentials: %v", err)
				}
			}
		})
	}
}

func TestProvider_Configuration(t *testing.T) {
	p := &LWSProvider{version: "test"}

	// Test metadata
	metaResp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, metaResp)

	if metaResp.TypeName != "lws" {
		t.Errorf("Expected TypeName 'lws', got '%s'", metaResp.TypeName)
	}

	if metaResp.Version != "test" {
		t.Errorf("Expected Version 'test', got '%s'", metaResp.Version)
	}

	// Test schema
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	expectedAttrs := []string{"login", "api_key", "base_url", "test_mode"}
	for _, attr := range expectedAttrs {
		if _, exists := schemaResp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute '%s' in provider schema", attr)
		}
	}
}

func TestProvider_ResourcesAndDataSources(t *testing.T) {
	p := &LWSProvider{version: "test"}

	// Test resources
	resources := p.Resources(context.Background())
	if len(resources) == 0 {
		t.Error("Expected at least one resource")
	}

	// Test data sources
	dataSources := p.DataSources(context.Background())
	if len(dataSources) == 0 {
		t.Error("Expected at least one data source")
	}

	// Verify resource and data source can be instantiated
	for _, resourceFunc := range resources {
		resource := resourceFunc()
		if resource == nil {
			t.Error("Resource function returned nil")
		}
	}

	for _, dataSourceFunc := range dataSources {
		dataSource := dataSourceFunc()
		if dataSource == nil {
			t.Error("Data source function returned nil")
		}
	}
}

func TestProvider_EnvironmentVariables(t *testing.T) {
	// This would test environment variable configuration
	// For now, just test the basic structure

	model := LWSProviderModel{
		Login:    types.StringValue("testlogin"),
		ApiKey:   types.StringValue("testkey"),
		BaseUrl:  types.StringValue("https://api.lws.net/v1"),
		TestMode: types.BoolValue(true),
	}

	if model.Login.ValueString() != "testlogin" {
		t.Errorf("Expected login 'testlogin', got '%s'", model.Login.ValueString())
	}

	if !model.TestMode.ValueBool() {
		t.Error("Expected test mode to be true")
	}

	// Test null values
	model.Login = types.StringNull()
	if !model.Login.IsNull() {
		t.Error("Expected login to be null")
	}
}
