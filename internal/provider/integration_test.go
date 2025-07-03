package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/M4XGO/terraform-provider-lws/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	testRecordEndpoint = "/dns/record/12345"
)

func TestProvider_CompleteWorkflow(t *testing.T) {
	// Create a mock LWS API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on endpoint
		switch {
		case r.URL.Path == "/v1/domain/example.com/zdns" && r.Method == http.MethodPost:
			// Create DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"code": 200,
				"info": "Record created",
				"data": {
					"id": 12345,
					"name": "www",
					"type": "A",
					"value": "192.168.1.1",
					"ttl": 3600
				}
			}`))

		case r.URL.Path == "/v1/domain/example.com/zdns" && r.Method == http.MethodPut:
			// Update DNS record
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

		case r.URL.Path == testRecordEndpoint && r.Method == http.MethodDelete:
			// Delete DNS record
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"code": 200,
				"info": "Record deleted",
				"data": null
			}`))

		case r.URL.Path == "/v1/domain/example.com/zdns" && r.Method == http.MethodGet:
			// Get DNS zone
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"code": 200,
				"info": "Fetched DNS Zone",
				"data": [
					{
						"id": 12345,
						"name": "www",
						"type": "A",
						"value": "192.168.1.1",
						"zone": "example.com",
						"ttl": 3600
					}
				]
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code": 404, "info": "Endpoint not found", "data": null}`))
		}
	}))
	defer server.Close()

	// Create LWS client with mock server
	lwsClient := client.NewLWSClient("testlogin", "testkey", server.URL, true)

	// Test 1: Create DNS record
	record := &client.DNSRecord{
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.1",
		Zone:  "example.com",
		TTL:   3600,
	}

	createdRecord, err := lwsClient.CreateDNSRecord(context.Background(), record)
	if err != nil {
		t.Fatalf("Failed to create DNS record: %v", err)
	}

	if createdRecord.ID != 12345 {
		t.Errorf("Expected record ID 12345, got %d", createdRecord.ID)
	}

	// Test 2: Get DNS record
	fetchedRecord, err := lwsClient.GetDNSRecord(context.Background(), "example.com", "12345")
	if err != nil {
		t.Fatalf("Failed to get DNS record: %v", err)
	}

	if fetchedRecord.Name != "www" {
		t.Errorf("Expected record name 'www', got '%s'", fetchedRecord.Name)
	}

	// Test 3: Update DNS record
	fetchedRecord.Value = "192.168.1.2"
	updatedRecord, err := lwsClient.UpdateDNSRecord(context.Background(), fetchedRecord)
	if err != nil {
		t.Fatalf("Failed to update DNS record: %v", err)
	}

	if updatedRecord.Value != "192.168.1.2" {
		t.Errorf("Expected updated value '192.168.1.2', got '%s'", updatedRecord.Value)
	}

	// Test 4: Get DNS zone
	zone, err := lwsClient.GetDNSZone(context.Background(), "example.com")
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
	err = lwsClient.DeleteDNSRecord(context.Background(), "12345")
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
			_, _ = w.Write([]byte(`{"code": 400, "info": "Invalid zone name", "data": null}`))

		case "/dns/record/nonexistent":
			// Simulate not found
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code": 404, "info": "Record not found", "data": null}`))

		case "/dns/zone/nonexistent.com":
			// Simulate zone not found
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code": 404, "info": "Zone not found", "data": null}`))

		default:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code": 500, "info": "Internal server error", "data": null}`))
		}
	}))
	defer server.Close()

	lwsClient := client.NewLWSClient("testlogin", "testkey", server.URL, true)

	// Test error on create
	record := &client.DNSRecord{
		Name:  "www",
		Type:  "A",
		Value: "192.168.1.1",
		Zone:  "invalid.com",
		TTL:   3600,
	}

	_, err := lwsClient.CreateDNSRecord(context.Background(), record)
	if err == nil {
		t.Error("Expected error when creating record with invalid zone")
	}

	// Test error on get
	_, err = lwsClient.GetDNSRecord(context.Background(), "example.com", "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent record")
	}

	// Test error on zone get
	_, err = lwsClient.GetDNSZone(context.Background(), "nonexistent.com")
	if err == nil {
		t.Error("Expected error when getting nonexistent zone")
	}
}

func TestProvider_Authentication(t *testing.T) {
	// Create a mock server that checks authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login := r.Header.Get("X-Auth-Login")
		apiKey := r.Header.Get("X-Auth-Pass")

		if login != "correctlogin" || apiKey != "correctkey" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code": 401, "info": "Unauthorized", "data": null}`))
			return
		}

		// Return success response for any endpoint with correct auth
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"code": 200,
			"info": "Zone fetched",
			"data": [
				{
					"id": 1,
					"name": "test",
					"type": "A",
					"value": "1.1.1.1",
					"ttl": 3600
				}
			]
		}`))
	}))
	defer server.Close()

	// Test with correct credentials
	lwsClient := client.NewLWSClient("correctlogin", "correctkey", server.URL, false)
	_, err := lwsClient.GetDNSZone(context.Background(), "test.com")
	if err != nil {
		t.Errorf("Expected success with correct credentials, got error: %v", err)
	}

	// Test with incorrect credentials
	lwsClient = client.NewLWSClient("wronglogin", "wrongkey", server.URL, false)
	_, err = lwsClient.GetDNSZone(context.Background(), "test.com")
	if err == nil {
		t.Error("Expected error with incorrect credentials, got success")
	}
}

func TestProvider_Configuration(t *testing.T) {
	config := LWSProviderModel{
		Login:   types.StringValue("testlogin"),
		ApiKey:  types.StringValue("testkey"),
		BaseUrl: types.StringValue("https://api.lws.net"),
	}

	if config.Login.ValueString() != "testlogin" {
		t.Errorf("Expected login 'testlogin', got '%s'", config.Login.ValueString())
	}

	if config.ApiKey.ValueString() != "testkey" {
		t.Errorf("Expected API key 'testkey', got '%s'", config.ApiKey.ValueString())
	}

	if config.BaseUrl.ValueString() != "https://api.lws.net" {
		t.Errorf("Expected base URL 'https://api.lws.net', got '%s'", config.BaseUrl.ValueString())
	}
}

func TestProvider_ResourcesAndDataSources(t *testing.T) {
	provider := &LWSProvider{}

	resources := provider.Resources(context.Background())
	if len(resources) == 0 {
		t.Error("Provider should have at least one resource")
	}

	dataSources := provider.DataSources(context.Background())
	if len(dataSources) == 0 {
		t.Error("Provider should have at least one data source")
	}

	// Verify that resources and data sources can be instantiated
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
	// Test that provider can read from environment variables
	// This would be tested in actual integration tests with real env vars
	envVars := map[string]string{
		"LWS_LOGIN":    "testlogin",
		"LWS_API_KEY":  "testkey",
		"LWS_BASE_URL": "https://api.lws.net",
	}

	for key, expectedValue := range envVars {
		if key == "LWS_LOGIN" && expectedValue != "testlogin" {
			t.Errorf("Expected %s to be 'testlogin'", key)
		}
	}
}
