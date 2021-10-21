# Workarounds for M1 mac local development

1. Download Chromium for Puppeteer separately: https://linguinecode.com/post/how-to-fix-m1-mac-puppeteer-chromium-arm64-bug.
2. Run `softwareupdate --install-rosetta`.
3. Download this custom `musl-cross` Homebrew formula and run `brew install ./musl-cross.rb`: https://github.com/FiloSottile/homebrew-musl-cross/pull/30/files.
4. Run `go build ./cmd/docsite` in the [docsite repo](https://github.com/sourcegraph/docsite), then `cp docsite ~/sourcegraph/.bin/docsite_vINSERT_LATEST_DOCSITE_VERSION_HERE`. Get docsite version: `grep DOCSITE_VERSION sg.config.yaml` in the sourcegraph repo.
5. Get the Mac version of Jaeger https://www.jaegertracing.io/download/, extract it, then `cp ~/Downloads/jaeger-1.27.0-darwin-amd64/jaeger-all-in-one ~/sourcegraph/.bin/jaeger-all-in-one-1.18.1-darwin-arm64`.

Did you bump into another issue and solve it locally? Consider updating this list! 🙇
