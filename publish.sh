#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

if [ -z "$1" ]; then
  # Try to find the highest existing tag
  LATEST_TAG=$(git tag --sort=-v:refname | head -n 1 2>/dev/null || true)
  
  if [ -z "$LATEST_TAG" ]; then
    TAG="v1.0.0"
    echo "🤖 No previous tags found. Auto-starting at $TAG"
  else
    # Extract the version numbers (remove the 'v' prefix)
    VERSION=${LATEST_TAG#v}
    IFS='.' read -r -a parts <<< "$VERSION"
    
    # Check if it's a standard semver tag (x.y.z)
    if [ ${#parts[@]} -eq 3 ]; then
      patch=$((parts[2] + 1))
      TAG="v${parts[0]}.${parts[1]}.$patch"
      echo "🤖 Auto-incrementing tag from $LATEST_TAG to $TAG"
    else
      echo "❌ Error: Could not auto-increment tag '$LATEST_TAG'. Please provide a version manually."
      echo "👉 Usage: ./publish.sh v1.0.1"
      exit 1
    fi
  fi
else
  TAG=$1
  # Automatically add the 'v' prefix if you forgot it
  if [[ $TAG != v* ]]; then
    TAG="v$TAG"
  fi
fi

echo "🚀 Preparing to publish release $TAG..."

# 1. Commit and push any uncommitted changes first
./auto_push.sh "chore: prep release $TAG"

# 2. Create the git tag locally
echo "🏷️  Tagging commit as $TAG..."
git tag $TAG

# 3. Push the specific tag to GitHub to trigger GoReleaser and NPM
echo "⬆️  Pushing tag $TAG to origin..."
git push origin $TAG

echo "✅ Done! GitHub Actions is now building and publishing your release."
echo "You can watch the progress on your GitHub repository under the 'Actions' tab."
