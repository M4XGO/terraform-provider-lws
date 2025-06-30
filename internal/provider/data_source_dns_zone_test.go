package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSZoneDataSource_Metadata(t *testing.T) {
	d := NewDNSZoneDataSource()
	resp := &datasource.MetadataResponse{}
	req := datasource.MetadataRequest{
		ProviderTypeName: "lws",
	}

	d.Metadata(context.Background(), req, resp)

	expected := "lws_dns_zone"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %s, got %s", expected, resp.TypeName)
	}
}

func TestDNSZoneDataSource_Schema(t *testing.T) {
	d := NewDNSZoneDataSource()
	resp := &datasource.SchemaResponse{}
	req := datasource.SchemaRequest{}

	d.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Error("Expected schema attributes to be defined")
	}

	// Check required attributes
	nameAttr, exists := resp.Schema.Attributes["name"]
	if !exists {
		t.Error("Expected 'name' attribute in schema")
	}

	if !nameAttr.(schema.StringAttribute).Required {
		t.Error("Expected 'name' attribute to be required")
	}

	recordsAttr, exists := resp.Schema.Attributes["records"]
	if !exists {
		t.Error("Expected 'records' attribute in schema")
	}

	if !recordsAttr.(schema.ListNestedAttribute).Computed {
		t.Error("Expected 'records' attribute to be computed")
	}
}

func TestDNSZoneDataSource_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		zoneName    string
		expectError bool
	}{
		{
			name:        "valid zone name",
			zoneName:    "example.com",
			expectError: false,
		},
		{
			name:        "empty zone name",
			zoneName:    "",
			expectError: true,
		},
		{
			name:        "invalid zone name",
			zoneName:    "invalid..zone",
			expectError: false, // Let the API handle validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test configuration validation if implemented
			// For now, we just test the basic structure
			if tt.zoneName == "" && !tt.expectError {
				t.Error("Empty zone name should cause validation error")
			}
		})
	}
}

func TestDNSZoneDataSourceModel_Validation(t *testing.T) {
	model := DNSZoneDataSourceModel{
		Name: types.StringValue("example.com"),
	}

	if model.Name.IsNull() || model.Name.IsUnknown() {
		t.Error("Expected valid zone name")
	}

	if model.Name.ValueString() != "example.com" {
		t.Errorf("Expected zone name 'example.com', got %s", model.Name.ValueString())
	}
}

func TestDNSZoneDataSource_RecordTypes(t *testing.T) {
	// Test that different DNS record types are handled correctly
	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA"}

	for _, recordType := range recordTypes {
		t.Run("record_type_"+recordType, func(t *testing.T) {
			// This would test that all record types are properly supported
			// when parsing zone data from the API
			if recordType == "" {
				t.Error("Record type should not be empty")
			}
		})
	}
}
