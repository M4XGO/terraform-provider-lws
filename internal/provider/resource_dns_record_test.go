package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDNSRecordResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSRecordResourceConfig("terraform-test", "A", "192.0.2.1", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.0.2.1"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "ttl", "3600"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "lws_dns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDNSRecordResourceConfig("terraform-test", "A", "192.0.2.2", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.0.2.2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDNSRecordResource_CNAME(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for CNAME
			{
				Config: testAccDNSRecordResourceConfig("terraform-cname", "CNAME", "example.com.", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-cname"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "example.com."),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_TXT(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for TXT
			{
				Config: testAccDNSRecordResourceConfig("terraform-txt", "TXT", "v=spf1 include:_spf.google.com ~all", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-txt"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "TXT"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "v=spf1 include:_spf.google.com ~all"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_AAAA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for AAAA
			{
				Config: testAccDNSRecordResourceConfig("terraform-ipv6", "AAAA", "2001:db8::1", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-ipv6"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "AAAA"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "2001:db8::1"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_MX(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for MX
			{
				Config: testAccDNSRecordResourceConfig("@", "MX", "10 mail.example.com.", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "@"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "MX"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "10 mail.example.com."),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_TTL_Values(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with minimum TTL
			{
				Config: testAccDNSRecordResourceConfigWithTTL("terraform-ttl-test", "A", "192.0.2.10", "example.com", 900),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "ttl", "900"),
				),
			},
			// Update to maximum TTL
			{
				Config: testAccDNSRecordResourceConfigWithTTL("terraform-ttl-test", "A", "192.0.2.10", "example.com", 86400),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "ttl", "86400"),
				),
			},
		},
	})
}

func testAccDNSRecordResourceConfig(name, recordType, value, zone string) string {
	return fmt.Sprintf(`
resource "lws_dns_record" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
  zone  = %[4]q
  ttl   = 3600
}
`, name, recordType, value, zone)
}

func testAccDNSRecordResourceConfigWithTTL(name, recordType, value, zone string, ttl int) string {
	return fmt.Sprintf(`
resource "lws_dns_record" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
  zone  = %[4]q
  ttl   = %[5]d
}
`, name, recordType, value, zone, ttl)
}

// Tests for existing record adoption scenarios
func TestAccDNSRecordResource_ExistingRecordAdoption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping existing record adoption test in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create initial record
			{
				Config: testAccDNSRecordResourceConfig("terraform-adoption-test", "A", "192.0.2.100", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-adoption-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.0.2.100"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
			// Step 2: Destroy the resource but leave record in DNS
			{
				Config: `# No resources defined, simulating external record creation`,
			},
			// Step 3: Recreate with same name/type but different value - should adopt and update
			{
				Config: testAccDNSRecordResourceConfig("terraform-adoption-test", "A", "192.0.2.101", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-adoption-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.0.2.101"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_CaseInsensitiveAdoption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping case insensitive adoption test in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create record with lowercase name and type
			{
				Config: testAccDNSRecordResourceConfig("terraform-case-test", "cname", "target.example.com.", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-case-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "cname"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "target.example.com."),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
			// Step 2: Update with mixed case - should be detected as same record
			{
				Config: testAccDNSRecordResourceConfigCaseMixed("TERRAFORM-CASE-TEST", "CNAME", "target.example.com.", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "TERRAFORM-CASE-TEST"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "target.example.com."),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_ACMValidationAdoption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ACM validation adoption test in short mode")
	}

	// This test simulates the exact scenario from the user's logs
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create an ACM validation record
			{
				Config: testAccDNSRecordResourceConfig(
					"_4f63eda418b21d585d04126b53ba4ef1.terraform-test",
					"CNAME",
					"_ee89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws.",
					"example.com",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "_4f63eda418b21d585d04126b53ba4ef1.terraform-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "_ee89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws."),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
			// Step 2: Update the validation value (simulates ACM renewal)
			{
				Config: testAccDNSRecordResourceConfig(
					"_4f63eda418b21d585d04126b53ba4ef1.terraform-test",
					"CNAME",
					"_new89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws.",
					"example.com",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "_new89810c7b27b5fb90b829b35ea3841a.xlfgrmvvlj.acm-validations.aws."),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_ImportWithZone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create record first
			{
				Config: testAccDNSRecordResourceConfig("terraform-import-test", "A", "192.0.2.200", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "terraform-import-test"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.0.2.200"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
					resource.TestCheckResourceAttrSet("lws_dns_record.test", "id"),
				),
			},
			// Test new import format with zone:record_id
			{
				ResourceName:      "lws_dns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["lws_dns_record.test"]
					if !ok {
						return "", fmt.Errorf("Not found: lws_dns_record.test")
					}
					return fmt.Sprintf("example.com:%s", rs.Primary.ID), nil
				},
			},
		},
	})
}

func testAccDNSRecordResourceConfigCaseMixed(name, recordType, value, zone string) string {
	return fmt.Sprintf(`
resource "lws_dns_record" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
  zone  = %[4]q
  ttl   = 3600
}
`, name, recordType, value, zone)
}
