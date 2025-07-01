#!/bin/bash

echo "🔐 Configuration GPG pour Terraform Registry"
echo "============================================="

# Vérifier si GPG est installé
if ! command -v gpg &> /dev/null; then
    echo "❌ GPG n'est pas installé"
    echo "   macOS: brew install gnupg"
    echo "   Ubuntu/Debian: sudo apt install gnupg"
    exit 1
fi

# Vérifier si gh CLI est installé
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) n'est pas installé"
    echo "   macOS: brew install gh"
    echo "   Ubuntu/Debian: sudo apt install gh"
    exit 1
fi

echo ""
echo "📋 Étape 1: Génération de la clé GPG"
echo "======================================"

# Vérifier s'il y a déjà des clés
existing_keys=$(gpg --list-secret-keys --keyid-format LONG | grep sec)
if [ -n "$existing_keys" ]; then
    echo "🔍 Clés existantes trouvées :"
    gpg --list-secret-keys --keyid-format LONG
    echo ""
    read -p "Voulez-vous créer une nouvelle clé ou utiliser une existante ? (n/e): " choice
    
    if [ "$choice" = "e" ]; then
        echo "📝 Sélectionnez une clé existante :"
        gpg --list-secret-keys --keyid-format LONG | grep sec | cut -d'/' -f2 | cut -d' ' -f1
        read -p "Entrez l'ID de la clé (après sec rsa4096/): " GPG_KEY_ID
    else
        echo "🔑 Génération d'une nouvelle clé GPG..."
        echo "Répondez aux questions suivantes :"
        echo "- Type de clé : RSA and RSA (default)"
        echo "- Taille : 4096"
        echo "- Validité : 2y (2 ans)"
        echo "- Nom : Votre nom"
        echo "- Email : maxime.nony@M4XGO.com"
        echo "- Comment : LWS Terraform Provider"
        
        gpg --full-generate-key
        
        # Récupérer l'ID de la nouvelle clé
        GPG_KEY_ID=$(gpg --list-secret-keys --keyid-format LONG | grep sec | tail -1 | cut -d'/' -f2 | cut -d' ' -f1)
    fi
else
    echo "🔑 Aucune clé trouvée. Génération d'une nouvelle clé..."
    echo "Répondez aux questions suivantes :"
    echo "- Type de clé : RSA and RSA (default)"
    echo "- Taille : 4096"
    echo "- Validité : 2y (2 ans)"
    echo "- Nom : Votre nom"
    echo "- Email : maxime.nony@M4XGO.com"
    echo "- Comment : LWS Terraform Provider"
    
    gpg --full-generate-key
    
    # Récupérer l'ID de la nouvelle clé
    GPG_KEY_ID=$(gpg --list-secret-keys --keyid-format LONG | grep sec | tail -1 | cut -d'/' -f2 | cut -d' ' -f1)
fi

echo ""
echo "📤 Étape 2: Export des clés"
echo "============================"

if [ -z "$GPG_KEY_ID" ]; then
    echo "❌ Impossible de déterminer l'ID de la clé GPG"
    exit 1
fi

echo "🔑 Utilisation de la clé : $GPG_KEY_ID"

# Créer un répertoire temporaire
temp_dir=$(mktemp -d)
echo "📁 Répertoire temporaire : $temp_dir"

# Exporter la clé privée
echo "🔐 Export de la clé privée..."
gpg --armor --export-secret-key $GPG_KEY_ID > "$temp_dir/private.key"

# Exporter la clé publique
echo "🔓 Export de la clé publique..."
gpg --armor --export $GPG_KEY_ID > "$temp_dir/public.key"

echo ""
echo "⚙️  Étape 3: Configuration des secrets GitHub"
echo "=============================================="

# Demander la passphrase
echo "🔒 Entrez la passphrase de votre clé GPG :"
read -s GPG_PASSPHRASE

# Configurer les secrets GitHub
echo "📤 Configuration des secrets GitHub..."

# Vérifier l'authentification GitHub
if ! gh auth status >/dev/null 2>&1; then
    echo "❌ Non authentifié sur GitHub"
    echo "   Exécutez : gh auth login"
    exit 1
fi

# Ajouter les secrets
echo "🔐 Ajout du secret GPG_PRIVATE_KEY..."
gh secret set GPG_PRIVATE_KEY --body-file "$temp_dir/private.key" --repo M4XGO/terraform-provider-lws

echo "🔒 Ajout du secret GPG_PASSPHRASE..."
echo "$GPG_PASSPHRASE" | gh secret set GPG_PASSPHRASE --repo M4XGO/terraform-provider-lws

echo ""
echo "🧹 Nettoyage"
echo "============="
rm -rf "$temp_dir"
echo "🗑️  Fichiers temporaires supprimés"

echo ""
echo "✅ Configuration terminée !"
echo "=========================="
echo ""
echo "📝 Prochaines étapes :"
echo "1. La prochaine release créera automatiquement :"
echo "   - ✅ Fichiers SHASUMS (checksums)"
echo "   - ✅ Signature SHASUMS.sig (GPG)"
echo "   - ✅ Binaires multi-plateformes"
echo ""
echo "2. Pour tester :"
echo "   git commit -m 'fix: test release with GPG' --allow-empty"
echo "   git push origin main"
echo ""
echo "3. Vérifier sur :"
echo "   - GitHub : https://github.com/M4XGO/terraform-provider-lws/releases"
echo "   - Registry : https://registry.terraform.io (après inscription)"
echo ""
echo "🎉 Votre provider sera maintenant compatible Terraform Registry !" 