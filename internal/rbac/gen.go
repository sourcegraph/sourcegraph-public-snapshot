package rbac

//go:generate env GO111MODULE=on go run yamldata.go -pkg rbac -o constants.go -lang go
//go:generate gofmt -s -w constants.go
//go:generate env GO111MODULE=on go run yamldata.go -pkg rbac -o ../../client/web/src/rbac/constants.ts -lang ts
