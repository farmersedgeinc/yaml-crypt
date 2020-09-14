#!/bin/bash

# Required env variables:
# - GOOS
# - GOARCH
# - VERSION

[[ -z "$GOOS" ]] && echo 'Required variable: $GOOS' && exit 1
[[ -z "$GOARCH" ]] && echo 'Required variable: $GOARCH' && exit 1
[[ -z "$VERSION" ]] && echo 'Required variable: $VERSION' && exit 1

PKGDIR="$(mktemp -d /tmp/tarball.XXXXXX)"
trap 'rm -r "$PKGDIR"' EXIT

# Install Binary
BINDIR="$PKGDIR/usr/bin"
mkdir -p "$BINDIR"
GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags "-X 'github.com/farmersedgeinc/yaml-crypt/cmd.version=$VERSION'" \
    -o "$BINDIR/yaml-crypt"

# Install Bash Completions
if [[ "$GOOS" -eq "linux" ]]; then
    BASHFILE="$PKGDIR/etc/bash_completion.d/yaml-crypt"
else
    BASHFILE="$PKGDIR/usr/local/etc/bash_completion.d/yaml-crypt"
fi
mkdir -p "$(dirname "$BASHFILE")"
"$BINDIR/yaml-crypt" completion bash > "$BASHFILE"

# Install Zsh Completions
if [[ "$GOOS" -eq "linux" ]]; then
    ZSHFILE="$PKGDIR/usr/share/zsh/vendor-completions/_yaml-crypt"
    mkdir -p "$(dirname "$ZSHFILE")"
    "$BINDIR/yaml-crypt" completion zsh > "$ZSHFILE"
fi

# Install Fish Completions
if [[ "$GOOS" -eq "linux" ]]; then
    FISHFILE="$PKGDIR/usr/share/fish/vendor_completions.d/yaml-crypt.fish"
    mkdir -p "$(dirname "$FISHFILE")"
    "$BINDIR/yaml-crypt" completion fish > "$FISHFILE"
fi

mkdir -p out
tar -cC "$PKGDIR" "." | gzip -9 > out/yaml-crypt.$GOOS.$GOARCH.tar.gz
