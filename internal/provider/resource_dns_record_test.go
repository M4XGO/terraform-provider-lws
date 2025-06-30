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
				Config: testAccDNSRecordResourceConfig("www", "A", "192.168.1.1", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "name", "www"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.168.1.1"),
					resource.TestCheckResourceAttr("lws_dns_record.test", "zone", "example.com"),
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
				Config: testAccDNSRecordResourceConfig("www", "A", "192.168.1.2", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("lws_dns_record.test", "value", "192.168.1.2"),
				),
			},
			// Delete testing automatically occurs in TestCase
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
