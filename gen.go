package sourcegraph

//go:generate go run remove_protobuf_json_snake_case_tags.go -w go-sourcegraph/sourcegraph Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-diff/diff Godeps/_workspace/src/sourcegraph.com/sourcegraph/go-vcs/vcs Godeps/_workspace/src/sourcegraph.com/sourcegraph/srclib/graph Godeps/_workspace/src/sourcegraph.com/sourcegraph/srclib/unit Godeps/_workspace/src/sourcegraph.com/sourcegraph/vcsstore/vcsclient
