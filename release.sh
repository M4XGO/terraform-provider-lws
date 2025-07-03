#!/bin/bash
set -e

# Vérifier qu'une version est fournie
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Exemple: $0 2.1.5"
  exit 1
fi

VERSION="$1"
TAG="v$VERSION"

echo "🚀 Création de la release $TAG..."

# 1. Mettre à jour go.mod et faire un commit
echo "📝 Préparation du commit de release..."
go mod tidy

# 2. Commit avec message conventional
git add .
git commit -m "chore(release): $VERSION"

# 3. Créer et pousser le tag
echo "🏷️ Création du tag $TAG..."
git tag -a "$TAG" -m "Release $TAG"

# 4. Pousser commit + tag
echo "📤 Push du commit et du tag..."
git push origin main
git push origin "$TAG"

echo "✅ Release $TAG créée avec succès !"
echo "Le workflow GoReleaser va se déclencher automatiquement."
