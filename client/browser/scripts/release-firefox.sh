#!/usr/bin/env bash

set -ex

# Setup
pnpm build
rm -rf build/web-ext
mkdir -p build/web-ext

# Sign the bundle
#
# Due to an limitation of the web-ext package, we can not rely on the exit
# status to know if the status was successful.
#
# c.f. https://github.com/mozilla/web-ext/issues/804
set +e
tmp="$(mktemp)"
ok="Your add-on has been submitted for review."
pnpm dlx web-ext sign --source-dir ./build/firefox --artifacts-dir ./build/web-ext --api-key "$FIREFOX_AMO_ISSUER" --api-secret "$FIREFOX_AMO_SECRET" |
  sed -n "s/\($ok\).*$/\0/;1,/$ok/p" |
  tee "$tmp"
error=${PIPESTATUS[0]}
if ! grep -q "$ok" "$tmp" && [ "$error" = 1 ]; then
  exit "$error"
fi
