package migrations

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/kevinburke/go-bindata/go-bindata
//go:generate $PWD/.bin/go-bindata -nometadata -pkg migrations -prefix ../../../../migrations/ -ignore README.md ../../../../migrations/
//go:generate gofmt -w bindata.go
