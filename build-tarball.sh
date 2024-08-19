#!/bin/bash

# Required env variables:
# - GOOS
# - GOARCH
# - VERSION

set -e

[[ -z "$GOOS" ]] && echo 'Required variable: $GOOS' && exit 1
[[ -z "$GOARCH" ]] && echo 'Required variable: $GOARCH' && exit 1
[[ -z "$VERSION" ]] && echo 'Required variable: $VERSION' && exit 1


PKGDIR="$(mktemp -d /tmp/tarball.XXXXXX)"
trap 'rm -r "$PKGDIR"' EXIT

function pkg_prefix {
    REL_DIR="${1#/}"
    if [ "$GOOS" == "darwin" ]; then
        printf %s/usr/local/%s "$PKGDIR" "${REL_DIR#usr/}"
    else
        printf %s/%s "$PKGDIR" "$REL_DIR"
    fi
}

function pkg_file {
    PKG_PATH="$(pkg_prefix "$1")"
    dirname "$PKG_PATH" \
        | xargs mkdir -p
    printf -- %s "$PKG_PATH"
}

function install_completion {
    GOOS="" GOARCH="" go run main.go completion "$1" > "$(pkg_file "$2")"
}

function install_helm_secrets_backend {
    local BACKEND_PATH
    BACKEND_PATH="$(pkg_file /usr/share/yaml-crypt/helm-secrets/_backend.sh)"
    cp helm-secrets/backend.sh "$BACKEND_PATH"
    cat <<EOF > "$(pkg_file /usr/share/yaml-crypt/helm-secrets/setup.sh)"
export HELM_SECRETS_BACKEND="${BACKEND_PATH#"$PKGDIR"}"
export HELM_SECRETS_YAML_CRYPT_BIN="${BIN_PATH#"$PKGDIR"}"
EOF
}

function ldflags {
    printf -- \
        "-X 'github.com/farmersedgeinc/yaml-crypt/cmd.version=%s'" \
        "${VERSION#refs/tags/v}"
}

BIN_PATH="$(pkg_file /usr/bin/yaml-crypt)"
# Install Binary
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$(ldflags)" -o "$BIN_PATH"

# Install Completions
install_completion bash /etc/bash_completion.d/yaml-crypt
install_completion zsh /usr/share/zsh/vendor-completions/_yaml-crypt
install_completion fish /usr/share/fish/vendor_completions.d/yaml-crypt.fish

# Install helm-secrets backend
install_helm_secrets_backend

mkdir -p out
tar --owner 1 --group 1 -cC "$PKGDIR" "." | gzip -9 > out/yaml-crypt.$GOOS.$GOARCH.tar.gz
