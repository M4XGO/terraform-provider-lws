# CI/CD du Provider Terraform LWS

Ce document décrit le système de CI/CD complet du provider Terraform LWS, conçu pour automatiser la publication sur le Terraform Registry public.

## 🚀 Vue d'ensemble

Le système CI/CD comprend plusieurs workflows GitHub Actions qui automatisent :
- Les tests et la validation du code
- La génération automatique de releases basée sur les Conventional Commits
- La publication sur le Terraform Registry
- La validation post-release

## 📋 Workflows Disponibles

### 1. CI (`.github/workflows/ci.yml`)

**Déclenchement :** Push sur `main`/`develop` et Pull Requests vers `main`

**Jobs :**
- **Tests** : Tests unitaires sur Go 1.21 et 1.22
- **Lint** : Analyse du code avec golangci-lint
- **Security** : Scan de sécurité avec gosec
- **Validation GoReleaser** : Vérification de la configuration de release
- **Validation Terraform** : Tests avec Terraform 1.5, 1.6, 1.7
- **Tests d'acceptance** : Tests avec de vraies credentials LWS (PR uniquement)
- **Build matrix** : Compilation pour toutes les plateformes cibles

### 2. Semantic Release (`.github/workflows/semantic-release.yml`)

**Déclenchement :** Push sur `main`

**Fonctionnalités :**
- Analyse des commits selon [Conventional Commits](https://www.conventionalcommits.org/)
- Génération automatique des numéros de version (semver)
- Création automatique de tags Git
- Génération du `CHANGELOG.md`
- Déclenchement automatique du workflow de release

**Types de commits supportés :**
```bash
feat: nouvelle fonctionnalité     → version mineure (0.1.0 → 0.2.0)
fix: correction de bug            → version patch (0.1.0 → 0.1.1)
perf: amélioration performance    → version patch
BREAKING CHANGE                   → version majeure (0.1.0 → 1.0.0)
docs, style, chore               → pas de release
```

### 3. Release (`.github/workflows/release.yml`)

**Déclenchement :** Push de tag `v*` (automatique via semantic-release)

**Fonctionnalités :**
- Compilation multi-plateforme avec GoReleaser
- Génération des checksums SHA256
- Signature GPG des releases
- Publication sur GitHub Releases
- Compatible avec le Terraform Registry

**Plateformes supportées :**
- `linux_amd64`, `linux_arm64`
- `darwin_amd64`, `darwin_arm64`
- `windows_amd64`

### 4. Post-Release Validation (`.github/workflows/post-release-validation.yml`)

**Déclenchement :** Publication d'une release GitHub

**Validations :**
- Téléchargement et vérification des assets
- Tests d'intégration Terraform sur toutes les plateformes
- Validation des checksums
- Tests de compatibilité avec le Terraform Registry

## 🔧 Configuration Requise

### Secrets GitHub

Configurez ces secrets dans votre repository GitHub :

```bash
# Pour la signature GPG des releases
GPG_PRIVATE_KEY=<votre-clé-privée-gpg-base64>
GPG_PASSPHRASE=<phrase-de-passe-gpg>

# Pour les tests d'acceptance (optionnel)
LWS_TEST_LOGIN=<login-lws-test>
LWS_TEST_API_KEY=<clé-api-lws-test>
```

### Variables d'environnement

Variables automatiquement disponibles :
- `GITHUB_TOKEN` : Token automatique pour les actions GitHub
- `GO_VERSION` : Version de Go (1.22 par défaut)
- `TERRAFORM_VERSION` : Version de Terraform (1.6.0 par défaut)

## 📝 Utilisation

### 1. Développement Local

```bash
# Installation des outils de développement
make tools

# Workflow de développement complet
make dev

# Tests complets comme en CI
make ci

# Validation avant commit
make pre-commit
```

### 2. Création d'une Release

#### Méthode Automatique (Recommandée)

Utilisez les Conventional Commits pour déclencher automatiquement les releases :

```bash
# Nouvelle fonctionnalité (version mineure)
git commit -m "feat: add support for AAAA DNS records"

# Correction de bug (version patch)
git commit -m "fix: handle empty DNS zone responses correctly"

# Breaking change (version majeure)
git commit -m "feat!: redesign provider configuration schema

BREAKING CHANGE: The provider configuration has been completely redesigned.
See migration guide for details."

# Push vers main pour déclencher la release
git push origin main
```

#### Méthode Manuelle

Pour forcer une release avec une version spécifique :

```bash
# Créer et pousser un tag manuellement
git tag v1.0.0
git push origin v1.0.0
```

### 3. Validation Post-Release

Le workflow de validation s'exécute automatiquement après chaque release et vérifie :
- La disponibilité des binaires sur toutes les plateformes
- L'intégrité des checksums
- La compatibilité avec différentes versions de Terraform
- L'initialisation réussie avec `terraform init`

## 🏷️ Badges pour le README

Ajoutez ces badges à votre README principal :

```markdown
[![CI](https://github.com/maximenony/terraform-provider-lws/actions/workflows/ci.yml/badge.svg)](https://github.com/maximenony/terraform-provider-lws/actions/workflows/ci.yml)
[![Release](https://github.com/maximenony/terraform-provider-lws/actions/workflows/release.yml/badge.svg)](https://github.com/maximenony/terraform-provider-lws/actions/workflows/release.yml)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4.svg)](https://registry.terraform.io/providers/maximenony/lws)
[![Go Report Card](https://goreportcard.com/badge/github.com/maximenony/terraform-provider-lws)](https://goreportcard.com/report/github.com/maximenony/terraform-provider-lws)
[![codecov](https://codecov.io/gh/maximenony/terraform-provider-lws/branch/main/graph/badge.svg)](https://codecov.io/gh/maximenony/terraform-provider-lws)
```

## 📊 Métriques et Monitoring

### Couverture de Code

- Rapports de couverture automatiques via Codecov
- Seuil minimum de 30% (ajustable)
- Rapports HTML générés pour analyse locale

### Qualité du Code

- **golangci-lint** : 20+ linters activés
- **gosec** : Analyse de sécurité
- **Go Report Card** : Évaluation publique de la qualité

### Tests

- Tests unitaires : 37.6% de couverture
- Tests d'intégration : Workflow complet avec mocks
- Tests d'acceptance : Avec vraies credentials (optionnel)

## 🔍 Dépannage

### Échec de Release

Si une release échoue :

1. Vérifiez les logs du workflow `release.yml`
2. Validez la configuration GoReleaser : `make release-check`
3. Testez localement : `make release-test`

### Problèmes de Signature GPG

```bash
# Générer une nouvelle clé GPG
gpg --full-generate-key

# Exporter en base64 pour les secrets GitHub
gpg --armor --export-secret-keys YOUR_KEY_ID | base64
```

### Tests d'Acceptance qui Échouent

Les tests d'acceptance nécessitent de vraies credentials LWS :

```bash
export LWS_LOGIN="your-lws-id"
export LWS_API_KEY="your-api-key"
export LWS_TEST_MODE="true"
make testacc
```

## 🚀 Améliorations Futures

- [ ] Intégration avec Dependabot pour les mises à jour automatiques
- [ ] Tests de performance automatisés
- [ ] Déploiement automatique de la documentation
- [ ] Integration avec Slack/Discord pour les notifications
- [ ] Tests end-to-end avec de vraies resources LWS

## 📚 Ressources

- [Terraform Registry Publishing](https://www.terraform.io/registry/providers/publishing)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Release](https://semantic-release.gitbook.io/)
- [GitHub Actions](https://docs.github.com/en/actions) 