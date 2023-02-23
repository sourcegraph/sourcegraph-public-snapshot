building repo-updater for the App app package

not sure it's actually used by App, though

```
VERSION=4.4.0
pkg=github.com/sourcegraph/sourcegraph/cmd/repo-updater
for a in arm64 amd64
do
GOARCH=${a} go build -trimpath -ldflags \
"-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  \
-X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
-buildmode exe -tags dist \
-o "$(basename "$pkg")-${a}" "$pkg"
done
lipo $(basename "$pkg")-arm64 $(basename "$pkg")-amd64  -create -output $(basename "$pkg")-universal
```
