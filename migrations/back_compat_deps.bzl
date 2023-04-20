load("@bazel_gazelle//:deps.bzl", "go_repository")

def patched_deps():
    go_repository(
        name = "back_compat_com_github_sourcegraph_scip",
        # This fixes the build for sourcegraph/scip which depends on sourcegraph/sourcegraph/lib but
        # gazelle doesn't know how to resolve those packages from within sourcegraph/scip.
        build_directives = [
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/errors @sourcegraph_back_compat//lib/errors",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol @sourcegraph_back_compat//lib/codeintel/lsif/protocol",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader @sourcegraph_back_compat//lib/codeintel/lsif/protocol/reader",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/writer @sourcegraph_back_compat//lib/codeintel/lsif/protocol/writer",
        ],
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sourcegraph/scip",
        sum = "h1:fWPxLkDObzzKTGe9vb6wpzK0FYkwcfSxmxUBvAOc8aw=",
        version = "v0.2.4-0.20221213205653-aa0e511dcfef",
    )

    go_repository(
        name = "back_compat_com_github_sourcegraph_conc",
        build_directives = [
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/errors @sourcegraph_back_compat//lib/errors",
        ],
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sourcegraph/conc",
        sum = "h1:96VpOCAtXDCQ8Oycz0ftHqdPyMi8w12ltN4L2noYg7s=",
        version = "v0.2.0",
    )
