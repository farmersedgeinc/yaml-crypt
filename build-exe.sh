#!/bin/bash

# Required env variables:
# - GOARCH
# - VERSION

set -Eeo pipefail

[[ -z "$GOARCH" ]] && echo 'Required variable: $GOARCH' && exit 1
[[ -z "$VERSION" ]] && echo 'Required variable: $VERSION' && exit 1

set -u

GOOS=windows GOARCH="$GOARCH" go build \
    -ldflags "-X 'github.com/farmersedgeinc/yaml-crypt/cmd.version=$(echo "$VERSION" | sed 's:^refs/tags/v::g')'" \
    -o "out/yaml-crypt.$GOARCH.exe"
