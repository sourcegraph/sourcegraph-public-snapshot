package rbac

//go:generate env GO111MODULE=on go run yamldata.go -pkg rbac -o constants.go
//go:generate gofmt -s -w constants.go
