# Building p4-fusion

In order to import Perforce depots into Sourcegraph we first convert them into git repositories. We use an open source tool called [p4-fusion](https://github.com/salesforce/p4-fusion):

> A fast Perforce depot to Git repository converter using the Helix Core C/C++ API as an attempt to mitigate the performance bottlenecks in git-p4.py.

[Building](https://github.com/salesforce/p4-fusion#build) p4-fusion can be a little tricky as it depends on some older libraries and also doesn't build on M1 Apple laptops. To get around this we use [nix](https://nixos.org).

## How to build

Below are the instructions for building p4-fusion locally, assuming you have the [Sourcegraph repository](https://github.com/sourcegraph/sourcegraph) checked out.

1. Follow [these instruction](https://nixos.org/download.html) to install Nix. (Tested with version 2.15.0)
1. Navigate to the root of your Sourcegraph directory
1. Run `nix build ".#p4-fusion" --verbose --extra-experimental-features nix-command --extra-experimental-features flakes`

If the build completes successfully you should have a `p4-fusion` binary in `./result/bin/p4-fusion` which you can copy somewhere in your `$PATH`

## Troubleshooting

### `nix build` fails with "hash mismatch" error referencing `helix-core-api.drv`

The `p4-fusion` dependencies specified in [the `srcs` array of `dev/nix/p4-fusion.nix`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@8d513301f8a12f7c7e0b5a66ed20ba14b6679cca/-/blob/dev/nix/p4-fusion.nix?L45-77) are sometimes updated without being versioned, so their hashes change, which causes "hash mismatch" errors.

**Example**

```bash
building '/nix/store/cb2ssx8vykl1ghb4k87yp3q6wnvfvjj2-helix-core-api.drv'...
error: hash mismatch in fixed-output derivation '/nix/store/cb2ssx8vykl1ghb4k87yp3q6wnvfvjj2-helix-core-api.drv':
         specified: sha256-8yX9sjE1f798ns0UmHXz5I8kdpUXHS01FG46SU8nsZw=
            got:    sha256-gaYvQOX8nvMIMHENHB0+uklyLcmeXT5gjGGcVC9TTtE=
error: 1 dependencies of derivation '/nix/store/m409z1rq40bwzvvndbnghrrxm000zd9v-p4-fusion.drv' failed to build
```

To update the hashes, we use `nix` tools to calculate the hashes of the unpacked archives, and convert the format to an SRI representation that includes the type of the hash.

Get the URL of the archive from the part of the `srcs` array that matches the OS and architecture of your local system.

**Example**

Replace `<url of archive>` with the value of the archive url from `srcs`

```bash
$ hash_type=sha256
$ hash_value=$(nix-prefetch-url --type "${hash_type}" --unpack <url of archive>)
$ nix --extra-experimental-features nix-command hash to-sri "${hash_type}:${hash_value}"
```

Copy the output from that sequence of commands and paste it into the value of the `hash` field in the `fetchzip` attribute set.

Since when one changes, they all probably change, here is an example of getting the updated hashes for all of the archives ([current archive URLs; double-check that they are still correct](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/nix/p4-fusion.nix))

```bash
hash_type=sha256
for url in \
"https://cdist2.perforce.com/perforce/r22.2/bin.macosx12arm64/p4api-openssl1.1.1.tgz" \
"https://cdist2.perforce.com/perforce/r22.2/bin.macosx12x86_64/p4api-openssl1.1.1.tgz" \
"https://cdist2.perforce.com/perforce/r22.2/bin.linux26x86_64/p4api-glibc2.3-openssl1.1.1.tgz"
do
  echo "${url}"
  hash_value=$(nix-prefetch-url --type "${hash_type}" --unpack "${url}")
  nix --extra-experimental-features nix-command hash to-sri "${hash_type}:${hash_value}"
  echo
done
```
