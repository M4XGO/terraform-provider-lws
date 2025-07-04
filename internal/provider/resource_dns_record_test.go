package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
