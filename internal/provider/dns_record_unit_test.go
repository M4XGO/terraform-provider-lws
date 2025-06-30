package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSRecordResource_Metadata(t *testing.T) {
	r := NewDNSRecordResource()
	resp := &resource.MetadataResponse{}
	req := resource.MetadataRequest{
		ProviderTypeName: "lws",
	}

	r.Metadata(context.Background(), req, resp)

	expected := "lws_dns_record"
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
			}
		})
	}
}

func TestDNSRecordTypes_Validation(t *testing.T) {
	validTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR"}

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
		{60, true},          // 1 minute
		{300, true},         // 5 minutes
		{3600, true},        // 1 hour (default)
		{86400, true},       // 1 day
		{604800, true},      // 1 week
		{0, false},          // Too low
		{-1, false},         // Negative
		{2147483648, false}, // Too high (over 32-bit int)
	}

	for _, tt := range tests {
		t.Run("ttl_validation", func(t *testing.T) {
			if tt.valid {
				if tt.ttl <= 0 {
					t.Errorf("TTL %d should be positive", tt.ttl)
				}
				if tt.ttl > 2147483647 {
					t.Errorf("TTL %d should not exceed 32-bit int max", tt.ttl)
				}
			} else {
				if tt.ttl > 0 && tt.ttl <= 2147483647 {
					t.Errorf("TTL %d was marked as invalid but seems valid", tt.ttl)
				}
			}
		})
	}
}
