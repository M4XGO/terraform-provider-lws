# Configuration du Terraform Registry

## 🔍 Pipeline de Release Unifié

Votre provider Terraform utilise maintenant un **pipeline de release unifié** qui combine semantic-release et GoReleaser en un seul workflow. Ce pipeline gère automatiquement la détection des versions et la signature PGP de tous les artefacts.

### 🚀 **Workflow Unique : `Release Pipeline`**

**Déclencheurs multiples :**
- ✅ **Push vers `main`** → Détection automatique de version avec semantic-release
- ✅ **Tags `v*`** → Release directe avec GoReleaser  
- ✅ **Workflow manuel** → Release forcée avec version personnalisée

**Pipeline complet :**
1. **Semantic Release** → Analyse des commits et création de tags GPG signés
2. **GoReleaser** → Génération des binaires et signatures PGP
3. **Validation Terraform** → Tests multi-versions et multi-OS
4. **Badges automatiques** → Mise à jour du README

### 🔐 Configuration PGP Active

**Variables d'environnement GitHub requises :**
- `GPG_PRIVATE_KEY` : Clé privée PGP (format ASCII armored)
- `GPG_PASSPHRASE` : Passphrase de la clé PGP

**Artefacts signés automatiquement :**
- ✅ `terraform-provider-lws_VERSION_SHA256SUMS` → `terraform-provider-lws_VERSION_SHA256SUMS.sig`
- ✅ `terraform-provider-lws_VERSION_linux_amd64.zip` → `terraform-provider-lws_VERSION_linux_amd64.zip.sig`
- ✅ `terraform-provider-lws_VERSION_darwin_amd64.zip` → `terraform-provider-lws_VERSION_darwin_amd64.zip.sig`
- ✅ Et tous les autres binaires multi-plateformes

### 🎯 **Déclenchement d'une Release**

#### Option 1 : Semantic Release (Automatique - Recommandée)
```bash
# Commit avec conventional commit format
git commit -m "feat: add new DNS record type support"
git push origin main

# Le pipeline détecte automatiquement le type de version (patch/minor/major)
# et crée une release complète avec artefacts signés
```

#### Option 2 : Release Manuelle
```bash
# Via GitHub Actions UI
# Actions → "Release Pipeline" → "Run workflow"
# Spécifier la version (ex: v1.0.5) ou laisser vide pour auto-détection
```

#### Option 3 : Git Tag Direct
```bash
# Créer un tag et pusher (déclenche automatiquement le pipeline)
git tag v1.0.5
git push origin v1.0.5
```

## ⚡ Configuration Rapide des Clés PGP

### Génération Automatique (Recommandée) 🤖

**En 1 clic depuis GitHub :**
1. Allez sur **Actions** dans votre repository GitHub
2. Sélectionnez **"Setup GPG for Terraform Registry"**  
3. Cliquez sur **"Run workflow"**
4. Attendez 2-3 minutes ⏳

**Le workflow va automatiquement :**
- ✅ Générer une clé PGP RSA 4096 bits sécurisée
- ✅ Configurer les secrets GitHub (`GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`)
- ✅ Déclencher une release test avec signature PGP
- ✅ Valider que toutes les signatures fonctionnent

### Configuration Manuelle 💻

Si vous avez déjà une clé PGP :

```bash
# 1. Exporter votre clé existante
gpg --armor --export-secret-key YOUR_KEY_ID > private.key

# 2. Ajouter dans GitHub Secrets :
# - GPG_PRIVATE_KEY : contenu de private.key  
# - GPG_PASSPHRASE : votre passphrase

# 3. Déclencher une release
git commit -m 'feat: enable PGP signing' --allow-empty
git push origin main
```

## ⚡ **Optimisations Performance**

### **Accélération Semantic-Release**
Le pipeline intègre plusieurs optimisations pour éviter les timeouts de plus d'1 heure :

- ✅ **Timeout strict** : 15 minutes maximum pour semantic-release
- ✅ **Installation optimisée** : Versions spécifiques des dépendances
- ✅ **Configuration optimisée** : Moins d'appels API GitHub
- ✅ **Debug réduit** : Logs moins verbeux pour accélérer
- ✅ **Plugin GitHub simplifié** : Désactivation des fonctionnalités lentes

### **Paramètres Anti-Timeout**
```yaml
semantic-release:
  timeout-minutes: 15  # Force l'arrêt si trop long
  
goreleaser:
  timeout-minutes: 30  # Temps pour build multi-plateformes

validate-terraform-registry:
  timeout-minutes: 20  # Validation rapide
```

**Note technique :** Le cache NPM n'est pas utilisé car semantic-release est installé globalement pour éviter la complexité d'un package.json dans un repository Terraform. L'installation prend ~30 secondes et reste plus simple à maintenir.

## 🔧 Fonctionnement Technique

### Pipeline Unifié

1. **Détection de Déclencheur**
   - Push vers main → `semantic-release` job
   - Tag v* → `goreleaser` job directement  
   - Workflow manuel → `goreleaser` job avec version spécifiée

2. **Semantic Release** (si déclenché par push main)
   - Analyse des commits conventional
   - Génération automatique du CHANGELOG.md
   - Création de tags GPG signés
   - Passage du relais à GoReleaser

3. **GoReleaser** (toujours exécuté si nouvelle version)
   ```yaml
   signs:
     - id: checksum_signature  # SHA256SUMS obligatoire
     - id: archive_signature   # Binaires ZIP (sécurité)
   ```

4. **Validation Multi-Plateforme**
   - Tests Terraform 1.5.0, 1.6.0, 1.7.0
   - Validation Ubuntu, macOS, Windows
   - Vérification `terraform init` et `terraform validate`

5. **Finalisation**
   - Mise à jour automatique des badges README
   - Rapport détaillé avec liens et usage

### Variables d'Environnement

Le pipeline passe automatiquement :
- `GPG_FINGERPRINT` : Empreinte de la clé (générée automatiquement)
- `GPG_TTY` : Terminal GPG pour l'interaction
- `GITHUB_TOKEN` : Token pour publier sur GitHub
- `GORELEASER_DEBUG` : Debug activé pour le dépannage

## ✅ Résultats Attendus

### Après chaque release, vous devriez voir :

**Sur GitHub Releases :**
```
terraform-provider-lws_1.0.5_linux_amd64.zip         (2.1 MB)
terraform-provider-lws_1.0.5_linux_amd64.zip.sig     (566 B)  ← Signature PGP
terraform-provider-lws_1.0.5_darwin_amd64.zip        (2.0 MB)  
terraform-provider-lws_1.0.5_darwin_amd64.zip.sig    (566 B)  ← Signature PGP
terraform-provider-lws_1.0.5_SHA256SUMS              (1.2 KB)
terraform-provider-lws_1.0.5_SHA256SUMS.sig          (566 B)  ← Signature PGP OBLIGATOIRE
```

**Validation Terraform Registry :**
- ✅ **SHASUMS file présent** 
- ✅ **SHASUMS signature présente et valide**
- ✅ **Plateformes multiples disponibles**
- ✅ **Provider accepté sur registry.terraform.io**

## 🏃 Publication sur Terraform Registry

### 1. Vérification Préalable
```bash
# Vérifier qu'une release signée existe
gh release list

# Vérifier les artefacts d'une release
gh release view v1.0.5 --json assets | jq '.assets[].name'
```

### 2. Inscription au Registry
1. Allez sur https://registry.terraform.io
2. Connectez-vous avec GitHub
3. Ajoutez le repository `maximenony/terraform-provider-lws`
4. Le registry détecte automatiquement les releases signées

### 3. Test d'Installation
```hcl
terraform {
  required_providers {
    lws = {
      source  = "maximenony/lws"
      version = "~> 1.0"
    }
  }
}

provider "lws" {
  # Configuration
}
```

```bash
terraform init  # Télécharge depuis le registry officiel
```

## 🔍 Dépannage

### Problème : Signatures manquantes
```bash
# Vérifier les secrets GitHub
gh secret list

# Re-générer les clés GPG
gh workflow run setup-gpg.yml -f force_regenerate=true
```

### Problème : GoReleaser échoue
```bash
# Tester localement
goreleaser check
goreleaser build --snapshot --clean

# Vérifier les variables d'environnement
echo $GPG_FINGERPRINT
gpg --list-secret-keys
```

### Problème : Semantic-release ne détecte pas les changements
```bash
# Vérifier le format des commits
git log --oneline -n 5

# Exemples de commits valides :
# feat: add new DNS record type
# fix: resolve API timeout issue  
# docs: update README
# BREAKING CHANGE: change API structure
```

### Problème : Signatures invalides
```bash
# Vérifier une signature manuellement
gpg --verify terraform-provider-lws_1.0.5_SHA256SUMS.sig terraform-provider-lws_1.0.5_SHA256SUMS
```

## 📊 **Avantages du Pipeline Unifié**

| Aspect | Avant (2 workflows) | Après (1 pipeline) |
|--------|---------------------|---------------------|
| **Complexité** | ❌ 2 workflows séparés | ✅ 1 pipeline unifié |
| **Maintenance** | ❌ Duplication config GPG | ✅ Configuration centralisée |
| **Déclencheurs** | ❌ Coordination manuelle | ✅ Automatique et flexible |
| **Debugging** | ❌ Logs dispersés | ✅ Vue d'ensemble complète |
| **Performance** | ❌ Double setup GPG | ✅ Import GPG unique |

## 📝 Notes Importantes

- **🔐 Sécurité** : Les clés PGP sont générées avec 4096 bits RSA
- **⏰ Expiration** : Clés valides 2 ans (renouvelables automatiquement)
- **🚀 Automatisation** : Pipeline 100% automatisé sans intervention
- **✅ Conformité** : Configuration 100% compatible Terraform Registry
- **🔄 Reproductibilité** : Signatures déterministes et vérifiables
- **🎯 Flexibilité** : 3 méthodes de déclenchement selon les besoins 