# Retrieve information about a DNS zone
data "lws_dns_zone" "example" {
  zone = "example.com"
}

# Use the zone data to create a DNS record
resource "lws_dns_record" "subdomain" {
  zone    = data.lws_dns_zone.example.zone
  name    = "api"
  type    = "A"
  content = "192.168.1.100"
  ttl     = 3600
}

# Output zone information
output "zone_info" {
  description = "DNS zone information"
  value = {
    zone    = data.lws_dns_zone.example.zone
    id      = data.lws_dns_zone.example.id
    records = data.lws_dns_zone.example.records
  }
} 