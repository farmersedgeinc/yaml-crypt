# Guide for users with encrypt-only permissions

Much of `yaml-crypt`'s functionality depends on both encrypt and decrypt operations. Because `yaml-crypt`'s goal is to provide *stable* encrypted representations of secrets (in order to improve comprehensibility of git diffs), the encryption method must be deterministic, and Google Cloud KMS does not offer a deterministic option. Thus, `yaml-crypt` stores a cache of all plaintext:ciphertext pairs, so that ciphertexts can be reused, allowing it to simulate deterministic encryption using non-deterministic primitives.

Unfortunately this means most of the ease-of-use functionality is not available with encrypt-only access to a repo's Cloud KMS keyring.

If you just want to duplicate a reference to an existing secret, you can just copy the opaque value, making sure to include the leading `!encrypted` tag.

If you want to add a new secret, you can encrypt it with `yaml-crypt encrypt-value --no-cache`. This will read a single line from STDIN and print an encrypted blob to STDOUT. For a multi-line secret, add the `--multi-line` flag, in which case it will read from STDIN until it is closed (either through shell redirection, or your terminal's EOF shortcut, typically control+D).

Make sure to add a `!encrypted` YAML tag prefix to the value, in order to signal to yaml-crypt that the value is encrypted.

## Examples

To encrypt the string "super-secret":

```
$ echo super-secret | yaml-crypt encrypt-value --no-cache
```

```
$ cat something-sensitive.pem | yaml-crypt encrypt-value --no-cache --multi-line
```

Paste the output into a `.encrypted.yaml` file, with the `!encrypted` YAML tag:

```
example-key: !encrypted GreatBigBlobOfBase64==
```
