//go:generate bash ../build-assets.sh
//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/shurcooL/vfsgen/cmd/vfsgendev
//go:generate $PWD/.bin/vfsgendev -source="sourcegraph.com/cmd/management-console/assets".Assets
//go:generate go run replace_hack.go

package assets
