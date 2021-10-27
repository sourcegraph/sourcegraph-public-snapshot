# Workarounds for M1 mac local development

1. Download Chromium for Puppeteer separately: https://linguinecode.com/post/how-to-fix-m1-mac-puppeteer-chromium-arm64-bug.
2. Run `softwareupdate --install-rosetta`.
3. Get the Mac version of Jaeger https://www.jaegertracing.io/download/, extract it, then `cp ~/Downloads/jaeger-1.27.0-darwin-amd64/jaeger-all-in-one ~/sourcegraph/.bin/jaeger-all-in-one-1.18.1-darwin-arm64`.

Did you bump into another issue and solve it locally? Consider updating this list! 🙇
