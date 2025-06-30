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
  # test_mode = true  # Enable for testing
}

resource "lws_dns_record" "example" {
  name  = "www"
  type  = "A"
  value = "192.168.1.1"
  zone  = "example.com"
  ttl   = 3600
} 