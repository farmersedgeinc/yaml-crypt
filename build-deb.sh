#!/bin/bash

# Required env variables:
# - GOARCH
# - VERSION

set -e

[[ -z "$GOARCH" ]] && echo 'Required variable: $GOARCH' && exit 1
[[ -z "$VERSION" ]] && echo 'Required variable: $VERSION' && exit 1

export GOARCH
export VERSION
export GOOS=linux

PKGDIR="$(mktemp -d /tmp/deb.XXXXXX)"
trap 'rm -r "$PKGDIR"' EXIT

TARBALL="out/yaml-crypt.linux.$GOARCH.tar.gz"
[[ ! -f "$TARBALL" ]] && ./build-tarball.sh

tar -xzf "$TARBALL" -C "$PKGDIR"

# Install Control File
mkdir -p "$PKGDIR/DEBIAN"
cat <<EOF > "$PKGDIR/DEBIAN/control"
Package: yaml-crypt
Description: Encrypt secret strings in your yaml config files using a cloud-based encryption service, while leaving the rest of the file readable.
Version: $(echo "$VERSION" | sed 's:^refs/tags/v::g')
Architecture: $GOARCH
Installed-Size: $(( ( "$(du -bs "$PKGDIR"| cut -f 1)" - "$(du -bs "$PKGDIR/DEBIAN"| cut -f 1)" ) / 1024 ))
Maintainer: https://github.com/farmersedgeinc
EOF

mkdir -p out
dpkg-deb --root-owner-group --build "$PKGDIR" "out/yaml-crypt.$GOARCH.deb"
