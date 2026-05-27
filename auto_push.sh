#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Default commit message if none is provided as an argument
COMMIT_MSG=${1:-"Auto-commit: Update project files"}

echo "📦 Adding changes..."
git add .

# Check if there are changes to commit before trying to commit
if git diff-index --quiet HEAD --; then
    echo "✨ No changes to commit."
else
    echo "📝 Committing with message: '$COMMIT_MSG'"
    git commit -m "$COMMIT_MSG"
fi

echo "🚀 Pushing to remote repository..."
git push

echo "✅ Done!"
