terraform {
  required_providers {
    lws = {
      source  = "M4XGO/lws"
      version = "~> 0.1.0"
    }
  }
}

# Configure the LWS Provider
provider "lws" {
  login   = var.lws_login
  api_key = var.lws_api_key

  # Optional: Custom API endpoint
  # base_url = "https://api.lws.net/v1"
}
