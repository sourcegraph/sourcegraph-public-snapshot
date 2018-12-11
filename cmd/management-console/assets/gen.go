//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/shurcooL/vfsgen/cmd/vfsgendev
//go:generate $PWD/.bin/vfsgendev -source="github.com/sourcegraph/sourcegraph/cmd/management-console/assets".Assets

package assets
