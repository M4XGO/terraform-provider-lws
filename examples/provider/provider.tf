terraform {
  required_providers {
    lws = {
      source  = "maximenony/lws"
      version = "~> 2.0"
    }
  }
}

# Configure the LWS Provider
provider "lws" {
  api_key    = var.lws_api_key
  api_secret = var.lws_api_secret
  
  # Optional: Custom API endpoint
  # endpoint = "https://api.lws.fr"
} 