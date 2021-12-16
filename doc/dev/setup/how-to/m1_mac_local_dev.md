# Workarounds for M1 Mac local development

## sg

`sg` seems to run into "too many open file" errors if downloaded with `curl` as described in "[Quickstart](../quickstart.md)".

The current workaround is to build `sg` yourself. To do that, run the following in the [`sourcegraph/sourcegraph`](https://github.com/sourcegraph/sourcegraph) repository:

```
./dev/sg/install.sh
```

This will print a location to where `sg` was installed. Use that `sg` to run `sg start`.

## Rosetta

Docker [requires Rosetta](https://docs.docker.com/desktop/mac/apple-silicon/#system-requirements) to run `amd64` binaries. It should be installed by default, but if that wasn't the case, run `softwareupdate --install-rosetta`.

## Jaeger

[Get the Mac version of Jaeger](https://github.com/jhchabran/jaeger/releases/download/v1.28.1/jaeger-1.28.1-darwin-arm64.tar.gz), extract it, then

```
cd ~/Downloads
curl https://github.com/jhchabran/jaeger/releases/download/v1.28.1/jaeger-1.28.1-darwin-arm64.tar.gz -L | tar -xz
cp ~/Downloads/jaeger-1.28.1-darwin-arm64/jaeger-all-in-one ~/code/sourcegraph/.bin/jaeger-all-in-one-1.18.1-darwin-arm64
```

(adjust `~/code/sourcegraph` to point to you local clone of `github.com/sourcegraph/sourcegraph`).

> NOTE: Did you bump into another issue and solve it locally? Consider updating this list! 🙇
