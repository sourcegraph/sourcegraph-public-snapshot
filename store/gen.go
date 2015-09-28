package store

//go:generate gen-mocks -w -i=.+ -o mockstore -outpkg mockstore -name_prefix=
//go:generate go run gen_context_and_mock.go -o1 context.go -o2 mockstore/mockstores.go -o3 cli/cli.go
