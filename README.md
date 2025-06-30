# Terraform Provider LWS

Ce provider Terraform permet de gérer les ressources DNS chez LWS (un fournisseur français).

## Prérequis

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Utilisation

### Installation

```hcl
terraform {
  required_providers {
    lws = {
      source = "maximenony/lws"
    }
  }
}
```

### Configuration du provider

```hcl
provider "lws" {
  login   = "your-lws-login"    # Peut être défini via LWS_LOGIN
  api_key = "your-lws-api-key"  # Peut être défini via LWS_API_KEY
  base_url = "https://api.lws.net/v1"  # Optionnel, valeur par défaut
  test_mode = false  # Optionnel, pour les tests
}
```

### Variables d'environnement

- `LWS_LOGIN` : Votre identifiant LWS
- `LWS_API_KEY` : Votre clé API LWS
- `LWS_BASE_URL` : URL de base de l'API (optionnel)
- `LWS_TEST_MODE` : Mode test (optionnel)

## Ressources

### `lws_dns_record`

Gère les enregistrements DNS.

```hcl
resource "lws_dns_record" "example" {
  name  = "www"
  type  = "A"
  value = "192.168.1.1"
  zone  = "example.com"
  ttl   = 3600
}
```

#### Arguments

- `name` (Required) : Nom de l'enregistrement DNS
- `type` (Required) : Type d'enregistrement (A, AAAA, CNAME, MX, TXT, etc.)
- `value` (Required) : Valeur de l'enregistrement
- `zone` (Required) : Zone DNS
- `ttl` (Optional) : TTL en secondes

#### Attributes

- `id` : Identifiant unique de l'enregistrement

## Data Sources

### `lws_dns_zone`

Récupère les informations d'une zone DNS.

```hcl
data "lws_dns_zone" "example" {
  name = "example.com"
}
```

#### Arguments

- `name` (Required) : Nom de la zone DNS

#### Attributes

- `records` : Liste des enregistrements DNS dans la zone

## Développement

### Construction

```bash
go build
```

### Tests

```bash
go test ./...
```

### Installation en local

```bash
make install
```

## License

Mozilla Public License 2.0 