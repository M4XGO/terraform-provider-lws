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
  # Optional: API request timeout in seconds
  # timeout = 30
  # Optional: Number of retries for an API request
  # retries = 3
  # Optional: Delay between retries in seconds
  # delay = 15
  # Optional: Backoff multiplier for delay between retries
  # backoff = 2
}
