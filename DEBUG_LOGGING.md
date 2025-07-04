# Guide de Débogage - Terraform Provider LWS

## Activation du Logging Détaillé

Pour déboguer les problèmes avec le provider Terraform LWS, vous pouvez activer le logging détaillé qui montrera les requêtes API et les réponses.

### Configuration du Logging

Avant d'exécuter vos commandes Terraform, définissez la variable d'environnement :

```bash
export TF_LOG=DEBUG
```

Ou pour un niveau de détail maximum :

```bash
export TF_LOG=TRACE
```

### Exemple de Sortie avec DEBUG

Avec `TF_LOG=DEBUG`, vous verrez des logs détaillés comme :

```
[INFO]  LWS API Request: GET https://api.lws.net/v1/domain/example.com/zdns
Headers: X-Auth-Login=your-login, X-Test-Mode=false
[DEBUG] Request Body: <empty>

[DEBUG] LWS API Response: status=200, url=https://api.lws.net/v1/domain/example.com/zdns
Response Body: {"code":200,"info":"Fetched DNS Zone","data":[...]}

[ERROR] Failed to read DNS zone: zone_name=example.com error=API returned empty response (status 404) for URL: https://api.lws.net/v1/domain/example.com/zdns
```

### Erreurs Communes et Diagnostics

#### Erreur 404 (Zone Introuvable)

Si vous voyez une erreur comme :
```
[ERROR] Provider Error: Unable to read DNS zone 'example.com', got error: API returned empty response (status 404)
```

**Causes possibles :**
1. Le nom de domaine n'existe pas dans votre compte LWS
2. Le domaine n'est pas encore configuré pour les DNS
3. Les identifiants d'authentification sont incorrects

**Solutions :**
1. Vérifiez que le domaine existe dans votre panel LWS
2. Assurez-vous que les DNS LWS sont activés pour ce domaine
3. Vérifiez vos identifiants `login` et `api_key`

#### Erreur 401 (Non Autorisé)

Si vous voyez :
```
[ERROR] API error: Unauthorized access
```

**Causes possibles :**
1. Login incorrect dans la configuration
2. Clé API incorrecte ou expirée
3. Permissions insuffisantes pour l'API

**Solutions :**
1. Vérifiez votre login LWS dans la configuration du provider
2. Régénérez votre clé API depuis le panel LWS
3. Contactez le support LWS pour vérifier les permissions

#### Erreur de Connectivité

Si vous voyez des erreurs de timeout ou de connexion :
```
[ERROR] Failed to make request: Get "https://api.lws.net/v1/domain/votre-domaine.com/zdns": dial tcp: i/o timeout
```

**Solutions :**
1. Vérifiez votre connexion internet
2. Assurez-vous que l'API LWS n'est pas en maintenance
3. Vérifiez les règles de pare-feu si vous êtes derrière un proxy

### Configuration d'Exemple pour Tests

Pour tester avec un domaine spécifique :

```hcl
terraform {
  required_providers {
    lws = {
      source = "M4XGO/lws"
    }
  }
}

provider "lws" {
  login   = "votre-login-lws"
  api_key = "votre-cle-api"
}

# Test avec data source
data "lws_dns_zone" "example" {
  name = "votre-domaine.com"
}

# Test avec resource
resource "lws_dns_record" "test" {
  name  = "test"
  type  = "A"
  value = "192.0.2.1"
  zone  = "votre-domaine.com"
  ttl   = 3600
}
```

### Nettoyage des Logs

Pour désactiver le logging après diagnostic :

```bash
unset TF_LOG
``` 