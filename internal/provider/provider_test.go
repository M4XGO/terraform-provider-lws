package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestLWSProvider(t *testing.T) {
	t.Parallel()

	p := New("dev")()

	// Verify the provider satisfies the framework interfaces
	var _ provider.Provider = p
}

func TestLWSProvider_Metadata(t *testing.T) {
	t.Parallel()

	p := New("1.0.0")()
	resp := &provider.MetadataResponse{}
	req := provider.MetadataRequest{}

	p.Metadata(context.Background(), req, resp)

	expected := ProviderTypeName
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName to be '%s', got %s", expected, resp.TypeName)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got %s", resp.Version)
	}
}
