package rbac

//go:generate env GO111MODULE=on go run yamldata.go -o constants.go -lang go -kind constants
//go:generate env GO111MODULE=on go run yamldata.go -o types/namespace.go -lang go -kind namespace
//go:generate env GO111MODULE=on go run yamldata.go -o types/action.go -lang go -kind action
//go:generate gofmt -s -w constants.go
//go:generate gofmt -s -w types/namespace.go
//go:generate gofmt -s -w types/action.go
//go:generate env GO111MODULE=on go run yamldata.go -o ../../client/web/src/rbac/constants.ts -lang ts -kind constants
