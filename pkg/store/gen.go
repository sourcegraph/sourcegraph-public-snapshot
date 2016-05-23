package store

//go:generate sh -c "env GOBIN=`pwd`/../../vendor/.bin go install ../../vendor/sourcegraph.com/sourcegraph/gen-mocks"
//go:generate ../../vendor/.bin/gen-mocks -w -i=.+ -o mockstore -outpkg mockstore -name_prefix=
//go:generate go run ../../api/sourcegraph/mock/gen/goreplace.go -from ChannelNotification -to store.ChannelNotification mockstore/channel_mock.go
//go:generate go run gen_context_and_mock.go -o1 context.go -o2 mockstore/mockstores.go
