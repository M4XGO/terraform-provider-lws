# Terraform Provider LWS

## Usage

Here's a quick example of how to use this provider:

```hcl
terraform {
  required_providers {
    lws = {
      source  = "M4XGO/lws"
      version = "~> 0.1.0"
    }
  }
}

provider "lws" {
  # Configuration options
}

# Example DNS zone data source
data "lws_dns_zone" "example" {
  zone = "example.com"
}

# Example DNS record resource
resource "lws_dns_record" "example" {
  zone  = data.lws_dns_zone.example.zone
  name  = "www"
  type  = "A"
  value = "192.168.1.1"
  ttl   = 3600
}
```

For more examples, see the [examples/](examples/) directory.

## Documentation

Full documentation is available in the [docs/](docs/) directory:

- [Provider Configuration](docs/index.md)
- [Data Sources](docs/data-sources/)
- [Resources](docs/resources/)

Un provider Terraform pour gérer les enregistrements DNS chez LWS (hébergeur français).

## 🚀 Installation Rapide

```hcl
terraform {
  required_providers {
    lws = {
      source  = "M4XGO/lws"
      version = "~> 0.1.0"
    }
  }
}

provider "lws" {
  api_key = var.lws_api_key
}
```

## 🔧 Résolution des Erreurs Terraform Registry

Si vous voyez ces erreurs sur registry.terraform.io :
- `Missing SHASUMS file`
- `Missing SHASUMS signature file` 
- `Missing platforms`

**Solution automatique en 1 clic :**

1. Allez sur **Actions** dans GitHub
2. Sélectionnez **"Setup GPG for Terraform Registry"**
3. Cliquez sur **"Run workflow"**
4. Attendez 2-3 minutes ⏳

**C'est tout !** Le workflow va automatiquement :
- ✅ Générer les clés GPG
- ✅ Configurer les secrets GitHub
- ✅ Déclencher une nouvelle release avec tous les artefacts requis

## 📖 Documentation

- [Configuration du Provider](docs/)
- [Guide Terraform Registry](docs/TERRAFORM_REGISTRY_SETUP.md)
- [Tests et Développement](internal/provider/README_TESTS.md)

## 🎯 Fonctionnalités

- ✅ **Gestion robuste des enregistrements DNS** - Création, modification, suppression avec validation automatique
- ✅ **Détection automatique des changements d'ID** - Récupération transparente quand l'API LWS change les IDs
- ✅ **Gestion d'erreurs améliorée** - Messages d'erreur détaillés et récupération gracieuse
- ✅ **Support de debug avancé** - Logs détaillés pour le troubleshooting
- ✅ **Support d'import** - Import des enregistrements DNS existants dans Terraform
- ✅ **Support des types A, AAAA, CNAME, MX, TXT, NS, SRV, etc.**
- ✅ **Tests unitaires et d'intégration complets** - Incluant tests de drift d'ID
- ✅ **Documentation français/anglais** - Avec exemples pratiques
- ✅ **CI/CD automatisé** - Releases signées GPG

## 📝 Exemple d'utilisation

```hcl
# Créer un enregistrement A
resource "lws_dns_record" "www" {
  zone  = "mondomaine.fr"
  name  = "www"
  type  = "A"
  value = "192.168.1.100"
  ttl   = 3600
}

# Créer un enregistrement CNAME
resource "lws_dns_record" "blog" {
  zone  = "mondomaine.fr"
  name  = "blog"
  type  = "CNAME"
  value = "www.mondomaine.fr"
  ttl   = 3600
}
```

## 🧪 Tests

```bash
# Tests unitaires
make test

# Tests d'intégration (nécessite API_KEY)
make testacc

# Tests avec coverage
make test-coverage
```

## 🔧 Debug et Troubleshooting

### Logs de debug détaillés

```bash
export TF_LOG=DEBUG
terraform plan
terraform apply
```

Les logs de debug incluent :
- 🔍 Détails des requêtes/réponses API
- 🎯 Suivi des IDs et détection des drifts
- ⚠️ Erreurs de validation et avertissements
- 📊 Opérations de gestion du state

### Problèmes courants

#### Changements d'ID de records
L'API LWS peut changer les IDs lors des mises à jour. Le provider détecte automatiquement ces changements et met à jour le state - aucune intervention manuelle requise.

#### Import d'enregistrements existants
```bash
terraform import lws_dns_record.example 12345
```

### Validation des champs
Le provider valide tous les champs requis avant les appels API :
- `name`: Ne peut pas être vide ou contenir seulement des espaces
- `type`: Doit être un type DNS valide
- `value`: Ne peut pas être vide ou contenir seulement des espaces
- `zone`: Doit être un nom de domaine valide

## 🤝 Contribution

1. Fork le projet
2. Créer une branche (`git checkout -b feature/AmazingFeature`)
3. Commit (`git commit -m 'Add some AmazingFeature'`)
4. Push (`git push origin feature/AmazingFeature`)
5. Créer une Pull Request

## 📄 Licence

Distribué sous licence MIT. Voir `LICENSE` pour plus d'informations. 
