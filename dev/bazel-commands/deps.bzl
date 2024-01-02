load("@bazel_gazelle//:deps.bzl", "go_repository")

def custom_commands_dependencies():
    go_repository(
        name = "build_aspect_cli",
        build_file_proto_mode = "disable_global",
        importpath = "aspect.build/cli",
        sum = "h1:UI53K9QCbwbmSaQygJAYov/09Y1UBXNUGixqZ36yHAM=",
        version = "v1.508.19",
    )
    go_repository(
        name = "com_github_alphadose_haxmap",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/alphadose/haxmap",
        sum = "h1:C/2LboOnPCZP27GmmSXOcwx360st0P8N0fTJ3voefKc=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_bazel_contrib_rules_jvm",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazel-contrib/rules_jvm",
        sum = "h1:oJxriPrBPGxugqkQiFceyzuBYgaFiUBTFM8xhrErPKs=",
        version = "v0.17.1-0.20230814153054-0ce5d051291b",
    )
    go_repository(
        name = "com_github_bazelbuild_bazel_gazelle",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazelbuild/bazel-gazelle",
        sum = "h1:YHkwssgvCXDRU7sLCq1kGqaGYO9pKNR1Wku7UT2LhoY=",
        version = "v0.34.0",
    )
    go_repository(
        name = "com_github_bazelbuild_bazelisk",
        build_file_proto_mode = "disable_global",
        build_file_generation = "on",
        importpath = "github.com/bazelbuild/bazelisk",
        sum = "h1:TDt+a1PYrnBF9on3WRJUisXXFhCMrhcNo8OebyS5Q34=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_github_bazelbuild_buildtools",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:VUHCI4QRifAGYsbVJYqJndLf7YqV12YthB+PLFsEKqo=",
        version = "v0.0.0-20231017121127-23aa65d4e117",
    )
    go_repository(
        name = "com_github_bazelbuild_rules_go",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bazelbuild/rules_go",
        sum = "h1:aY2smc3JWyUKOjGYmOKVLX70fPK9ON0rtwQojuIeUHc=",
        version = "v0.42.0",
    )
    go_repository(
        name = "com_github_bgentry_go_netrc",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bgentry/go-netrc",
        sum = "h1:xDfNPAt8lFiC1UJrqV3uuy861HCTo708pDMbjHHdCas=",
        version = "v0.0.0-20140422174119-9fd32a8b3d3d",
    )
    go_repository(
        name = "com_github_bmatcuk_doublestar_v4",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bmatcuk/doublestar/v4",
        sum = "h1:FH9SifrbvJhnlQpztAx++wlkk70QBf0iBWDwNy7PA4I=",
        version = "v4.6.1",
    )
    go_repository(
        name = "com_github_bufbuild_protocompile",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/bufbuild/protocompile",
        sum = "h1:LbFKd2XowZvQ/kajzguUp2DC9UEIQhIq77fZZlaQsNA=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_googleapis_google_cloud_go_testing",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/googleapis/google-cloud-go-testing",
        sum = "h1:tlyzajkF3030q6M8SvmJSemC9DTHL/xaMa18b65+JM4=",
        version = "v0.0.0-20200911160855-bcd43fbb19e8",
    )
    go_repository(
        name = "com_github_manifoldco_promptui",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/manifoldco/promptui",
        sum = "h1:3V4HzJk1TtXW1MTZMP7mdlwbBpIinw3HztaIlYthEiA=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_msolo_jsonr",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/msolo/jsonr",
        sum = "h1:A8w3/wSixa6RqN8bLYQ9mayk+d8xxq6jQCmlKs/jCRo=",
        version = "v0.0.0-20231023064044-62fbfc3a0313",
    )
    go_repository(
        name = "com_github_sagikazarmark_crypt",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sagikazarmark/crypt",
        sum = "h1:96E1qrToLBU6fGzo+PRRz7KGOc9FkYFiPnR3/zf8Smg=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_github_tejzpr_ordered_concurrently_v3",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/tejzpr/ordered-concurrently/v3",
        sum = "h1:TLHtzlQEDshbmGveS8S+hxLw4s5u67aoJw5LLf+X2xY=",
        version = "v3.0.1",
    )
    go_repository(
        name = "com_github_twmb_murmur3",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/twmb/murmur3",
        sum = "h1:8Yt9taO/WN3l08xErzjeschgZU2QSrwm1kclYq+0aRg=",
        version = "v1.1.8",
    )
    go_repository(
        name = "com_github_yargevad_filepathx",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/yargevad/filepathx",
        sum = "h1:SYcT+N3tYGi+NvazubCNlvgIPbzAk7i7y2dwg3I5FYc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "org_golang_x_tools_go_vcs",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/tools/go/vcs",
        sum = "h1:cOIJqWBl99H1dH5LWizPa+0ImeeJq3t3cJjaeOWUAL4=",
        version = "v0.1.0-deprecated",
    )
