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
        name = "com_github_ashanbrown_forbidigo",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/ashanbrown/forbidigo",
        sum = "h1:WXhzLjOlnuDYPYQo/eFlcFMi8X/kLfvWLYu6CSoebis=",
        version = "v1.5.1",
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
        name = "com_github_gordonklaus_ineffassign",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/gordonklaus/ineffassign",
        version = "v0.0.0-20230107090616-13ace0543b28",
        sum = "h1:9alfqbrhuD+9fLZ4iaAVwhlp5PEhmnBt7yvK2Oy5C1U=",
    )

    go_repository(
        name = "com_github_go_critic_go_critic",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-critic/go-critic",
        version = "v0.6.7",
        sum = "h1:1evPrElnLQ2LZtJfmNDzlieDhjnq36SLgNzisx06oPM=",
    )

    go_repository(
        name = "com_github_go_toolsmith_astcast",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/astcast",
        version = "v1.1.0",
        sum = "h1:+JN9xZV1A+Re+95pgnMgDboWNVnIMMQXwfBwLRPgSC8=",
    )

    go_repository(
        name = "com_github_go_toolsmith_astcopy",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/astcopy",
        version = "v1.0.2",
        sum = "h1:YnWf5Rnh1hUudj11kei53kI57quN/VH6Hp1n+erozn0=",
    )

    go_repository(
        name = "com_github_go_toolsmith_astequal",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/astequal",
        version = "v1.1.0",
        sum = "h1:kHKm1AWqClYn15R0K1KKE4RG614D46n+nqUQ06E1dTw=",
    )

    go_repository(
        name = "com_github_go_toolsmith_astfmt",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/astfmt",
        version = "v1.1.0",
        sum = "h1:iJVPDPp6/7AaeLJEruMsBUlOYCmvg0MoCfJprsOmcco=",
    )

    go_repository(
        name = "com_github_go_toolsmith_astp",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/astp",
        version = "v1.1.0",
        sum = "h1:dXPuCl6u2llURjdPLLDxJeZInAeZ0/eZwFJmqZMnpQA=",
    )

    go_repository(
        name = "com_github_go_toolsmith_pkgload",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/pkgload",
        version = "v1.0.2-0.20220101231613-e814995d17c5",
        sum = "h1:eD9POs68PHkwrx7hAB78z1cb6PfGq/jyWn3wJywsH1o=",
    )

    go_repository(
        name = "com_github_go_toolsmith_strparse",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/strparse",
        version = "v1.1.0",
        sum = "h1:GAioeZUK9TGxnLS+qfdqNbA4z0SSm5zVNtCQiyP2Bvw=",
    )

    go_repository(
        name = "com_github_go_toolsmith_typep",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/go-toolsmith/typep",
        version = "v1.1.0",
        sum = "h1:fIRYDyF+JywLfqzyhdiHzRop/GQDxxNhLGQ6gFUNHus=",
    )

    go_repository(
        name = "com_github_quasilyte_go_ruleguard",
        importpath = "github.com/quasilyte/go-ruleguard",
        version = "v0.3.19",
        sum = "h1:tfMnabXle/HzOb5Xe9CUZYWXKfkS1KwRmZyPmD9nVcc=",
    )

    go_repository(
        name = "com_github_quasilyte_gogrep",
        importpath = "github.com/quasilyte/gogrep",
        version = "v0.5.0",
        sum = "h1:eTKODPXbI8ffJMN+W2aE0+oL0z/nh8/5eNdiO34SOAo=",
    )

    go_repository(
        name = "com_github_quasilyte_regex_syntax",
        importpath = "github.com/quasilyte/regex/syntax",
        version = "v0.0.0-20200407221936-30656e2c4a95",
        sum = "h1:L8QM9bvf68pVdQ3bCFZMDmnt9yqcMBro1pC7F+IPYMY=",
    )

    go_repository(
        name = "com_github_quasilyte_stdinfo",
        importpath = "github.com/quasilyte/stdinfo",
        version = "v0.0.0-20220114132959-f7386bf02567",
        sum = "h1:M8mH9eK4OUR4lu7Gd+PU1fV2/qnDNfzT635KRSObncs=",
    )

    go_repository(
        name = "org_golang_x_exp_typeparams",
        importpath = "golang.org/x/exp/typeparams",
        version = "v0.0.0-20230203172020-98cc5a0785f9",
        sum = "h1:6WHiuFL9FNjg8RljAaT7FNUuKDbvMqS1i5cr2OE2sLQ=",
    )

    go_repository(
        name = "com_github_kyoh86_exportloopref",
        importpath = "github.com/kyoh86/exportloopref",
        version = "v0.1.11",
        sum = "h1:1Z0bcmTypkL3Q4k+IDHMWTcnCliEZcaPiIe0/ymEyhQ=",
    )

    go_repository(
        name = "co_honnef_go_tools",
        importpath = "honnef.co/go/tools",
        version = "v0.4.3",
        sum = "h1:o/n5/K5gXqk8Gozvs2cnL0F2S1/g1vcGCAx2vETjITw=",
    )

    go_repository(
        name = "com_github_openpeedeep_depguard_v2",
        importpath = "github.com/OpenPeeDeeP/depguard/v2",
        version = "v2.0.1",
        sum = "h1:yr9ZswukmNxl/hmJHEoLEjCF1d+f2pQrC0m1jzVljAE=",
    )

    go_repository(
        name = "cc_mvdan_unparam",
        importpath = "mvdan.cc/unparam",
        version = "v0.0.0-20230312165513-e84e2d14e3b8",
        sum = "h1:VuJo4Mt0EVPychre4fNlDWDuE5AjXtPJpRUWqZDQhaI=",
    )

    go_repository(
        name = "com_4d63_gocheckcompilerdirectives",
        importpath = "4d63.com/gocheckcompilerdirectives",
        version = "v1.2.1",
        sum = "h1:AHcMYuw56NPjq/2y615IGg2kYkBdTvOaojYCBcRE7MA=",
    )
