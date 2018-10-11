package migrations

//go:generate env GOBIN=../../../../.bin GO111MODULE=on go install github.com/kevinburke/go-bindata/go-bindata
//go:generate ../../../../.bin/go-bindata -nometadata -pkg migrations -prefix ../../../../migrations/ -ignore README.md ../../../../migrations/
//go:generate gofmt -w bindata.go
