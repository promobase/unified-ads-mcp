#!/usr/bin/env bash

# Parse command line arguments
minor=false
major=false
while [ "$#" -gt 0 ]; do
  case "$1" in
    --minor) minor=true; shift 1;;
    --major) major=true; shift 1;;
    *) echo "Unknown parameter: $1"; exit 1;;
  esac
done

git fetch --force --tags

# Get the latest Git tag
latest_tag=$(git tag --sort=-v:refname | grep -E '^v[0-9]' | head -1)

# If there is no tag, start with v0.0.0
if [ -z "$latest_tag" ]; then
    echo "No tags found, starting with v0.0.0"
    latest_tag="v0.0.0"
fi

echo "Latest tag: $latest_tag"

# Remove the 'v' prefix and split the tag into major, minor, and patch numbers
version_without_v=${latest_tag#v}
IFS='.' read -ra VERSION <<< "$version_without_v"

if [ "$major" = true ]; then
    # Increment the major version and reset minor and patch to 0
    major_number=${VERSION[0]}
    let "major_number++"
    new_version="v$major_number.0.0"
elif [ "$minor" = true ]; then
    # Increment the minor version and reset patch to 0
    minor_number=${VERSION[1]}
    let "minor_number++"
    new_version="v${VERSION[0]}.$minor_number.0"
else
    # Increment the patch version
    patch_number=${VERSION[2]}
    let "patch_number++"
    new_version="v${VERSION[0]}.${VERSION[1]}.$patch_number"
fi

echo "New version: $new_version"

# Create and push the tag
git tag -a $new_version -m "Release $new_version"
git push origin $new_version

echo ""
echo "✅ Tag $new_version created and pushed!"
echo "🚀 GitHub Actions will now build and create the release."
echo ""

# Try to get GitHub repository from git remote
GITHUB_REPO=$(git config --get remote.origin.url | sed -E 's/.*github.com[:/](.*)\.git/\1/' 2>/dev/null || echo "")
if [ -n "$GITHUB_REPO" ]; then
    echo "You can monitor the progress at:"
    echo "https://github.com/$GITHUB_REPO/actions"
fi