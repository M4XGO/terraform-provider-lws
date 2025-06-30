terraform {
  required_providers {
    lws = {
      source = "maximenony/lws"
    }
  }
}

provider "lws" {
  login   = "your-lws-login"
  api_key = "your-lws-api-key"
}

data "lws_dns_zone" "example" {
  name = "example.com"
}

# Display all DNS records in the zone
output "dns_records" {
  value = data.lws_dns_zone.example.records
} 