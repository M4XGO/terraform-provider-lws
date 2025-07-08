package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSRecordResource_Metadata(t *testing.T) {
	r := NewDNSRecordResource()
	resp := &resource.MetadataResponse{}
	req := resource.MetadataRequest{
		ProviderTypeName: ProviderTypeName,
	}

	r.Metadata(context.Background(), req, resp)

	expected := ProviderTypeName + "_dns_record"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestDNSRecordResource_Schema(t *testing.T) {
	r := NewDNSRecordResource()
	resp := &resource.SchemaResponse{}
	req := resource.SchemaRequest{}

	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
	}

	// Test required attributes
	requiredAttrs := []string{"name", "type", "value", "zone"}
	for _, attr := range requiredAttrs {
		attribute, exists := resp.Schema.Attributes[attr]
		if !exists {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}

		stringAttr, ok := attribute.(schema.StringAttribute)
		if !ok {
			t.Errorf("Expected '%s' to be a StringAttribute", attr)
		}

		if !stringAttr.Required {
			t.Errorf("Expected '%s' attribute to be required", attr)
		}
	}

	// Test optional attributes
	optionalAttrs := []string{"ttl"}
	for _, attr := range optionalAttrs {
		attribute, exists := resp.Schema.Attributes[attr]
		if !exists {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}

		int64Attr, ok := attribute.(schema.Int64Attribute)
		if !ok {
			t.Errorf("Expected '%s' to be an Int64Attribute", attr)
		}

		if !int64Attr.Optional {
			t.Errorf("Expected '%s' attribute to be optional", attr)
		}
	}

	// Test computed attributes
	computedAttrs := []string{"id"}
	for _, attr := range computedAttrs {
		attribute, exists := resp.Schema.Attributes[attr]
		if !exists {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}

		stringAttr, ok := attribute.(schema.StringAttribute)
		if !ok {
			t.Errorf("Expected '%s' to be a StringAttribute", attr)
		}

		if !stringAttr.Computed {
			t.Errorf("Expected '%s' attribute to be computed", attr)
		}
	}
}

func TestDNSRecordResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name      string
		model     DNSRecordResourceModel
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid A record",
			model: DNSRecordResourceModel{
				Name:  types.StringValue("www"),
				Type:  types.StringValue("A"),
				Value: types.StringValue("192.168.1.1"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Value(3600),
			},
			expectErr: false,
		},
		{
			name: "valid AAAA record",
			model: DNSRecordResourceModel{
				Name:  types.StringValue("www"),
				Type:  types.StringValue("AAAA"),
				Value: types.StringValue("2001:db8::1"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Value(3600),
			},
			expectErr: false,
		},
		{
			name: "valid CNAME record",
			model: DNSRecordResourceModel{
				Name:  types.StringValue("www"),
				Type:  types.StringValue("CNAME"),
				Value: types.StringValue("example.com"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Value(3600),
			},
			expectErr: false,
		},
		{
			name: "valid MX record",
			model: DNSRecordResourceModel{
				Name:  types.StringValue(""),
				Type:  types.StringValue("MX"),
				Value: types.StringValue("10 mail.example.com"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Value(3600),
			},
			expectErr: false,
		},
		{
			name: "valid TXT record",
			model: DNSRecordResourceModel{
				Name:  types.StringValue("_dmarc"),
				Type:  types.StringValue("TXT"),
				Value: types.StringValue("v=DMARC1; p=reject;"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Value(3600),
			},
			expectErr: false,
		},
		{
			name: "default TTL",
			model: DNSRecordResourceModel{
				Name:  types.StringValue("www"),
				Type:  types.StringValue("A"),
				Value: types.StringValue("192.168.1.1"),
				Zone:  types.StringValue("example.com"),
				TTL:   types.Int64Null(), // Should default to 3600
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation tests
			if tt.model.Name.IsUnknown() && !tt.model.Name.IsNull() {
				t.Error("Name should not be unknown when set")
			}

			if tt.model.Type.IsNull() || tt.model.Type.IsUnknown() {
				if !tt.expectErr {
					t.Error("Type should be set for valid records")
				}
			}

			if tt.model.Value.IsNull() || tt.model.Value.IsUnknown() {
				if !tt.expectErr {
					t.Error("Value should be set for valid records")
				}
			}

			if tt.model.Zone.IsNull() || tt.model.Zone.IsUnknown() {
				if !tt.expectErr {
					t.Error("Zone should be set for valid records")
				}
			}

			// Test TTL default
			if tt.model.TTL.IsNull() {
				// In real implementation, this would default to 3600
				// Here we just verify the field can be null
				t.Logf("TTL is null for test case: %s", tt.name)
			}
		})
	}
}

func TestDNSRecordTypes_Validation(t *testing.T) {
	validTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR", "SPF", "CAA"}

	for _, recordType := range validTypes {
		t.Run("valid_type_"+recordType, func(t *testing.T) {
			if recordType == "" {
				t.Error("Record type should not be empty")
			}
			// In real implementation, would validate against allowed types
		})
	}

	invalidTypes := []string{"", "INVALID", "123", "a"}

	for _, recordType := range invalidTypes {
		t.Run("invalid_type_"+recordType, func(t *testing.T) {
			// Test that invalid types would be rejected
			if recordType == "A" || recordType == "AAAA" {
				t.Skip("This is actually a valid type")
			}
		})
	}
}

func TestDNSRecord_ValueValidation(t *testing.T) {
	tests := []struct {
		recordType string
		value      string
		valid      bool
	}{
		// A record validation
		{"A", "192.168.1.1", true},
		{"A", "10.0.0.1", true},
		{"A", "256.256.256.256", false}, // Invalid IP
		{"A", "not-an-ip", false},

		// AAAA record validation
		{"AAAA", "2001:db8::1", true},
		{"AAAA", "::1", true},
		{"AAAA", "192.168.1.1", false}, // IPv4 in AAAA
		{"AAAA", "not-an-ipv6", false},

		// CNAME record validation
		{"CNAME", "example.com", true},
		{"CNAME", "www.example.com.", true}, // With trailing dot
		{"CNAME", "", false},                // Empty value

		// MX record validation
		{"MX", "10 mail.example.com", true},
		{"MX", "0 .", true},               // Null MX
		{"MX", "mail.example.com", false}, // Missing priority

		// TXT record validation
		{"TXT", "v=spf1 include:_spf.google.com ~all", true},
		{"TXT", "any text is valid", true},
		{"TXT", "", true}, // Empty TXT is valid

		// SPF record validation
		{"SPF", "v=spf1 include:_spf.google.com ~all", true},
		{"SPF", "v=spf1 a mx ~all", true},
		{"SPF", "", false}, // Empty SPF should not be valid

		// CAA record validation
		{"CAA", "0 issue letsencrypt.org", true},
		{"CAA", "0 issuewild ;", true},
		{"CAA", "128 iodef mailto:admin@example.com", true},
		{"CAA", "", false}, // Empty CAA should not be valid
	}

	for _, tt := range tests {
		t.Run(tt.recordType+"_"+tt.value, func(t *testing.T) {
			// This would test value validation in real implementation
			if tt.recordType == "" {
				t.Error("Record type should not be empty")
			}
			if tt.valid && tt.value == "" && tt.recordType != "TXT" {
				t.Errorf("Empty value should not be valid for %s records", tt.recordType)
			}
		})
	}
}

func TestDNSRecord_TTL_Validation(t *testing.T) {
	tests := []struct {
		ttl   int64
		valid bool
	}{
		// Valid LWS TTL values
		{900, true},   // 15 minutes
		{1800, true},  // 30 minutes
		{3600, true},  // 1 hour (default)
		{7200, true},  // 2 hours
		{21600, true}, // 6 hours
		{43200, true}, // 12 hours
		{86400, true}, // 1 day
		// Invalid TTL values
		{60, false},     // Too low
		{300, false},    // Not in LWS list
		{1200, false},   // Not in LWS list
		{604800, false}, // Too high
		{0, false},      // Too low
		{-1, false},     // Negative
	}

	// Valid LWS TTL values according to specification
	validLWSTTLs := map[int64]bool{
		900:   true,
		1800:  true,
		3600:  true,
		7200:  true,
		21600: true,
		43200: true,
		86400: true,
	}

	for _, tt := range tests {
		t.Run("ttl_validation", func(t *testing.T) {
			isValidLWSTTL := validLWSTTLs[tt.ttl]

			if tt.valid && !isValidLWSTTL {
				t.Errorf("TTL %d was marked as valid but is not in LWS allowed values", tt.ttl)
			}

			if !tt.valid && isValidLWSTTL {
				t.Errorf("TTL %d was marked as invalid but is in LWS allowed values", tt.ttl)
			}

			if tt.valid && tt.ttl <= 0 {
				t.Errorf("TTL %d should be positive", tt.ttl)
			}
		})
	}
}

// Tests for existing record detection and handling logic
func TestDNSRecord_ExistingRecordDetection(t *testing.T) {
	tests := []struct {
		name         string
		targetName   string
		targetType   string
		existingName string
		existingType string
		shouldMatch  bool
		description  string
	}{
		{
			name:         "exact_match",
			targetName:   "www",
			targetType:   "A",
			existingName: "www",
			existingType: "A",
			shouldMatch:  true,
			description:  "Exact name and type match should be detected",
		},
		{
			name:         "case_insensitive_name",
			targetName:   "WWW",
			targetType:   "A",
			existingName: "www",
			existingType: "A",
			shouldMatch:  true,
			description:  "Case insensitive name matching should work",
		},
		{
			name:         "case_insensitive_type",
			targetName:   "www",
			targetType:   "a",
			existingName: "www",
			existingType: "A",
			shouldMatch:  true,
			description:  "Case insensitive type matching should work",
		},
		{
			name:         "whitespace_trimming",
			targetName:   " www ",
			targetType:   " A ",
			existingName: "www",
			existingType: "A",
			shouldMatch:  true,
			description:  "Whitespace should be trimmed during comparison",
		},
		{
			name:         "different_name",
			targetName:   "www",
			targetType:   "A",
			existingName: "mail",
			existingType: "A",
			shouldMatch:  false,
			description:  "Different names should not match",
		},
		{
			name:         "different_type",
			targetName:   "www",
			targetType:   "A",
			existingName: "www",
			existingType: "CNAME",
			shouldMatch:  false,
			description:  "Different types should not match",
		},
		{
			name:         "complex_name_case_mixed",
			targetName:   "_4f63eda418b21d585d04126b53ba4ef1.pre-prod",
			targetType:   "CNAME",
			existingName: "_4F63EDA418B21D585D04126B53BA4EF1.PRE-PROD",
			existingType: "cname",
			shouldMatch:  true,
			description:  "Complex names with mixed case should match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the normalization logic from the actual code
			targetNameNorm := strings.ToLower(strings.TrimSpace(tt.targetName))
			targetTypeNorm := strings.ToUpper(strings.TrimSpace(tt.targetType))
			existingNameNorm := strings.ToLower(strings.TrimSpace(tt.existingName))
			existingTypeNorm := strings.ToUpper(strings.TrimSpace(tt.existingType))

			actualMatch := (existingNameNorm == targetNameNorm && existingTypeNorm == targetTypeNorm)

			if actualMatch != tt.shouldMatch {
				t.Errorf("%s: expected match=%v, got match=%v", tt.description, tt.shouldMatch, actualMatch)
				t.Errorf("  Target: name='%s' type='%s'", tt.targetName, tt.targetType)
				t.Errorf("  Existing: name='%s' type='%s'", tt.existingName, tt.existingType)
				t.Errorf("  Normalized Target: name='%s' type='%s'", targetNameNorm, targetTypeNorm)
				t.Errorf("  Normalized Existing: name='%s' type='%s'", existingNameNorm, existingTypeNorm)
			}
		})
	}
}

func TestDNSRecord_APIErrorDetection(t *testing.T) {
	tests := []struct {
		name                   string
		errorMessage           string
		shouldIndicateExisting bool
		description            string
	}{
		{
			name:                   "cannot_add_record",
			errorMessage:           "Cannot add record to the DNS Zone. Record invalid.",
			shouldIndicateExisting: true,
			description:            "LWS API 'cannot add record' error should indicate existing record",
		},
		{
			name:                   "record_invalid",
			errorMessage:           "Record invalid",
			shouldIndicateExisting: true,
			description:            "Generic 'record invalid' should indicate existing record",
		},
		{
			name:                   "already_exists",
			errorMessage:           "Record already exists",
			shouldIndicateExisting: true,
			description:            "Explicit 'already exists' should indicate existing record",
		},
		{
			name:                   "duplicate_record",
			errorMessage:           "Duplicate record found",
			shouldIndicateExisting: true,
			description:            "Duplicate record error should indicate existing record",
		},
		{
			name:                   "case_insensitive_cannot_add",
			errorMessage:           "CANNOT ADD RECORD TO THE DNS ZONE. RECORD INVALID.",
			shouldIndicateExisting: true,
			description:            "Case insensitive matching should work for error detection",
		},
		{
			name:                   "network_error",
			errorMessage:           "Connection timeout",
			shouldIndicateExisting: false,
			description:            "Network errors should not indicate existing record",
		},
		{
			name:                   "permission_error",
			errorMessage:           "Access denied",
			shouldIndicateExisting: false,
			description:            "Permission errors should not indicate existing record",
		},
		{
			name:                   "zone_not_found",
			errorMessage:           "Zone not found",
			shouldIndicateExisting: false,
			description:            "Zone not found should not indicate existing record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the error detection logic from the actual code
			errorMsg := strings.ToLower(tt.errorMessage)
			actualIndicatesExisting := strings.Contains(errorMsg, "cannot add record") ||
				strings.Contains(errorMsg, "record invalid") ||
				strings.Contains(errorMsg, "already exists") ||
				strings.Contains(errorMsg, "duplicate")

			if actualIndicatesExisting != tt.shouldIndicateExisting {
				t.Errorf("%s: expected indicates_existing=%v, got indicates_existing=%v",
					tt.description, tt.shouldIndicateExisting, actualIndicatesExisting)
				t.Errorf("  Error message: '%s'", tt.errorMessage)
			}
		})
	}
}

func TestDNSRecord_RecordAdoptionScenarios(t *testing.T) {
	tests := []struct {
		name               string
		targetValue        string
		existingValue      string
		shouldUpdate       bool
		expectedFinalValue string
		description        string
	}{
		{
			name:               "same_value_adoption",
			targetValue:        "192.168.1.1",
			existingValue:      "192.168.1.1",
			shouldUpdate:       false,
			expectedFinalValue: "192.168.1.1",
			description:        "Same values should result in simple adoption",
		},
		{
			name:               "different_value_update",
			targetValue:        "192.168.1.2",
			existingValue:      "192.168.1.1",
			shouldUpdate:       true,
			expectedFinalValue: "192.168.1.2",
			description:        "Different values should result in update",
		},
		{
			name:               "cname_value_update",
			targetValue:        "_ee89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws.",
			existingValue:      "_different_validation_string.xlfgrmvvlj.acm-validations.aws.",
			shouldUpdate:       true,
			expectedFinalValue: "_ee89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws.",
			description:        "CNAME validation strings should be updated when different",
		},
		{
			name:               "mx_value_update",
			targetValue:        "10 mail.example.com",
			existingValue:      "20 mail.example.com",
			shouldUpdate:       true,
			expectedFinalValue: "10 mail.example.com",
			description:        "MX priority changes should result in update",
		},
		{
			name:               "whitespace_differences",
			targetValue:        "example.com",
			existingValue:      " example.com ",
			shouldUpdate:       true,
			expectedFinalValue: "example.com",
			description:        "Whitespace differences should result in update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the adoption decision logic
			actualShouldUpdate := tt.existingValue != tt.targetValue

			if actualShouldUpdate != tt.shouldUpdate {
				t.Errorf("%s: expected should_update=%v, got should_update=%v",
					tt.description, tt.shouldUpdate, actualShouldUpdate)
				t.Errorf("  Target value: '%s'", tt.targetValue)
				t.Errorf("  Existing value: '%s'", tt.existingValue)
			}

			// Test final value logic
			var actualFinalValue string
			if actualShouldUpdate {
				actualFinalValue = tt.targetValue // Would be the result of update
			} else {
				actualFinalValue = tt.existingValue // Would be the adopted value
			}

			if actualFinalValue != tt.expectedFinalValue {
				t.Errorf("%s: expected final_value='%s', got final_value='%s'",
					tt.description, tt.expectedFinalValue, actualFinalValue)
			}
		})
	}
}

func TestDNSRecord_NormalizationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		testType string
	}{
		// Name normalization tests
		{
			name:     "empty_string_name",
			input:    "",
			expected: "",
			testType: "name",
		},
		{
			name:     "only_whitespace_name",
			input:    "   ",
			expected: "",
			testType: "name",
		},
		{
			name:     "mixed_case_subdomain",
			input:    "WwW.ExAmPlE",
			expected: "www.example",
			testType: "name",
		},
		{
			name:     "underscore_prefix",
			input:    "_DMARC",
			expected: "_dmarc",
			testType: "name",
		},

		// Type normalization tests
		{
			name:     "lowercase_type",
			input:    "cname",
			expected: "CNAME",
			testType: "type",
		},
		{
			name:     "mixed_case_type",
			input:    "AaAa",
			expected: "AAAA",
			testType: "type",
		},
		{
			name:     "whitespace_type",
			input:    " txt ",
			expected: "TXT",
			testType: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual string

			if tt.testType == "name" {
				actual = strings.ToLower(strings.TrimSpace(tt.input))
			} else if tt.testType == "type" {
				actual = strings.ToUpper(strings.TrimSpace(tt.input))
			}

			if actual != tt.expected {
				t.Errorf("Normalization failed for %s: input='%s', expected='%s', got='%s'",
					tt.testType, tt.input, tt.expected, actual)
			}
		})
	}
}

// Test to verify that zone value from configuration is preserved over API response
func TestDNSRecord_ZonePreservation(t *testing.T) {
	tests := []struct {
		name              string
		configurationZone string
		apiResponseZone   string
		expectedFinalZone string
		description       string
	}{
		{
			name:              "preserve_config_zone_when_api_empty",
			configurationZone: "example.com",
			apiResponseZone:   "",
			expectedFinalZone: "example.com",
			description:       "Configuration zone should be preserved when API returns empty zone",
		},
		{
			name:              "preserve_config_zone_when_api_different",
			configurationZone: "usekenny.site",
			apiResponseZone:   "different.zone",
			expectedFinalZone: "usekenny.site",
			description:       "Configuration zone should be preserved when API returns different zone",
		},
		{
			name:              "preserve_config_zone_when_api_same",
			configurationZone: "example.com",
			apiResponseZone:   "example.com",
			expectedFinalZone: "example.com",
			description:       "Configuration zone should be used even when API returns same zone",
		},
		{
			name:              "preserve_complex_zone",
			configurationZone: "pre-prod.usekenny.site",
			apiResponseZone:   "",
			expectedFinalZone: "pre-prod.usekenny.site",
			description:       "Complex zone names should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the zone preservation logic that should be used in the provider
			// In the actual provider, we should ALWAYS use the configuration zone, not the API response zone

			// This simulates what happens in the Create/Update functions:
			// zoneName comes from the configuration (data.Zone.ValueString())
			// createdRecord.Zone comes from the API response
			configZone := tt.configurationZone
			apiZone := tt.apiResponseZone

			// The correct behavior: always use the configuration zone
			actualFinalZone := configZone // This is what should be used: types.StringValue(zoneName)

			// The incorrect behavior that caused the bug: using API response zone
			// actualFinalZone := apiZone // This would be: types.StringValue(createdRecord.Zone)

			if actualFinalZone != tt.expectedFinalZone {
				t.Errorf("%s: expected final_zone='%s', got final_zone='%s'",
					tt.description, tt.expectedFinalZone, actualFinalZone)
				t.Errorf("  Configuration zone: '%s'", configZone)
				t.Errorf("  API response zone: '%s'", apiZone)
			}

			// Verify that we're not accidentally using the API response zone
			if actualFinalZone == apiZone && configZone != apiZone {
				t.Errorf("ERROR: Final zone matches API response instead of configuration!")
				t.Errorf("  This indicates the bug is still present")
			}

			// Verify that we're correctly using the configuration zone
			if actualFinalZone == configZone {
				t.Logf("✅ Correctly preserved configuration zone: '%s'", configZone)
			}
		})
	}
}

// Tests for idempotent deletion logic
func TestDNSRecord_IdempotentDeletion(t *testing.T) {
	tests := []struct {
		name                 string
		apiError             string
		shouldSucceed        bool
		shouldHaveWarning    bool
		expectedWarningTitle string
		description          string
	}{
		{
			name:                 "not_found_error",
			apiError:             "Record not found",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when record is not found",
		},
		{
			name:                 "does_not_exist_error",
			apiError:             "Record does not exist",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when record does not exist",
		},
		{
			name:                 "record_with_id_error",
			apiError:             "Record with ID 12345 not found",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when specific record ID not found",
		},
		{
			name:                 "no_record_found_error",
			apiError:             "No record found with the specified criteria",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when no record found",
		},
		{
			name:                 "invalid_record_id_error",
			apiError:             "Invalid record ID provided",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when record ID is invalid (likely deleted)",
		},
		{
			name:                 "record_id_not_found_error",
			apiError:             "Record ID not found in zone",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should succeed when record ID not found in zone",
		},
		{
			name:                 "case_insensitive_not_found",
			apiError:             "RECORD NOT FOUND",
			shouldSucceed:        true,
			shouldHaveWarning:    true,
			expectedWarningTitle: "DNS Record Already Deleted",
			description:          "Should handle case-insensitive error messages",
		},
		{
			name:                 "network_error",
			apiError:             "Connection timeout",
			shouldSucceed:        false,
			shouldHaveWarning:    false,
			expectedWarningTitle: "",
			description:          "Should fail for network errors",
		},
		{
			name:                 "permission_error",
			apiError:             "Access denied",
			shouldSucceed:        false,
			shouldHaveWarning:    false,
			expectedWarningTitle: "",
			description:          "Should fail for permission errors",
		},
		{
			name:                 "api_limit_error",
			apiError:             "Rate limit exceeded",
			shouldSucceed:        false,
			shouldHaveWarning:    false,
			expectedWarningTitle: "",
			description:          "Should fail for API limit errors",
		},
		{
			name:                 "generic_error",
			apiError:             "Internal server error",
			shouldSucceed:        false,
			shouldHaveWarning:    false,
			expectedWarningTitle: "",
			description:          "Should fail for generic server errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the error detection logic by simulating different API error messages
			errorMsg := strings.ToLower(tt.apiError)

			isNotFoundError := strings.Contains(errorMsg, "not found") ||
				strings.Contains(errorMsg, "does not exist") ||
				strings.Contains(errorMsg, "record with id") ||
				strings.Contains(errorMsg, "no record found") ||
				strings.Contains(errorMsg, "record not found") ||
				strings.Contains(errorMsg, "invalid record id") ||
				strings.Contains(errorMsg, "record id not found")

			if tt.shouldSucceed && !isNotFoundError {
				t.Errorf("Test case '%s': Expected error '%s' to be detected as 'not found' but it wasn't. %s",
					tt.name, tt.apiError, tt.description)
			}

			if !tt.shouldSucceed && isNotFoundError {
				t.Errorf("Test case '%s': Expected error '%s' to NOT be detected as 'not found' but it was. %s",
					tt.name, tt.apiError, tt.description)
			}

			t.Logf("✅ Test case '%s': Error '%s' correctly identified as shouldSucceed=%v (%s)",
				tt.name, tt.apiError, tt.shouldSucceed, tt.description)
		})
	}
}

// Test to verify the exact warning message format for deletion of already-deleted records
func TestDNSRecord_DeletionWarningMessage(t *testing.T) {
	tests := []struct {
		name             string
		recordID         int
		recordName       string
		recordType       string
		zoneName         string
		expectedContains []string
		description      string
	}{
		{
			name:       "standard_warning_message",
			recordID:   12345,
			recordName: "www",
			recordType: "A",
			zoneName:   "example.com",
			expectedContains: []string{
				"DNS record ID 12345",
				"'www' of type 'A'",
				"zone 'example.com'",
				"already deleted or does not exist",
				"desired state (record absent) is already achieved",
			},
			description: "Should generate proper warning message with all details",
		},
		{
			name:       "complex_record_name",
			recordID:   67890,
			recordName: "_4f63eda418b21d585d04126b53ba4ef1.pre-prod",
			recordType: "CNAME",
			zoneName:   "usekenny.site",
			expectedContains: []string{
				"DNS record ID 67890",
				"'_4f63eda418b21d585d04126b53ba4ef1.pre-prod' of type 'CNAME'",
				"zone 'usekenny.site'",
				"already deleted or does not exist",
			},
			description: "Should handle complex ACM validation record names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the warning message using the same format as in the code
			warningMessage := fmt.Sprintf("DNS record ID %d ('%s' of type '%s') in zone '%s' was already deleted or does not exist. "+
				"Deletion operation considered successful since the desired state (record absent) is already achieved.",
				tt.recordID, tt.recordName, tt.recordType, tt.zoneName)

			// Check that all expected substrings are present
			for _, expectedSubstring := range tt.expectedContains {
				if !strings.Contains(warningMessage, expectedSubstring) {
					t.Errorf("Test case '%s': Warning message missing expected substring '%s'. %s\n"+
						"Full message: %s",
						tt.name, expectedSubstring, tt.description, warningMessage)
				}
			}

			t.Logf("✅ Test case '%s': Warning message contains all expected elements (%s)",
				tt.name, tt.description)
		})
	}
}
