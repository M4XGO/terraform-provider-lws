# Terraform Provider LWS

Un provider Terraform pour gérer les enregistrements DNS chez LWS (hébergeur français).

## 🚀 Installation Rapide

```hcl
terraform {
  required_providers {
    lws = {
      source  = "maximenony/lws"
      version = "~> 2.0"
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

- ✅ Gestion complète des enregistrements DNS
- ✅ Support des types A, AAAA, CNAME, MX, TXT
- ✅ Validation automatique des données
- ✅ Tests unitaires et d'intégration complets
- ✅ Documentation français/anglais
- ✅ CI/CD automatisé avec releases signées GPG

## 📝 Exemple d'utilisation

```hcl
# Créer un enregistrement A
resource "lws_dns_record" "www" {
  zone_name = "mondomaine.fr"
  name      = "www"
  type      = "A"
  value     = "192.168.1.100"
  ttl       = 3600
}

# Créer un enregistrement CNAME
resource "lws_dns_record" "blog" {
  zone_name = "mondomaine.fr"
  name      = "blog"
  type      = "CNAME"
  value     = "www.mondomaine.fr"
  ttl       = 3600
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

## 🤝 Contribution

1. Fork le projet
2. Créer une branche (`git checkout -b feature/AmazingFeature`)
3. Commit (`git commit -m 'Add some AmazingFeature'`)
4. Push (`git push origin feature/AmazingFeature`)
5. Créer une Pull Request

## 📄 Licence

Distribué sous licence MIT. Voir `LICENSE` pour plus d'informations. 