package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LWSClient represents the LWS API client
type LWSClient struct {
	Login    string
	ApiKey   string
	BaseURL  string
	TestMode bool
	client   *http.Client
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl,omitempty"`
	Zone  string `json:"zone"`
}

// DNSZone represents a DNS zone
type DNSZone struct {
	Name    string      `json:"name"`
	Records []DNSRecord `json:"records,omitempty"`
}

// APIResponse represents the standard LWS API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewLWSClient creates a new LWS API client
func NewLWSClient(login, apiKey, baseURL string, testMode bool) *LWSClient {
	return &LWSClient{
		Login:    login,
		ApiKey:   apiKey,
		BaseURL:  baseURL,
		TestMode: testMode,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest makes an HTTP request to the LWS API
func (c *LWSClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Auth-Login", c.Login)
	req.Header.Set("X-Auth-Pass", c.ApiKey)

	if c.TestMode {
		req.Header.Set("X-Test-Mode", "true")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &apiResp, fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiResp.Error)
	}

	return &apiResp, nil
}

// GetDNSZone retrieves DNS zone information
func (c *LWSClient) GetDNSZone(ctx context.Context, zoneName string) (*DNSZone, error) {
	endpoint := fmt.Sprintf("dns/zone/%s", zoneName)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling zone data: %w", err)
	}

	var zone DNSZone
	if err := json.Unmarshal(dataBytes, &zone); err != nil {
		return nil, fmt.Errorf("error unmarshaling zone data: %w", err)
	}

	return &zone, nil
}

// CreateDNSRecord creates a new DNS record
func (c *LWSClient) CreateDNSRecord(ctx context.Context, record *DNSRecord) (*DNSRecord, error) {
	endpoint := "dns/record"
	resp, err := c.makeRequest(ctx, "POST", endpoint, record)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling record data: %w", err)
	}

	var createdRecord DNSRecord
	if err := json.Unmarshal(dataBytes, &createdRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling record data: %w", err)
	}

	return &createdRecord, nil
}

// GetDNSRecord retrieves a DNS record by ID
func (c *LWSClient) GetDNSRecord(ctx context.Context, recordID string) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("dns/record/%s", recordID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling record data: %w", err)
	}

	var record DNSRecord
	if err := json.Unmarshal(dataBytes, &record); err != nil {
		return nil, fmt.Errorf("error unmarshaling record data: %w", err)
	}

	return &record, nil
}

// UpdateDNSRecord updates an existing DNS record
func (c *LWSClient) UpdateDNSRecord(ctx context.Context, record *DNSRecord) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("dns/record/%s", record.ID)
	resp, err := c.makeRequest(ctx, "PUT", endpoint, record)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling record data: %w", err)
	}

	var updatedRecord DNSRecord
	if err := json.Unmarshal(dataBytes, &updatedRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling record data: %w", err)
	}

	return &updatedRecord, nil
}

// DeleteDNSRecord deletes a DNS record
func (c *LWSClient) DeleteDNSRecord(ctx context.Context, recordID string) error {
	endpoint := fmt.Sprintf("dns/record/%s", recordID)
	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("API error: %s", resp.Error)
	}

	return nil
}
