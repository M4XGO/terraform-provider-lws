# Example DNS A record
resource "lws_dns_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  content = "192.168.1.1"
  ttl     = 3600
}

# Example DNS CNAME record
resource "lws_dns_record" "blog" {
  zone    = "example.com"
  name    = "blog"
  type    = "CNAME"
  content = "www.example.com"
  ttl     = 3600
}

# Example DNS MX record
resource "lws_dns_record" "mail" {
  zone     = "example.com"
  name     = "@"
  type     = "MX"
  content  = "10 mail.example.com"
  ttl      = 3600
}

# Example DNS TXT record for domain verification
resource "lws_dns_record" "verification" {
  zone    = "example.com"
  name    = "@"
  type    = "TXT"
  content = "v=spf1 include:_spf.google.com ~all"
  ttl     = 300
} 