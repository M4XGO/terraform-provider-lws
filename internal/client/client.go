package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LWSClient represents the LWS API client
type LWSClient struct {
	Login    string
	ApiKey   string
	BaseURL  string
	TestMode bool
	client   *http.Client
	retries  int
	delay    int
	backoff  int
	mu       sync.Mutex
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

// LWSAPIResponse represents the standard API response from LWS
type LWSAPIResponse struct {
	Code int         `json:"code"`
	Info interface{} `json:"info"` // Can be string or object
	Data interface{} `json:"data"`
}

// GetInfoMessage extracts a readable message from the info field
func (r *LWSAPIResponse) GetInfoMessage() string {
	switch v := r.Info.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Handle object format like {"name": "error message"}
		for _, value := range v {
			if str, ok := value.(string); ok {
				return str
			}
		}
		return fmt.Sprintf("API error (code %d)", r.Code)
	default:
		return fmt.Sprintf("API error (code %d)", r.Code)
	}
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
func NewLWSClient(login, apiKey, baseURL string, testMode bool, timeout int, retries int, delay int, backoff int) *LWSClient {
	return &LWSClient{
		Login:    login,
		ApiKey:   apiKey,
		BaseURL:  baseURL,
		TestMode: testMode,
		retries:  retries,
		delay:    delay,
		backoff:  backoff,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// makeRequest makes an HTTP request to the LWS API
func (c *LWSClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*LWSAPIResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var reqBody io.Reader
	var reqBodyBytes []byte
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBodyBytes = jsonData
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

	// Debug logging - log the request details
	log.Printf("[DEBUG] LWS API Request: %s %s", method, url)
	log.Printf("[DEBUG] Headers: X-Auth-Login=%s, X-Auth-Pass=[REDACTED], X-Test-Mode=%s",
		c.Login, req.Header.Get("X-Test-Mode"))
	if reqBodyBytes != nil {
		log.Printf("[DEBUG] Request Body: %s", string(reqBodyBytes))
	}

	var resp *http.Response
	retry := 0
	delay := c.delay
	for {
		if resp != nil {
			_ = resp.Body.Close()
		}
		log.Printf("[DEBUG] Sending request: %d/%d", retry+1, c.retries+1)
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 400 {
			break
		}
		if retry < c.retries {
			log.Printf("[DEBUG] Request error, retrying in %ds", delay)
			time.Sleep(time.Duration(delay) * time.Second)
			retry += 1
			delay *= c.backoff
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request to %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %w", url, err)
	}

	// Debug logging - log the response details
	log.Printf("[DEBUG] LWS API Response: Status %d (%s)", resp.StatusCode, resp.Status)
	log.Printf("[DEBUG] Response Headers: %v", resp.Header)
	log.Printf("[DEBUG] Response Body: %q", string(responseBody))

	// Check if response is empty
	if len(responseBody) == 0 {
		return nil, fmt.Errorf("API returned empty response (status %d) for URL: %s. This usually means the endpoint doesn't exist or authentication failed", resp.StatusCode, url)
	}

	var apiResp LWSAPIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		// Check if the response is HTML instead of JSON (common with Cloudflare challenges)
		responseStr := string(responseBody)
		if strings.Contains(responseStr, "<!DOCTYPE html>") ||
			strings.Contains(responseStr, "<html") ||
			strings.Contains(responseStr, "Just a moment...") ||
			strings.Contains(responseStr, "cloudflare") ||
			strings.Contains(responseStr, "challenge") {

			// Extract the main error info from the response if it contains JSON within
			if strings.Contains(responseStr, "Invalid response from upstream server") {
				return nil, fmt.Errorf("LWS API is temporarily protected by Cloudflare challenge system (HTTP %d). "+
					"This is usually temporary and indicates either:\n"+
					"1. High traffic or suspicious activity detected\n"+
					"2. LWS API server is having temporary issues\n"+
					"3. Rate limiting or IP blocking\n\n"+
					"Solutions:\n"+
					"- Wait a few minutes and try again\n"+
					"- Check LWS status page for service issues\n"+
					"- Contact LWS support if the issue persists\n\n"+
					"Technical details: The API returned an HTML challenge page instead of JSON response",
					resp.StatusCode)
			}

			return nil, fmt.Errorf("LWS API returned HTML challenge page instead of JSON (HTTP %d). "+
				"This indicates the API is protected by Cloudflare and requires browser-based verification. "+
				"This is usually temporary - wait a few minutes and try again. "+
				"If this persists, check LWS service status or contact support", resp.StatusCode)
		}

		// For other JSON parsing errors, provide the original detailed error
		return nil, fmt.Errorf("error unmarshaling response from %s (status %d, body: %q): %w", url, resp.StatusCode, string(responseBody), err)
	}

	// LWS API uses code 200 for success, other codes for errors
	if resp.StatusCode >= 400 || apiResp.Code != 200 {
		return &apiResp, fmt.Errorf("API error for %s (HTTP %d): Code=%d, Info=%s", url, resp.StatusCode, apiResp.Code, apiResp.GetInfoMessage())
	}

	return &apiResp, nil
}

// GetDNSZone retrieves DNS zone information
func (c *LWSClient) GetDNSZone(ctx context.Context, zoneName string) (*DNSZone, error) {
	endpoint := fmt.Sprintf("domain/%s/zdns", zoneName)
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("API error: %s", resp.GetInfoMessage())
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
	endpoint := fmt.Sprintf("domain/%s/zdns", record.Zone)

	// Prepare request body (only type, name, value, ttl)
	reqBody := CreateDNSRecordRequest{
		Type:  record.Type,
		Name:  record.Name,
		Value: record.Value,
		TTL:   record.TTL,
	}

	// Log detailed request information
	log.Printf("[DEBUG] Creating DNS record for zone '%s':", record.Zone)
	log.Printf("[DEBUG] - Endpoint: %s/%s", c.BaseURL, endpoint)
	log.Printf("[DEBUG] - Record Type: %s", reqBody.Type)
	log.Printf("[DEBUG] - Record Name: '%s'", reqBody.Name)
	log.Printf("[DEBUG] - Record Value: %s", reqBody.Value)
	log.Printf("[DEBUG] - Record TTL: %d", reqBody.TTL)
	log.Printf("[DEBUG] - Test Mode: %t", c.TestMode)

	resp, err := c.makeRequest(ctx, "POST", endpoint, reqBody)
	if err != nil {
		log.Printf("[ERROR] Failed to make request: %v", err)
		return nil, err
	}

	if resp.Code != 200 {
		log.Printf("[ERROR] API returned error code %d: %s", resp.Code, resp.GetInfoMessage())
		return nil, fmt.Errorf("API error: %s", resp.GetInfoMessage())
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

	log.Printf("[DEBUG] Successfully created record, but no ID returned from API")

	// Since the API doesn't return an ID after creation, we need to find it
	// by searching for the record we just created
	log.Printf("[DEBUG] Searching for newly created record to get its ID...")
	foundRecord, err := c.findDNSRecordByName(ctx, record.Zone, record.Name, record.Type)
	if err != nil {
		log.Printf("[ERROR] Failed to find newly created record: %v", err)
		return nil, fmt.Errorf("record was created but failed to retrieve its ID: %w", err)
	}

	// Update the created record with the found ID
	createdRecord.ID = foundRecord.ID

	log.Printf("[DEBUG] Successfully created record with ID: %d", createdRecord.ID)

	return &createdRecord, nil
}

// findDNSRecordByName finds a DNS record by name and type in the zone
func (c *LWSClient) findDNSRecordByName(ctx context.Context, zoneName, recordName, recordType string) (*DNSRecord, error) {
	// Get the entire zone first
	zone, err := c.GetDNSZone(ctx, zoneName)
	if err != nil {
		return nil, fmt.Errorf("error getting DNS zone '%s': %w", zoneName, err)
	}

	// Find the record with the matching name and type
	for _, record := range zone.Records {
		if record.Name == recordName && record.Type == recordType {
			// Set the zone since it's not in API response
			record.Zone = zoneName
			// Return a copy to avoid modifying the original slice
			foundRecord := record
			return &foundRecord, nil
		}
	}

	return nil, fmt.Errorf("DNS record with name '%s' and type '%s' not found in zone '%s'", recordName, recordType, zoneName)
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
			// Return a copy to avoid modifying the original slice
			foundRecord := record
			return &foundRecord, nil
		}
	}

	return nil, fmt.Errorf("record with ID %s not found in domain %s", recordID, domain)
}

// UpdateDNSRecord updates an existing DNS record
// The record.ID must already be set (from the state)
func (c *LWSClient) UpdateDNSRecord(ctx context.Context, record *DNSRecord) (*DNSRecord, error) {
	if record.ID == 0 {
		return nil, fmt.Errorf("record ID is required for update operation")
	}

	endpoint := fmt.Sprintf("domain/%s/zdns", record.Zone)

	// Prepare request body (id, type, name, value, ttl)
	reqBody := UpdateDNSRecordRequest{
		ID:    record.ID,
		Type:  record.Type,
		Name:  record.Name,
		Value: record.Value,
		TTL:   record.TTL,
	}

	log.Printf("[DEBUG] Updating DNS record for zone '%s':", record.Zone)
	log.Printf("[DEBUG] - Endpoint: %s/%s", c.BaseURL, endpoint)
	log.Printf("[DEBUG] - Record ID: %d", reqBody.ID)
	log.Printf("[DEBUG] - Record Type: %s", reqBody.Type)
	log.Printf("[DEBUG] - Record Name: '%s'", reqBody.Name)
	log.Printf("[DEBUG] - Record Value: %s", reqBody.Value)
	log.Printf("[DEBUG] - Record TTL: %d", reqBody.TTL)
	log.Printf("[DEBUG] - Test Mode: %t", c.TestMode)

	resp, err := c.makeRequest(ctx, "PUT", endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("API error: %s", resp.GetInfoMessage())
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

	log.Printf("[DEBUG] Successfully updated record with ID: %d", updatedRecord.ID)

	return &updatedRecord, nil
}

// DeleteDNSRecord deletes a DNS record by ID
// The recordID must be provided (from the state)
func (c *LWSClient) DeleteDNSRecord(ctx context.Context, recordID int, zoneName string) error {
	if recordID == 0 {
		return fmt.Errorf("record ID is required for delete operation")
	}

	endpoint := fmt.Sprintf("domain/%s/zdns", zoneName)

	// Prepare request body with ID
	reqBody := map[string]interface{}{
		"id": recordID,
	}

	// Log detailed request information
	log.Printf("[DEBUG] Deleting DNS record:")
	log.Printf("[DEBUG] - Endpoint: %s/%s", c.BaseURL, endpoint)
	log.Printf("[DEBUG] - Record ID: %d", recordID)
	log.Printf("[DEBUG] - Zone: %s", zoneName)
	log.Printf("[DEBUG] - Test Mode: %t", c.TestMode)

	resp, err := c.makeRequest(ctx, "DELETE", endpoint, reqBody)
	if err != nil {
		log.Printf("[ERROR] Failed to make delete request: %v", err)
		return err
	}

	// Accept both 200 and 201 as success codes for deletion
	if resp.Code != 200 && resp.Code != 201 {
		log.Printf("[ERROR] API returned error code %d: %s", resp.Code, resp.GetInfoMessage())
		return fmt.Errorf("API error: %s", resp.GetInfoMessage())
	}

	log.Printf("[DEBUG] Successfully deleted record with ID: %d", recordID)

	return nil
}

// DeleteDNSRecordByID deletes a DNS record by ID (legacy method for backward compatibility)
func (c *LWSClient) DeleteDNSRecordByID(ctx context.Context, recordID string, zoneName string) error {
	endpoint := fmt.Sprintf("domain/%s/zdns", zoneName)

	// Convert string ID to int
	recordIDInt, err := strconv.Atoi(recordID)
	if err != nil {
		return fmt.Errorf("invalid record ID '%s': %w", recordID, err)
	}

	// Prepare request body with ID
	reqBody := map[string]interface{}{
		"id": recordIDInt,
	}

	// Log detailed request information
	log.Printf("[DEBUG] Deleting DNS record by ID:")
	log.Printf("[DEBUG] - Endpoint: %s/%s", c.BaseURL, endpoint)
	log.Printf("[DEBUG] - Record ID: %d", recordIDInt)
	log.Printf("[DEBUG] - Zone: %s", zoneName)
	log.Printf("[DEBUG] - Test Mode: %t", c.TestMode)

	resp, err := c.makeRequest(ctx, "DELETE", endpoint, reqBody)
	if err != nil {
		log.Printf("[ERROR] Failed to make delete request: %v", err)
		return err
	}

	// Accept both 200 and 201 as success codes for deletion
	if resp.Code != 200 && resp.Code != 201 {
		log.Printf("[ERROR] API returned error code %d: %s", resp.Code, resp.GetInfoMessage())
		return fmt.Errorf("API error: %s", resp.GetInfoMessage())
	}

	log.Printf("[DEBUG] Successfully deleted record with ID: %d", recordIDInt)

	return nil
}
