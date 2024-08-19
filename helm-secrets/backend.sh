#!/usr/bin/env sh

_YAML_CRYPT="${HELM_SECRETS_YAML_CRYPT_PATH:-${HELM_SECRETS_YAML_CRYPT_BIN:-"yaml-crypt"}}"

_yaml_crypt() {
    # shellcheck disable=SC2086
    set -- ${SECRET_BACKEND_ARGS} "$@"
    (set -x; $_YAML_CRYPT "$@")
}

_yaml_crypt_filename() {
    printf -- %s "$1" \
        | sed -Ee 's/\.(encrypted|decrypted|plain)\.yaml$//' \
        | xargs -I{} -- printf -- %s.%s.yaml {} "$2"
}

_yaml_crypt_filename_encrypted() {
    _yaml_crypt_filename "$1" encrypted
}

_yaml_crypt_filename_decrypted() {
    _yaml_crypt_filename "$1" decrypted
}

_yaml_crypt_filename_plain() {
    _yaml_crypt_filename "$1" plain
}

_custom_backend_is_file_encrypted() {
    _yaml_crypt_filename_encrypted "$1" | grep '\.encrypted\.yaml$' > /dev/null
}

_custom_backend_encrypt_file() {
    fatal "Encrypting files is not supported"
}
_yaml_crypt_type() {
    type="${1}"
    if [ "$type" != "yaml" ] && [ "$type" != "auto" ]; then
        fatal "Type $type not supported"
    fi
}

_custom_backend_decrypt_file() {
    _yaml_crypt_type "$1"
    input="$(_yaml_crypt_filename_encrypted "$2")"
    output="$3"
    if [ "${output}" = "" ] || [ "${output}" = "-" ]; then
        _yaml_crypt decrypt --progress=false --plain --stdout "$input"
    else
        _yaml_crypt decrypt --progress=false --plain --stdout "$input" > "$output"
    fi
}

_custom_backend_decrypt_literal() {
    _yaml_crypt decrypt-value
}

_custom_backend_edit_file() {
    _yaml_crypt_type "$1"
    _yaml_crypt edit "$2"
}
