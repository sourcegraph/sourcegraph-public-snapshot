package rbac

//go:generate env GO111MODULE=on go run yamldata.go -o constants.go -lang go -kind constants
//go:generate env GO111MODULE=on go run yamldata.go -o types/namaespace.go -lang go -kind types
//go:generate gofmt -s -w constants.go
//go:generate gofmt -s -w types/namespace.go
//go:generate env GO111MODULE=on go run yamldata.go -o ../../client/web/src/rbac/constants.ts -lang ts -kind constants
