load("@bazel_gazelle//:deps.bzl", "go_repository")

def linter_dependencies():
  go_repository(
    name = "com_github_timakin_bodyclose",
    build_file_proto_mode = "disable_global",
    importpath = "github.com/timakin/bodyclose",
    sum = "h1:ig99OeTyDwQWhPe2iw9lwfQVF1KB3Q4fpP3X7/2VBG8=",
    version = "v0.0.0-20200424151742-cb6215831a94",
  )

  go_repository(
      name = "com_github_gostaticanalysis_analysisutil",
      build_file_proto_mode = "disable_global",
      importpath = "github.com/gostaticanalysis/analysisutil",
      version = "v0.7.1",
      sum = "h1:ZMCjoue3DtDWQ5WyU16YbjbQEQ3VuzwxALrpYd+HeKk=",
  )

  go_repository(
      name = "com_github_gostaticanalysis_comment",
      build_file_proto_mode = "disable_global",
      importpath = "github.com/gostaticanalysis/comment",
      version = "v1.4.2",
      sum = "h1:hlnx5+S2fY9Zo9ePo4AhgYsYHbM2+eAv8m/s1JiCd6Q=",
  )

  go_repository(
      name = "com_github_golangci_gofmt",
      build_file_proto_mode = "disable_global",
      importpath = "github.com/golangci/gofmt",
      version = "v0.0.0-20220901101216-f2edd75033f2",
      sum = "h1:amWTbTGqOZ71ruzrdA+Nx5WA3tV1N0goTspwmKCQvBY=",
  )
