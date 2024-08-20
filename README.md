# yaml-crypt

Encrypt secret strings in your yaml config files using a cloud-based encryption service, while leaving the rest of the file readable.

![Test](https://github.com/farmersedgeinc/yaml-crypt/workflows/Test/badge.svg?branch=tests)

## Concepts

Each file managed by yaml-crypt has 3 "versions": the _encrypted version_, the _decrypted version_, and the _plain version_. The _encrypted version_ has any secret values in the file replaced with encrypted represenations; the _decrypted version_ contains plaintext values, with any secrets marked with the tag `!secret`, and the _plain version_ contains a plain yaml file, with all yaml-crypt-specific tags removed.

The **encrypted version** of a file is what you want to track in git; all non-secret values are visible as normal, while encrypted representations of secrets live inline in the document flow, only changing when their contents change. This means that git diffs will look normal and be perfectly readable, with the exception that the secret values are not visible, other than revealing _whether they have changed at all_. With the exception of the secret values, this file is also directly editable; if someone does not have access to the encryption service, they are still able to view and edit non-secret values.

The **decrypted version** of a file is the fully editable version of a file. Secret values are shown as plaintext YAML strings, prefixed with the YAML tag `!secret`. New secrets can be added by simply adding values with the `!secret` tag to any YAML string value in the file. To avoid conflicts, always run `decrypt` on a file before editing it, in case the underlying _encrypted version_ has changed.

The **plain version** of a file is basically just the _decrypted version_ without the `!secret` tags. This means that this file is basically read-only; since there are no tags, yaml-crypt has no way of knowing which values should be encrypted, so don't edit this file! This file is generated for external applications to consume.

## Requirements

### Google

You need valid Google Cloud credentials set up. 
* Install [Google Cloud SDK](https://cloud.google.com/sdk/gcloud/)
* If you've never, ever logged-in via gcloud: `gcloud auth login`
* `gcloud auth application-default set-quota-project shared-92kdnmcv0fk`
* `gcloud auth application-default login`

Your account needs access to Google [Cloud KMS](https://cloud.google.com/security-key-management), and the role `roles/cloudkms.cryptoKeyEncrypterDecrypter` for the key to be used.

## Installation

Pre-built binaries are available for a variety of operating systems [here](https://github.com/farmersedgeinc/yaml-crypt/releases/latest)

## Usage

```
yaml-crypt --help
```

Although, mostly you'll just `yaml-crypt edit` to edit files, and `yaml-crypt decrypt --plain` in CI scripts.

If you're performing bulk edits on many files, you can run `yaml-crypt` before editing, and `yaml-crypt encrypt` afterwards.

To **create a new file**, just create a file with the _decrypted version_ suffix, (by default, that's `.decrypted.yaml`), and add your content, prefixing any string values you want to protect with the `!secret` YAML tag, and run `yaml-crypt encrypt <yourfile>`, and `git add` the new _encrypted version_ (by default, `<yourfile>.encrypted.yaml`).

To **set up a new repo**, run `yaml-crypt init <provider>` with the name of the encryption provider (currently, the only supported one is `google`). A `.yamlcrypt.yaml` file will be created, containing all the configuration for your repository, as well as some keys with blank values in the `config` section, for configuring the provider.

### Note About Editors

**If you're not the sort of nerd who customizes your environment, you probably don't need to worry about this.** `yaml-crypt edit` is basically the equivalent of running `yaml-crypt decrypt "$FILE" && "$EDITOR" "$FILE" && yaml-crypt encrypt "$FILE"`. This process makes one critical assumption: that your editor will only exit after you've finished editing the file. This holds true for any terminal-based text editor (`vim`, `nano`, `emacs`, etc), and for some GUI editors like `gedit` and `mousepad`. However, Sublime Text (`subl`), Atom (`atom`), and VSCode (`code`), all fork to a background process and immediately exit, which breaks the core assumption of `yaml-crypt edit`. `subl`, `atom`, and `code` all accept a `-w` flag to make the process wait for the window/tab to be closed before exiting though. You can set `EDITORFLAGS=-w` in your shell config (`.bashrc`, etc) to fix editing if your `$EDITOR` is `subl`, `atom`, or `code`.

### Decrypted Git Diffs

To see decrypted secret values in your git diffs, add the following to your repo's `.gitattributes`:

```
*.encrypted.yaml diff=yamlcrypt
```

Assuming you have yaml-crypt globally installed on your machine, you can then add the following to either your local repo config (`.git/config`), or your global git config (`~/.gitconfig`):

```
[diff "yamlcrypt"]
    textconv = yaml-crypt decrypt --stdout --progress=false
    cachetextconv = true
```

Setting `cachetextconv = true` may help with performance by having git cache the plaintexts of each git object. **Warning:** you should only be using yaml-crypt on a secure machine anyways, but be aware that git's local `textconv` cache is additional place that sensitive data can reside.

### Integration with [helm-secrets](https://github.com/jkroepke/helm-secrets)

To use yaml-crypt as a custom backend for [helm-secrets](https://github.com/jkroepke/helm-secrets), set the `HELM_SECRETS_BACKEND` environment variable to point to the `helm-secrets/backend.sh` file from this repo. The tarball and deb packages place this file in `/usr/share/yaml-crypt/helm-secrets/_backend.sh`, and provide a file at `/usr/share/yaml-crypt/helm-secret/setup.sh` that you can source in your shell config to automatically set the needed variables.

On MacOS, the paths mentioned start with `/usr/local/share` instead of `/usr/share`.

Once helm-secrets is intalled and this is configured, helm-secrets will transparently decrypt any values files prefixed with `secret://` using yaml-crypt.

#### Example: Workstation Setup

This assumes yaml-crypt has been installed from one of the packages in Releases:

```
helm plugin install https://github.com/jkroepke/helm-secrets
# replace .bashrc with .zshrc or whatever your shell's equivalent is
cat <<<EOF >> ~/.bashrc
[ -f /usr/share/yaml-crypt/helm-secrets/setup.sh ] && . /usr/share/yaml-crypt/helm-secrets/setup.sh
[ -f /usr/local/share/yaml-crypt/helm-secrets/setup.sh ] && . /usr/local/share/yaml-crypt/helm-secrets/setup.sh
EOF
```

#### Example: CI worker setup

In the Dockerfile for your helm image:

```
# install helm-secrets
RUN helm plugin install https://github.com/jkroepke/helm-secrets
# install yaml-crypt
RUN curl -L "https://github.com/farmersedgeinc/yaml-crypt/releases/download/v$YAML_CRYPT_VERSION/yaml-crypt.linux.$ARCH.tar.gz" | tar -xzvC /
# enable helm-secrets backend
ENV HELM_SECRETS_BACKEND=/usr/share/yaml-crypt/helm-secrets/_backend.sh
# disable cache
ENV SECRET_BACKEND_ARGS --no-cache
```

## Security Notes

Yaml-crypt stores a cache of ciphertexts and plaintexts in the directory `.yamlcrypt.cache` at the root of the repo. This cache is obviously very sensitive, as it contains a mapping between encrypted and decrypted values! Yaml-crypt automatically adds the cache directory, and the suffixes for the _decrypted_ and _plain_ versions of files to the `.gitignore`, but it is still the user's responsibility to make sure to protect these files and make sure they never end up in git history!

## Examples

```
% cat <<EOF > testfile.decrypted.yaml
key: value
just: a regular value
except: !secret this one is a secret,
but: not this one
EOF
% ./yaml-crypt encrypt testfile.decrypted.yaml
% cat testfile.encrypted.yaml
key: value
just: a regular value
except: !encrypted CiQAGygkv+rQz7w4siamOGllx/2WLW1vRJljndLAeaShWFJh8/ESPgDnaSWzeXJ+9wtBoG/j+Y3VHn+5AZP78aTMBsIVgR5s5h4om58otx/Tdez+iTy0ZVkVDEPrcPsDQ2JPuxvU
but: not this one
% ./yaml-crypt decrypt testfile.encrypted.yaml
% cat testfile.decrypted.yaml
key: value
just: a regular value
except: !secret this one is a secret,
but: not this one
```

## Development

Pretty simple, just builds with `go build`. Test with `go test -v ./...`; the tests will automatically detect if you have credentials for the crypto providers, however, if you have credentials that aren't valid the tests will fail.

Make sure to `go fmt` any code you submit!
