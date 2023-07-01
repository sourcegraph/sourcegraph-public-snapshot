load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")

# go_stringer provides an equivalent to `//go:generate stringer`.
# Files can be updated and generated with `bazel run //dev:write_all`.
def go_stringer(src, typ, name, additional_args=[]):
    output_file = typ.lower() + "_string.go"

    native.genrule(
        name = name,
        srcs = [src],  # Accessed below using `$<`.
        outs = [output_file],
        # golang.org/x/tools executes commands via
        # golang.org/x/sys/execabs which requires all PATH lookups to
        # result in absolute paths. To account for this, we resolve the
        # relative path returned by location to an absolute path.
        cmd = """\
GO_REL_PATH=`dirname $(location @go_sdk//:bin/go)`
GO_ABS_PATH=`cd $$GO_REL_PATH && pwd`
# Set GOPATH to something to workaround https://github.com/golang/go/issues/43938
env PATH=$$GO_ABS_PATH HOME=$(GENDIR) GOPATH=/nonexist-gopath \
$(location @org_golang_x_tools//cmd/stringer:stringer) -output=$@ -type={typ} {args} $<
""".format(
         typ = typ,
         args = " ".join(additional_args),
        ),
        visibility = [":__pkg__", "//pkg/gen:__pkg__"],
        exec_tools = [
            "@go_sdk//:bin/go",
            "@org_golang_x_tools//cmd/stringer",
        ],
    )

    write_source_files(
        name = "write_" + name,
        files = {
            output_file: output_file,
        },
        tags = ["go_generate"],
        suggested_update_target = "//dev:write_all",
        visibility = ["//visibility:public"],
    )

