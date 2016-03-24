package mock

//go:generate sh -c "env GOBIN=`pwd`/../../../vendor/.bin go install ../../../vendor/sourcegraph.com/sourcegraph/gen-mocks"
//go:generate ../../../vendor/.bin/gen-mocks -p .. -w -i=.+(Server|Client|Service)$ -o . -outpkg mock -name_prefix= -no_pass_args=opts

//go:generate go run gen_client_helpers.go

// The pbtypes package selector is emitted as pbtypes1 when more than
// one pbtypes type is used. Fix this up so that goimports works.
//
//go:generate go run gen/goreplace.go -from "pbtypes1" -to "pbtypes" sourcegraph.pb_mock.go

//go:generate go get golang.org/x/tools/cmd/goimports
//go:generate goimports -w sourcegraph.pb_mock.go
