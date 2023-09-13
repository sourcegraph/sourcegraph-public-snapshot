load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")

# go_stringer provides an equivalent to `//go:generate stringer`.
# Files can be updated and generated with `bazel run //dev:write_all_generated`.
def go_stringer(src, typ, name, additional_args = []):
    output_file = "_out_" + typ.lower() + "_string.go_in"
    output_file_source = typ.lower() + "_string.go"

    native.genrule(
        name = name,
        srcs = [src],  # Accessed below using `$<`.
        outs = [output_file],
        # golang.org/x/tools executes commands via
        # golang.org/x/sys/execabs which requires all PATH lookups to
        # result in absolute paths. To account for this, we resolve the
        # relative path returned by location to an absolute path.
        cmd = """\
GO_ABS_PATH=`cd $$(dirname $(location @go_sdk//:bin/go)) && pwd`
GO_SDK_ABS_PATH=`dirname $$GO_ABS_PATH`

env \
    PATH=$$GO_ABS_PATH \
    GOCACHE=$$(mktemp -d) \
    GOROOT=$$GO_SDK_ABS_PATH \
    $(location @org_golang_x_tools//cmd/stringer:stringer) \
        -output=$@ \
        -type={typ} \
        {args} \
        $<; \

# Because stringer will add a comment on the file saying how it generated that file, we end up
# in a delicate case, where the comment mentions the path the bazel sandbox, which varies
# depending on your OS (like bazel-out/darwin-fastbuild or bazel-out/linux-fastbuild) which
# in turns breaks the diff_test that ensure that file is correctly up to date in CI.
# So to make it work, we replace that OS dependent string with a something that's the same
# across all envs, even if it's slighly inaccurate.
sed -i'' -e 's=$@='`basename $@`'=' $@
""".format(
            typ = typ,
            args = " ".join(additional_args),
        ),
        visibility = [":__pkg__", "//pkg/gen:__pkg__"],
        tools = [
            "@go_sdk//:bin/go",
            "@go_sdk//:srcs",
            "@go_sdk//:tools",
            "@org_golang_x_tools//cmd/stringer",
        ],
    )

    write_source_files(
        name = "write_" + name,
        files = {
            output_file_source: output_file,
        },
        tags = ["go_generate"],
        suggested_update_target = "//dev:write_all_generated",
        visibility = ["//visibility:public"],
    )
