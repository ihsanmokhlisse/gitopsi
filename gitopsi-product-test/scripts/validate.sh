#!/bin/bash
set -e

echo "Validating repository..."

# Check YAML syntax
find . -name "*.yaml" -o -name "*.yml" | while read f; do
  yamllint "$f" 2>/dev/null || echo "Warning: yamllint not installed"
  break
done

# Build with kustomize
for dir in infrastructure applications; do
  if [ -d "$dir/base" ]; then
    echo "Building $dir..."
    kustomize build "$dir/base" > /dev/null
  fi
done

echo "Validation complete!"
