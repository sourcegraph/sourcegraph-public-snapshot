package migrations

//go:generate go build -o ../../../../vendor/.bin/go-bindata github.com/sourcegraph/sourcegraph/vendor/github.com/kevinburke/go-bindata/go-bindata
//go:generate ../../../../vendor/.bin/go-bindata -nometadata -pkg migrations -prefix ../../../../migrations/ -ignore README.md ../../../../migrations/
//go:generate gofmt -w bindata.go
