package graphqlbackend

//go:generate env GO111MODULE=on go run gen/schema_generate.go
//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/efritz/go-mockgen
//go:generate $PWD/.bin/go-mockgen -f github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend -i prometheusQuerier -o prometheus_mock.go
