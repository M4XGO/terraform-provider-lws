# Guide de Débogage - Terraform Provider LWS

## Activer le Logging Détaillé

Pour diagnostiquer les erreurs 404 et autres problèmes API, vous pouvez activer plusieurs niveaux de logging :

### 1. Logging Terraform (Recommandé)

```bash
# Activer tous les logs de débogage
export TF_LOG=DEBUG
terraform plan

# Ou seulement les logs du provider
export TF_LOG_PROVIDER=DEBUG
terraform plan

# Logs très détaillés (includes HTTP requests/responses)
export TF_LOG=TRACE
terraform plan
```

### 2. Logs dans un fichier

```bash
# Rediriger les logs vers un fichier
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform-debug.log
terraform plan
```

### 3. Exemple de sortie avec nos améliorations

Avec le logging activé, vous verrez maintenant :

```
[INFO]  Reading DNS zone: zone_name=example.com base_url=https://api.lws.net login=votre_login
[DEBUG] LWS API Request: GET https://api.lws.net/v1/domain/example.com/zdns
[DEBUG] Headers: X-Auth-Login=votre_login, X-Auth-Pass=[REDACTED], X-Test-Mode=
[DEBUG] LWS API Response: Status 404 (Not Found)
[DEBUG] Response Headers: map[Content-Type:[application/json]]
[DEBUG] Response Body: ""
[ERROR] Failed to read DNS zone: zone_name=example.com error=API returned empty response (status 404) for URL: https://api.lws.net/v1/domain/example.com/zdns
```

### 4. Points de vérification

Avec ces logs, vérifiez :

1. **URL correcte** : L'endpoint appelé correspond-il à votre attente ?
2. **Authentification** : Les headers X-Auth-Login sont-ils corrects ?
3. **Domaine** : Le nom de domaine existe-t-il dans votre compte LWS ?
4. **Base URL** : Utilisez-vous la bonne URL de l'API LWS ?

### 5. Configuration Provider

Assurez-vous que votre configuration provider est correcte :

```hcl
provider "lws" {
  login    = "votre_login"       # ou variable d'environnement LWS_LOGIN
  api_key  = "votre_api_key"     # ou variable d'environnement LWS_API_KEY
  base_url = "https://api.lws.net" # par défaut
}

data "lws_dns_zone" "site" {
  name = "votre-domaine.com"  # Assurez-vous que ce domaine existe
}
```

### 6. Désactiver le logging

```bash
unset TF_LOG
unset TF_LOG_PROVIDER
unset TF_LOG_PATH
```

## Erreurs Communes

### Erreur 404 - "API returned empty response"

**Causes possibles :**
- Le domaine n'existe pas dans votre compte LWS
- Mauvaise URL de base (vérifiez `base_url`)
- Authentification incorrecte
- Le domaine n'est pas géré par LWS

**Solution :**
1. Vérifiez que le domaine existe dans votre interface LWS
2. Testez l'authentification avec curl :
```bash
curl -H "X-Auth-Login: votre_login" \
     -H "X-Auth-Pass: votre_api_key" \
     https://api.lws.net/v1/domain/votre-domaine.com/zdns
```

### Erreur 401 - "Unauthorized"

**Causes :**
- Login ou API key incorrects
- API key expirée

**Solution :**
- Vérifiez vos identifiants LWS
- Générez une nouvelle API key si nécessaire 