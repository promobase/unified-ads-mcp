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

# Update VERSION file
echo "${new_version#v}" > VERSION

# Commit the version update
git add VERSION
git commit -m "Release $new_version" || true

# Create and push the tag
git tag -a $new_version -m "Release $new_version"
git push origin main --tags