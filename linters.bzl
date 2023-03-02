load("@bazel_gazelle//:deps.bzl", "go_repository")

def linter_dependencies():
  go_repository(
    name = "com_github_timakin_bodyclose",
    build_file_proto_mode = "disable_global",
    importpath = "github.com/timakin/bodyclose",
    # NOTE: Gazelle will not generate this for you
    # To retrieve these values you can create a go.mod in this directory
    # and then run `go mod tidy` and delete the go.mod you created.
    sum = "h1:ig99OeTyDwQWhPe2iw9lwfQVF1KB3Q4fpP3X7/2VBG8=",
    version = "v0.0.0-20200424151742-cb6215831a94",
  )
