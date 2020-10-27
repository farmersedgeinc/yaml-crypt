#!/bin/bash

# Required env variables:
# - GOOS
# - GOARCH
# - VERSION

set -e

[[ -z "$GOOS" ]] && echo 'Required variable: $GOOS' && exit 1
[[ -z "$GOARCH" ]] && echo 'Required variable: $GOARCH' && exit 1
[[ -z "$VERSION" ]] && echo 'Required variable: $VERSION' && exit 1

unset GOOS GOARCH

PKGDIR="$(mktemp -d /tmp/tarball.XXXXXX)"
trap 'rm -r "$PKGDIR"' EXIT

# Install Binary
BINDIR="$PKGDIR/usr/bin"
mkdir -p "$BINDIR"

GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags "-X 'github.com/farmersedgeinc/yaml-crypt/cmd.version=$(echo "$VERSION" | sed 's:^refs/tags/v::g')'" \
    -o "$BINDIR/yaml-crypt"

# Install Bash Completions
if [[ "$GOOS" -eq "linux" ]]; then
    BASHFILE="$PKGDIR/etc/bash_completion.d/yaml-crypt"
else
    BASHFILE="$PKGDIR/usr/local/etc/bash_completion.d/yaml-crypt"
fi
mkdir -p "$(dirname "$BASHFILE")"
go run main.go completion bash > "$BASHFILE"

# Install Zsh Completions
if [[ "$GOOS" -eq "linux" ]]; then
    ZSHFILE="$PKGDIR/usr/share/zsh/vendor-completions/_yaml-crypt"
    mkdir -p "$(dirname "$ZSHFILE")"
    go run main.go completion zsh > "$ZSHFILE"
fi

# Install Fish Completions
if [[ "$GOOS" -eq "linux" ]]; then
    FISHFILE="$PKGDIR/usr/share/fish/vendor_completions.d/yaml-crypt.fish"
    mkdir -p "$(dirname "$FISHFILE")"
    go run main.go completion fish > "$FISHFILE"
fi

mkdir -p out
tar -cC "$PKGDIR" "." | gzip -9 > out/yaml-crypt.$GOOS.$GOARCH.tar.gz
