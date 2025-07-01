#!/bin/bash

echo "🔍 Vérification des Releases GitHub et Terraform Registry"
echo "=================================================="

# Vérifier les tags Git
echo ""
echo "📋 Tags Git locaux :"
git tag --sort=-version:refname | head -5

echo ""
echo "📋 Tags Git distants :"
git ls-remote --tags origin | grep -E 'refs/tags/v[0-9]' | sort -V -k2 | tail -5

# Vérifier les releases GitHub (si gh CLI est installé)
echo ""
if command -v gh &> /dev/null; then
    echo "🚀 Releases GitHub :"
    gh release list --limit 5 2>/dev/null || echo "❌ Erreur : Authentification GitHub requise (gh auth login)"
else
    echo "⚠️  GitHub CLI (gh) non installé"
    echo "   Pour installer : brew install gh"
fi

# Vérifier si les artefacts existent pour une release
echo ""
echo "📦 Vérification des artefacts (dernière release) :"
LATEST_TAG=$(git tag --sort=-version:refname | head -1)
if [ -n "$LATEST_TAG" ]; then
    echo "   Tag analysé : $LATEST_TAG"
    if command -v gh &> /dev/null; then
        gh release view "$LATEST_TAG" --json assets -q '.assets[].name' 2>/dev/null || echo "   ❌ Pas d'artefacts trouvés"
    else
        echo "   ⚠️  GitHub CLI requis pour vérifier les artefacts"
    fi
else
    echo "   ❌ Aucun tag trouvé"
fi

# Instructions pour Terraform Registry
echo ""
echo "📝 Instructions Terraform Registry :"
echo "   1. Allez sur https://registry.terraform.io"
echo "   2. Connectez-vous avec votre compte GitHub"  
echo "   3. Ajoutez le repository : M4XGO/terraform-provider-lws"
echo "   4. Le registry détectera automatiquement les releases signées GPG"

# Test d'installation locale
echo ""
echo "🧪 Test d'installation locale :"
echo "   mkdir -p test-install && cd test-install"
echo "   cat > main.tf << 'EOF'"
echo "terraform {"
echo "  required_providers {"
echo "    lws = {"
echo "      source  = \"M4XGO/lws\""
echo "      version = \">= 1.0.0\""
echo "    }"
echo "  }"
echo "}"
echo "EOF"
echo "   terraform init"

echo ""
echo "✅ Vérification terminée" 