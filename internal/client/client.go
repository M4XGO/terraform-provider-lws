package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl,omitempty"`
	Zone  string `json:"zone,omitempty"`
}

// DNSZone represents a DNS zone
type DNSZone struct {
	Name    string      `json:"name"`
	Records []DNSRecord `json:"records,omitempty"`
}

// LWSAPIResponse represents the actual LWS API response format
type LWSAPIResponse struct {
	Code int         `json:"code"`
	Info string      `json:"info"`
	Data interface{} `json:"data"`
}

// CreateDNSRecordRequest represents the request body for creating a DNS record
type CreateDNSRecordRequest struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// UpdateDNSRecordRequest represents the request body for updating a DNS record
type UpdateDNSRecordRequest struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
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
func (c *LWSClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*LWSAPIResponse, error) {
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

	// Check if response is empty
	if len(responseBody) == 0 {
		return nil, fmt.Errorf("API returned empty response (status %d)", resp.StatusCode)
	}

	var apiResp LWSAPIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response (body: %q): %w", string(responseBody), err)
	}

	// LWS API uses code 200 for success, other codes for errors
	if resp.StatusCode >= 400 || apiResp.Code != 200 {
		return &apiResp, fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiResp.Info)
	}

	return &apiResp, nil
}

// GetDNSZone retrieves DNS zone information
func (c *LWSClient) GetDNSZone(ctx context.Context, zoneName string) (*DNSZone, error) {
	endpoint := fmt.Sprintf("v1/domain/%s/zdns", zoneName)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Info)
	}

	// For DNS zone, the data is an array of records
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling zone data: %w", err)
	}

	var records []DNSRecord
	if err := json.Unmarshal(dataBytes, &records); err != nil {
		return nil, fmt.Errorf("error unmarshaling zone records: %w", err)
	}

	zone := &DNSZone{
		Name:    zoneName,
		Records: records,
	}

	return zone, nil
}

// CreateDNSRecord creates a new DNS record
func (c *LWSClient) CreateDNSRecord(ctx context.Context, record *DNSRecord) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("v1/domain/%s/zdns", record.Zone)

	// Prepare request body (only type, name, value, ttl)
	reqBody := CreateDNSRecordRequest{
		Type:  record.Type,
		Name:  record.Name,
		Value: record.Value,
		TTL:   record.TTL,
	}

	resp, err := c.makeRequest(ctx, "POST", endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Info)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling record data: %w", err)
	}

	var createdRecord DNSRecord
	if err := json.Unmarshal(dataBytes, &createdRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling record data: %w", err)
	}

	// Set the zone since it's not in API response
	createdRecord.Zone = record.Zone

	return &createdRecord, nil
}

// GetDNSRecord retrieves a DNS record by ID from a specific domain
func (c *LWSClient) GetDNSRecord(ctx context.Context, domain, recordID string) (*DNSRecord, error) {
	// Get the entire zone first
	zone, err := c.GetDNSZone(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Find the record with the matching ID
	recordIDInt := 0
	if recordID != "" {
		if id, err := strconv.Atoi(recordID); err == nil {
			recordIDInt = id
		}
	}

	for _, record := range zone.Records {
		if record.ID == recordIDInt {
			// Set the zone since it's not in API response
			record.Zone = domain
			return &record, nil
		}
	}

	return nil, fmt.Errorf("record with ID %s not found in domain %s", recordID, domain)
}

// UpdateDNSRecord updates an existing DNS record
func (c *LWSClient) UpdateDNSRecord(ctx context.Context, record *DNSRecord) (*DNSRecord, error) {
	endpoint := fmt.Sprintf("v1/domain/%s/zdns", record.Zone)

	// Prepare request body (id, type, name, value, ttl)
	reqBody := UpdateDNSRecordRequest{
		ID:    record.ID,
		Type:  record.Type,
		Name:  record.Name,
		Value: record.Value,
		TTL:   record.TTL,
	}

	resp, err := c.makeRequest(ctx, "PUT", endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Info)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling record data: %w", err)
	}

	var updatedRecord DNSRecord
	if err := json.Unmarshal(dataBytes, &updatedRecord); err != nil {
		return nil, fmt.Errorf("error unmarshaling record data: %w", err)
	}

	// Set the zone since it's not in API response
	updatedRecord.Zone = record.Zone

	return &updatedRecord, nil
}

// DeleteDNSRecord deletes a DNS record
func (c *LWSClient) DeleteDNSRecord(ctx context.Context, recordID string) error {
	endpoint := fmt.Sprintf("dns/record/%s", recordID)
	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		return fmt.Errorf("API error: %s", resp.Info)
	}

	return nil
}
