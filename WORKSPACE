load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.20.1/rules_go-v0.20.1.tar.gz",
    ],
    sha256 = "842ec0e6b4fbfdd3de6150b61af92901eeb73681fd4d185746644c338f51d4c0",
)

http_archive(
    name = "bazel_gazelle",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.19.0/bazel-gazelle-v0.19.0.tar.gz",
    ],
    sha256 = "41bff2a0b32b02f20c227d234aa25ef3783998e5453f7eade929704dcff7cd4b",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "cc_mvdan_interfacer",
    importpath = "mvdan.cc/interfacer",
    sum = "h1:WX1yoOaKQfddO/mLzdV4wptyWgoH/6hwLs7QHTixo0I=",
    version = "v0.0.0-20180901003855-c20040233aed",
)

go_repository(
    name = "cc_mvdan_lint",
    importpath = "mvdan.cc/lint",
    sum = "h1:DxJ5nJdkhDlLok9K6qO+5290kphDJbHOQO1DFFFTeBo=",
    version = "v0.0.0-20170908181259-adc824a0674b",
)

go_repository(
    name = "cc_mvdan_unparam",
    importpath = "mvdan.cc/unparam",
    sum = "h1:ujHaqw4YdQP41+lVP79aaouzIU8ZIRMT9/S3YUru+JI=",
    version = "v0.0.0-20190917161559-b83a221c10a2",
)

go_repository(
    name = "co_honnef_go_tools",
    importpath = "honnef.co/go/tools",
    sum = "h1:3JgtbtFHMiCmsznwGVTUWbgGov+pVqnlf1dEJTNAXeM=",
    version = "v0.0.1-2019.2.3",
)

go_repository(
    name = "com_github_akavel_rsrc",
    importpath = "github.com/akavel/rsrc",
    sum = "h1:zjWn7ukO9Kc5Q62DOJCcxGpXC18RawVtYAGdz2aLlfw=",
    version = "v0.8.0",
)

go_repository(
    name = "com_github_alcortesm_tgz",
    importpath = "github.com/alcortesm/tgz",
    sum = "h1:uSoVVbwJiQipAclBbw+8quDsfcvFjOpI5iCf4p/cqCs=",
    version = "v0.0.0-20161220082320-9c5fe88206d7",
)

go_repository(
    name = "com_github_alecthomas_assert",
    importpath = "github.com/alecthomas/assert",
    sum = "h1:smF2tmSOzy2Mm+0dGI2AIUHY+w0BUc+4tn40djz7+6U=",
    version = "v0.0.0-20170929043011-405dbfeb8e38",
)

go_repository(
    name = "com_github_alecthomas_chroma",
    importpath = "github.com/alecthomas/chroma",
    sum = "h1:1hKci+AyKOxJrugR9veaocu9DQGR2/GecI72BpaO0Rg=",
    version = "v0.6.7",
)

go_repository(
    name = "com_github_alecthomas_colour",
    importpath = "github.com/alecthomas/colour",
    sum = "h1:JHZL0hZKJ1VENNfmXvHbgYlbUOvpzYzvy2aZU5gXVeo=",
    version = "v0.0.0-20160524082231-60882d9e2721",
)

go_repository(
    name = "com_github_alecthomas_kong",
    importpath = "github.com/alecthomas/kong",
    sum = "h1:C4Q9m+oXOxcSWwYk9XzzafY2xAVAaeubZbUHJkw3PlY=",
    version = "v0.2.1-0.20190708041108-0548c6b1afae",
)

go_repository(
    name = "com_github_alecthomas_kong_hcl",
    importpath = "github.com/alecthomas/kong-hcl",
    sum = "h1:atLL+K8Hg0e8863K2X+k7qu+xz3M2a/mWFIACAPf55M=",
    version = "v0.1.8-0.20190615233001-b21fea9723c8",
)

go_repository(
    name = "com_github_alecthomas_repr",
    importpath = "github.com/alecthomas/repr",
    sum = "h1:GDQdwm/gAcJcLAKQQZGOJ4knlw+7rfEQQcmwTbt4p5E=",
    version = "v0.0.0-20181024024818-d37bc2a10ba1",
)

go_repository(
    name = "com_github_alecthomas_template",
    importpath = "github.com/alecthomas/template",
    sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
    version = "v0.0.0-20190718012654-fb15b899a751",
)

go_repository(
    name = "com_github_alecthomas_units",
    importpath = "github.com/alecthomas/units",
    sum = "h1:Hs82Z41s6SdL1CELW+XaDYmOH4hkBN4/N9og/AsOv7E=",
    version = "v0.0.0-20190717042225-c3de453c63f4",
)

go_repository(
    name = "com_github_andygrunwald_go_gerrit",
    importpath = "github.com/andygrunwald/go-gerrit",
    sum = "h1:9PvNa6zH6gOW4VVfbAx5rjDLpxunG+RSaXQB+8TEv4w=",
    version = "v0.0.0-20190825170856-5959a9bf9ff8",
)

go_repository(
    name = "com_github_anmitsu_go_shlex",
    importpath = "github.com/anmitsu/go-shlex",
    sum = "h1:kFOfPq6dUM1hTo4JG6LR5AXSUEsOjtdm0kw0FtQtMJA=",
    version = "v0.0.0-20161002113705-648efa622239",
)

go_repository(
    name = "com_github_apache_thrift",
    importpath = "github.com/apache/thrift",
    sum = "h1:pODnxUFNcjP9UTLZGTdeh+j16A8lJbRvD3rOtrk/7bs=",
    version = "v0.12.0",
)

go_repository(
    name = "com_github_armon_consul_api",
    importpath = "github.com/armon/consul-api",
    sum = "h1:G1bPvciwNyF7IUmKXNt9Ak3m6u9DE1rF+RmtIkBpVdA=",
    version = "v0.0.0-20180202201655-eb2c6b5be1b6",
)

go_repository(
    name = "com_github_armon_go_socks5",
    importpath = "github.com/armon/go-socks5",
    sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
    version = "v0.0.0-20160902184237-e75332964ef5",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    sum = "h1:/4+rDPe0W95KBmNGYCG+NUvdL8ssPYBMxL+aSCg6nIA=",
    version = "v1.17.7",
)

go_repository(
    name = "com_github_aws_aws_sdk_go_v2",
    importpath = "github.com/aws/aws-sdk-go-v2",
    sum = "h1:GE6BJJLTfsv1VqwwnY2KcUlrgQ8GtZAM83jeQChgNDM=",
    version = "v0.14.0",
)

go_repository(
    name = "com_github_azure_go_ansiterm",
    importpath = "github.com/Azure/go-ansiterm",
    sum = "h1:w+iIsaOQNcT7OZ575w+acHgRric5iCyQh+xv+KJ4HB8=",
    version = "v0.0.0-20170929234023-d6e3b3328b78",
)

go_repository(
    name = "com_github_beevik_etree",
    importpath = "github.com/beevik/etree",
    sum = "h1:T0xke/WvNtMoCqgzPhkX2r4rjY3GDZFi+FjpRZY2Jbs=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_beorn7_perks",
    importpath = "github.com/beorn7/perks",
    sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_bitly_go_hostpool",
    importpath = "github.com/bitly/go-hostpool",
    sum = "h1:mXoPYz/Ul5HYEDvkta6I8/rnYM5gSdSV2tJ6XbZuEtY=",
    version = "v0.0.0-20171023180738-a3a6125de932",
)

go_repository(
    name = "com_github_bmizerany_assert",
    importpath = "github.com/bmizerany/assert",
    sum = "h1:DDGfHa7BWjL4YnC6+E63dPcxHo2sUxDIu8g3QgEJdRY=",
    version = "v0.0.0-20160611221934-b7ed37b82869",
)

go_repository(
    name = "com_github_boj_redistore",
    importpath = "github.com/boj/redistore",
    sum = "h1:RmdPFa+slIr4SCBg4st/l/vZWVe9QJKMXGO60Bxbe04=",
    version = "v0.0.0-20180917114910-cd5dcc76aeff",
)

go_repository(
    name = "com_github_bombsimon_wsl",
    importpath = "github.com/bombsimon/wsl",
    sum = "h1:26f0nCKNzDod+z/9J4eK4jOcYgGTuy7NcYBL/t6pFQQ=",
    version = "v1.2.3",
)

go_repository(
    name = "com_github_burntsushi_toml",
    importpath = "github.com/BurntSushi/toml",
    sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_burntsushi_xgb",
    importpath = "github.com/BurntSushi/xgb",
    sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
    version = "v0.0.0-20160522181843-27f122750802",
)

go_repository(
    name = "com_github_certifi_gocertifi",
    importpath = "github.com/certifi/gocertifi",
    sum = "h1:8k9FLYBLKT+9v2HQJ/a95ZemmTx+/ltJcAiRhVushG8=",
    version = "v0.0.0-20190905060710-a5e0173ced67",
)

go_repository(
    name = "com_github_cespare_xxhash",
    importpath = "github.com/cespare/xxhash",
    sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_client9_misspell",
    importpath = "github.com/client9/misspell",
    sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
    version = "v0.3.4",
)

go_repository(
    name = "com_github_cockroachdb_apd",
    importpath = "github.com/cockroachdb/apd",
    sum = "h1:3LFP3629v+1aKXU5Q37mxmRxX/pIu1nijXydLShEq5I=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_cockroachdb_cockroach_go",
    importpath = "github.com/cockroachdb/cockroach-go",
    sum = "h1:2zRrJWIt/f9c9HhNHAgrRgq0San5gRRUJTBXLkchal0=",
    version = "v0.0.0-20181001143604-e0a95dfd547c",
)

go_repository(
    name = "com_github_codahale_hdrhistogram",
    importpath = "github.com/codahale/hdrhistogram",
    sum = "h1:qMd81Ts1T2OTKmB4acZcyKaMtRnY5Y44NuXGX2GFJ1w=",
    version = "v0.0.0-20161010025455-3a0bb77429bd",
)

go_repository(
    name = "com_github_containerd_containerd",
    importpath = "github.com/containerd/containerd",
    sum = "h1:xjvXQWABwS2uiv3TWgQt5Uth60Gu86LTGZXMJkjc7rY=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_coreos_bbolt",
    importpath = "github.com/coreos/bbolt",
    sum = "h1:wZwiHHUieZCquLkDL0B8UhzreNWsPHooDAG3q34zk0s=",
    version = "v1.3.2",
)

go_repository(
    name = "com_github_coreos_etcd",
    importpath = "github.com/coreos/etcd",
    sum = "h1:jFneRYjIvLMLhDLCzuTuU4rSJUjRplcJQ7pD7MnhC04=",
    version = "v3.3.10+incompatible",
)

go_repository(
    name = "com_github_coreos_go_etcd",
    importpath = "github.com/coreos/go-etcd",
    sum = "h1:bXhRBIXoTm9BYHS3gE0TtQuyNZyeEMux2sDi4oo5YOo=",
    version = "v2.0.0+incompatible",
)

go_repository(
    name = "com_github_coreos_go_oidc",
    importpath = "github.com/coreos/go-oidc",
    sum = "h1:sdJrfw8akMnCuUlaZU3tE/uYXFgfqom8DBE9so9EBsM=",
    version = "v2.1.0+incompatible",
)

go_repository(
    name = "com_github_coreos_go_semver",
    importpath = "github.com/coreos/go-semver",
    sum = "h1:wkHLiw0WNATZnSG7epLsujiMCgPAc9xhjJ4tgnAxmfM=",
    version = "v0.3.0",
)

go_repository(
    name = "com_github_coreos_go_systemd",
    importpath = "github.com/coreos/go-systemd",
    sum = "h1:Wf6HqHfScWJN9/ZjdUKyjop4mf3Qdd+1TvvltAvM3m8=",
    version = "v0.0.0-20190321100706-95778dfbb74e",
)

go_repository(
    name = "com_github_coreos_pkg",
    importpath = "github.com/coreos/pkg",
    sum = "h1:lBNOc5arjvs8E5mO2tbpBpLoyyu8B6e44T7hJy6potg=",
    version = "v0.0.0-20180928190104-399ea9e2e55f",
)

go_repository(
    name = "com_github_cosiner_argv",
    importpath = "github.com/cosiner/argv",
    sum = "h1:2iAFN+sWPktbZ4tvxm33Ei8VY66FPCxdOxpncUGpAXE=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_cpuguy83_go_md2man",
    importpath = "github.com/cpuguy83/go-md2man",
    sum = "h1:BSKMNlYxDvnunlTymqtgONjNnaRV1sTpcovwwjF22jk=",
    version = "v1.0.10",
)

go_repository(
    name = "com_github_creack_pty",
    importpath = "github.com/creack/pty",
    sum = "h1:6pwm8kMQKCmgUg0ZHTm5+/YvRK0s3THD/28+T6/kk4A=",
    version = "v1.1.7",
)

go_repository(
    name = "com_github_crewjam_saml",
    importpath = "github.com/crewjam/saml",
    sum = "h1:+cz/lCIhz+eg8+jC8cWk5LBLbbpH39IKyHliN6GZyUE=",
    version = "v0.0.0-20190521120225-344d075952c9",
)

go_repository(
    name = "com_github_cznic_b",
    importpath = "github.com/cznic/b",
    sum = "h1:UHFGPvSxX4C4YBApSPvmUfL8tTvWLj2ryqvT9K4Jcuk=",
    version = "v0.0.0-20180115125044-35e9bbe41f07",
)

go_repository(
    name = "com_github_cznic_fileutil",
    importpath = "github.com/cznic/fileutil",
    sum = "h1:7uSNgsgcarNk4oiN/nNkO0J7KAjlsF5Yv5Gf/tFdHas=",
    version = "v0.0.0-20180108211300-6a051e75936f",
)

go_repository(
    name = "com_github_cznic_golex",
    importpath = "github.com/cznic/golex",
    sum = "h1:CVAqftqbj+exlab+8KJQrE+kNIVlQfJt58j4GxCMF1s=",
    version = "v0.0.0-20170803123110-4ab7c5e190e4",
)

go_repository(
    name = "com_github_cznic_internal",
    importpath = "github.com/cznic/internal",
    sum = "h1:FHpbUtp2K8X53/b4aFNj4my5n+i3x+CQCZWNuHWH/+E=",
    version = "v0.0.0-20180608152220-f44710a21d00",
)

go_repository(
    name = "com_github_cznic_lldb",
    importpath = "github.com/cznic/lldb",
    sum = "h1:AIA+ham6TSJ+XkMe8imQ/g8KPzMUVWAwqUQQdtuMsHs=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_cznic_mathutil",
    importpath = "github.com/cznic/mathutil",
    sum = "h1:XNT/Zf5l++1Pyg08/HV04ppB0gKxAqtZQBRYiYrUuYk=",
    version = "v0.0.0-20180504122225-ca4c9f2c1369",
)

go_repository(
    name = "com_github_cznic_ql",
    importpath = "github.com/cznic/ql",
    sum = "h1:lcKp95ZtdF0XkWhGnVIXGF8dVD2X+ClS08tglKtf+ak=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_cznic_sortutil",
    importpath = "github.com/cznic/sortutil",
    sum = "h1:hxuZop6tSoOi0sxFzoGGYdRqNrPubyaIf9KoBG9tPiE=",
    version = "v0.0.0-20150617083342-4c7342852e65",
)

go_repository(
    name = "com_github_cznic_strutil",
    importpath = "github.com/cznic/strutil",
    sum = "h1:0rkFMAbn5KBKNpJyHQ6Prb95vIKanmAe62KxsrN+sqA=",
    version = "v0.0.0-20171016134553-529a34b1c186",
)

go_repository(
    name = "com_github_cznic_zappy",
    importpath = "github.com/cznic/zappy",
    sum = "h1:YKKpTb2BrXN2GYyGaygIdis1vXbE7SSAG9axGWIMClg=",
    version = "v0.0.0-20160723133515-2533cb5b45cc",
)

go_repository(
    name = "com_github_daaku_go_zipexe",
    importpath = "github.com/daaku/go.zipexe",
    sum = "h1:VSOgZtH418pH9L16hC/JrgSNJbbAL26pj7lmD1+CGdY=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_danwakefield_fnmatch",
    importpath = "github.com/danwakefield/fnmatch",
    sum = "h1:y5HC9v93H5EPKqaS1UYVg1uYah5Xf51mBfIoWehClUQ=",
    version = "v0.0.0-20160403171240-cbb64ac3d964",
)

go_repository(
    name = "com_github_datadog_zstd",
    importpath = "github.com/DataDog/zstd",
    sum = "h1:3oxKN3wbHibqx897utPC2LTQU4J+IHWWJO+glkAkpFM=",
    version = "v1.4.1",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    importpath = "github.com/davecgh/go-spew",
    sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_daviddengcn_go_colortext",
    importpath = "github.com/daviddengcn/go-colortext",
    sum = "h1:d/cVoZOrJPJHKH1NdeUjyVAWKp4OpOT+Q+6T1sH7jeU=",
    version = "v0.0.0-20180409174941-186a3d44e920",
)

go_repository(
    name = "com_github_denisenkom_go_mssqldb",
    importpath = "github.com/denisenkom/go-mssqldb",
    sum = "h1:tkum0XDgfR0jcVVXuTsYv/erY2NnEDqwRojbxR1rBYA=",
    version = "v0.0.0-20190515213511-eb9f6a1743f3",
)

go_repository(
    name = "com_github_depado_bfchroma",
    importpath = "github.com/Depado/bfchroma",
    sum = "h1:kYP1mmZw7G4//kglQl/2xgu7Zd7u4psxTLTjQfBB5sw=",
    version = "v1.1.2",
)

go_repository(
    name = "com_github_dghubble_gologin",
    importpath = "github.com/dghubble/gologin",
    replace = "github.com/sourcegraph/gologin",
    sum = "h1:K7hzuWsJGoU8ILHJzrXxsuvXvLHpP/g4iUk7VFj2lY8=",
    version = "v1.0.2-0.20181110030308-c6f1b62954d8",
)

go_repository(
    name = "com_github_dgrijalva_jwt_go",
    importpath = "github.com/dgrijalva/jwt-go",
    sum = "h1:7qlOGliEKZXTDg6OTjfoBKDXWrumCAMpl/TFQ4/5kLM=",
    version = "v3.2.0+incompatible",
)

go_repository(
    name = "com_github_dgryski_go_sip13",
    importpath = "github.com/dgryski/go-sip13",
    sum = "h1:RMLoZVzv4GliuWafOuPuQDKSm1SJph7uCRnnS61JAn4=",
    version = "v0.0.0-20181026042036-e10d5fee7954",
)

go_repository(
    name = "com_github_dhui_dktest",
    importpath = "github.com/dhui/dktest",
    sum = "h1:NVUdB50k8tml431Ho1hcQBNeC52Qe8oSDPAjseA67Y8=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_dlclark_regexp2",
    importpath = "github.com/dlclark/regexp2",
    sum = "h1:8sAhBGEM0dRWogWqWyQeIJnxjWO6oIjl8FKqREDsGfk=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_dnaeon_go_vcr",
    importpath = "github.com/dnaeon/go-vcr",
    sum = "h1:r8L/HqC0Hje5AXMu1ooW8oyQyOFv4GxqpL0nRP7SLLY=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_docker_distribution",
    importpath = "github.com/docker/distribution",
    sum = "h1:a5mlkVzth6W5A4fOsS3D2EO5BUmsJpcB+cRlLU7cSug=",
    version = "v2.7.1+incompatible",
)

go_repository(
    name = "com_github_docker_docker",
    importpath = "github.com/docker/docker",
    sum = "h1:mzrx39dGtGq0VEnTHjnakmczd4uFbhx2cZU3BJDsLdc=",
    version = "v0.7.3-0.20190817195342-4760db040282",
)

go_repository(
    name = "com_github_docker_go_connections",
    importpath = "github.com/docker/go-connections",
    sum = "h1:El9xVISelRB7BuFusrZozjnkIM5YnzCViNKohAFqRJQ=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_docker_go_units",
    importpath = "github.com/docker/go-units",
    sum = "h1:3uh0PgVws3nIA0Q+MwDC8yjEPf9zjRfZZWXZYDct3Tw=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_eapache_go_resiliency",
    importpath = "github.com/eapache/go-resiliency",
    sum = "h1:1NtRmCAqadE2FN4ZcN6g90TP3uk8cg9rn9eNK2197aU=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_eapache_go_xerial_snappy",
    importpath = "github.com/eapache/go-xerial-snappy",
    sum = "h1:YEetp8/yCZMuEPMUDHG0CW/brkkEp8mzqk2+ODEitlw=",
    version = "v0.0.0-20180814174437-776d5712da21",
)

go_repository(
    name = "com_github_eapache_queue",
    importpath = "github.com/eapache/queue",
    sum = "h1:YOEu7KNc61ntiQlcEeUIoDTJ2o8mQznoNvUhiigpIqc=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_edsrzf_mmap_go",
    importpath = "github.com/edsrzf/mmap-go",
    sum = "h1:aaQcKT9WumO6JEJcRyTqFVq4XUZiUcKR2/GI31TOcz8=",
    version = "v0.0.0-20170320065105-0bce6a688712",
)

go_repository(
    name = "com_github_emersion_go_imap",
    importpath = "github.com/emersion/go-imap",
    sum = "h1:/7HHNiSOk13DErenBZaQfTBmUy+quc6X7s3RNnuVtUM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_emersion_go_message",
    importpath = "github.com/emersion/go-message",
    sum = "h1:OYhqtJI4eOLvGtRIsUfP87VMJ1J/o6ks1tah9DlYkn4=",
    version = "v0.10.4-0.20190609165112-592ace5bc1ca",
)

go_repository(
    name = "com_github_emersion_go_sasl",
    importpath = "github.com/emersion/go-sasl",
    sum = "h1:ba7YsgX5OV8FjGi5ZWml8Jng6oBrJAb3ahqWMJ5Ce8Q=",
    version = "v0.0.0-20190817083125-240c8404624e",
)

go_repository(
    name = "com_github_emersion_go_textwrapper",
    importpath = "github.com/emersion/go-textwrapper",
    sum = "h1:40SWqY0zE3qCi6ZrtTf5OUdNm5lDnGnjRSq9GgmeTrg=",
    version = "v0.0.0-20160606182133-d0e65e56babe",
)

go_repository(
    name = "com_github_emirpasic_gods",
    importpath = "github.com/emirpasic/gods",
    sum = "h1:QAUIPSaCu4G+POclxeqb3F+WPpdKqFGlw36+yOzGlrg=",
    version = "v1.12.0",
)

go_repository(
    name = "com_github_ericchiang_k8s",
    importpath = "github.com/ericchiang/k8s",
    sum = "h1:vxrMwEzY43oxu8aZyD/7b1s8tsBM+xoUoxjWECWFbPI=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_etdub_goparsetime",
    importpath = "github.com/etdub/goparsetime",
    sum = "h1:iguwbR+9xsizl84VMHU47I4OOWYSex1HZRotEoqziWQ=",
    version = "v0.0.0-20160315173935-ea17b0ac3318",
)

go_repository(
    name = "com_github_facebookgo_clock",
    importpath = "github.com/facebookgo/clock",
    sum = "h1:yDWHCSQ40h88yih2JAcL6Ls/kVkSE8GFACTGVnMPruw=",
    version = "v0.0.0-20150410010913-600d898af40a",
)

go_repository(
    name = "com_github_facebookgo_ensure",
    importpath = "github.com/facebookgo/ensure",
    sum = "h1:0JZ+dUmQeA8IIVUMzysrX4/AKuQwWhV2dYQuPZdvdSQ=",
    version = "v0.0.0-20160127193407-b4ab57deab51",
)

go_repository(
    name = "com_github_facebookgo_limitgroup",
    importpath = "github.com/facebookgo/limitgroup",
    sum = "h1:IeaD1VDVBPlx3viJT9Md8if8IxxJnO+x0JCGb054heg=",
    version = "v0.0.0-20150612190941-6abd8d71ec01",
)

go_repository(
    name = "com_github_facebookgo_muster",
    importpath = "github.com/facebookgo/muster",
    sum = "h1:a4DFiKFJiDRGFD1qIcqGLX/WlUMD9dyLSLDt+9QZgt8=",
    version = "v0.0.0-20150708232844-fd3d7953fd52",
)

go_repository(
    name = "com_github_facebookgo_stack",
    importpath = "github.com/facebookgo/stack",
    sum = "h1:JWuenKqqX8nojtoVVWjGfOF9635RETekkoH6Cc9SX0A=",
    version = "v0.0.0-20160209184415-751773369052",
)

go_repository(
    name = "com_github_facebookgo_subset",
    importpath = "github.com/facebookgo/subset",
    sum = "h1:E2s37DuLxFhQDg5gKsWoLBOB0n+ZW8s599zru8FJ2/Y=",
    version = "v0.0.0-20150612182917-8dac2c3c4870",
)

go_repository(
    name = "com_github_fatih_astrewrite",
    importpath = "github.com/fatih/astrewrite",
    sum = "h1:qzMLs39qKQl6FKU/yNIDKSPrfV56il5U24BX2NvRrxE=",
    version = "v0.0.0-20190527122930-f5295d6854fb",
)

go_repository(
    name = "com_github_fatih_color",
    importpath = "github.com/fatih/color",
    sum = "h1:DkWD4oS2D8LGGgTQ6IvwJJXSL5Vp2ffcQg58nFV38Ys=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_felixfbecker_stringscore",
    importpath = "github.com/felixfbecker/stringscore",
    sum = "h1:LjuZCUHGVEmJy/t1Kt+KM1kWUCiYiQTLT0Zhalwaulg=",
    version = "v0.0.0-20170928081130-e71a9f1b0749",
)

go_repository(
    name = "com_github_felixge_httpsnoop",
    importpath = "github.com/felixge/httpsnoop",
    sum = "h1:lvB5Jl89CsZtGIWuTcDM1E/vkVs49/Ml7JJe07l8SPQ=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_flynn_go_shlex",
    importpath = "github.com/flynn/go-shlex",
    sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
    version = "v0.0.0-20150515145356-3f9db97f8568",
)

go_repository(
    name = "com_github_fsnotify_fsnotify",
    importpath = "github.com/fsnotify/fsnotify",
    sum = "h1:IXs+QLmnXW2CcXuY+8Mzv/fWEsPGWxqefPtCP5CnV9I=",
    version = "v1.4.7",
)

go_repository(
    name = "com_github_fsouza_fake_gcs_server",
    importpath = "github.com/fsouza/fake-gcs-server",
    sum = "h1:Un0BXUXrRWYSmYyC1Rqm2e2WJfTPyDy/HGMz31emTi8=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_gchaincl_sqlhooks",
    importpath = "github.com/gchaincl/sqlhooks",
    sum = "h1:c/JqKEQFqiqnV8D7b834AT5gyrOsFvxdmBWJPAgXVz4=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_geertjohan_go_incremental",
    importpath = "github.com/GeertJohan/go.incremental",
    sum = "h1:7AH+pY1XUgQE4Y1HcXYaMqAI0m9yrFqo/jt0CW30vsg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_geertjohan_go_rice",
    importpath = "github.com/GeertJohan/go.rice",
    sum = "h1:KkI6O9uMaQU3VEKaj01ulavtF7o1fWT7+pk/4voiMLQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_getsentry_raven_go",
    importpath = "github.com/getsentry/raven-go",
    sum = "h1:no+xWJRb5ZI7eE8TWgIq1jLulQiIoLG0IfYxv5JYMGs=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_gfleury_go_bitbucket_v1",
    importpath = "github.com/gfleury/go-bitbucket-v1",
    sum = "h1:vtRLrgKfQV38uxXEMCBRPiMcK/IDXcS0a58Y28sfOfA=",
    version = "v0.0.0-20190725203704-66612acc3762",
)

go_repository(
    name = "com_github_ghodss_yaml",
    importpath = "github.com/ghodss/yaml",
    sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_gin_contrib_sse",
    importpath = "github.com/gin-contrib/sse",
    sum = "h1:Y/yl/+YNO8GZSjAhjMsSuLt29uWRFHdHYUb5lYOV9qE=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_gin_gonic_gin",
    importpath = "github.com/gin-gonic/gin",
    sum = "h1:3tMoCCfM7ppqsR0ptz/wi1impNpT7/9wQtMZ8lr1mCQ=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_gitchander_permutation",
    importpath = "github.com/gitchander/permutation",
    sum = "h1:9NC8seTx6/zRmMTAdsHj/uOMi0EGHGQtjyLafBjk77Q=",
    version = "v0.0.0-20181107151852-9e56b92e9909",
)

go_repository(
    name = "com_github_gliderlabs_ssh",
    importpath = "github.com/gliderlabs/ssh",
    sum = "h1:6zsha5zo/TWhRhwqCD3+EarCAgZ2yN28ipRnGPnwkI0=",
    version = "v0.2.2",
)

go_repository(
    name = "com_github_glycerine_go_unsnap_stream",
    importpath = "github.com/glycerine/go-unsnap-stream",
    sum = "h1:FQqoVvjbiUioBBFUL5up+h+GdCa/AnJsL/1bIs/veSI=",
    version = "v0.0.0-20190901134440-81cf024a9e0a",
)

go_repository(
    name = "com_github_glycerine_goconvey",
    importpath = "github.com/glycerine/goconvey",
    sum = "h1:gclg6gY70GLy3PbkQ1AERPfmLMMagS60DKF78eWwLn8=",
    version = "v0.0.0-20190410193231-58a59202ab31",
)

go_repository(
    name = "com_github_go_critic_go_critic",
    importpath = "github.com/go-critic/go-critic",
    sum = "h1:GYXWx7Vr3+zv833u+8IoXbNnQY0AdXsxAgI0kX7xcwA=",
    version = "v0.3.5-0.20190904082202-d79a9f0c64db",
)

go_repository(
    name = "com_github_go_delve_delve",
    importpath = "github.com/go-delve/delve",
    sum = "h1:LI0X4aAE0N7ay8uR48aMJ6GNTxamIpcxi1THjwJjenI=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_go_gl_glfw",
    importpath = "github.com/go-gl/glfw",
    sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
    version = "v0.0.0-20190409004039-e6da0acd62b1",
)

go_repository(
    name = "com_github_go_kit_kit",
    importpath = "github.com/go-kit/kit",
    sum = "h1:wDJmvq38kDhkVxi50ni9ykkdUr1PKgqKOoi01fa0Mdk=",
    version = "v0.9.0",
)

go_repository(
    name = "com_github_go_lintpack_lintpack",
    importpath = "github.com/go-lintpack/lintpack",
    sum = "h1:DI5mA3+eKdWeJ40nU4d6Wc26qmdG8RCi/btYq0TuRN0=",
    version = "v0.5.2",
)

go_repository(
    name = "com_github_go_logfmt_logfmt",
    importpath = "github.com/go-logfmt/logfmt",
    sum = "h1:MP4Eh7ZCb31lleYCFuwm0oe4/YGak+5l1vA2NOE80nA=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_go_ole_go_ole",
    importpath = "github.com/go-ole/go-ole",
    sum = "h1:2lOsA72HgjxAuMlKpFiCbHTvu44PIVkZ5hqm3RSdI/E=",
    version = "v1.2.1",
)

go_repository(
    name = "com_github_go_redsync_redsync",
    importpath = "github.com/go-redsync/redsync",
    sum = "h1:EgfKAu/NI6Y3kZMvyRCT6arWTR+m4IedgdHz0HlIKK0=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_go_sql_driver_mysql",
    importpath = "github.com/go-sql-driver/mysql",
    sum = "h1:g24URVg0OFbNUTx9qqY1IRZ9D9z3iPyi5zKhQZpNwpA=",
    version = "v1.4.1",
)

go_repository(
    name = "com_github_go_stack_stack",
    importpath = "github.com/go-stack/stack",
    sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
    version = "v1.8.0",
)

go_repository(
    name = "com_github_go_toolsmith_astcast",
    importpath = "github.com/go-toolsmith/astcast",
    sum = "h1:JojxlmI6STnFVG9yOImLeGREv8W2ocNUM+iOhR6jE7g=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_astcopy",
    importpath = "github.com/go-toolsmith/astcopy",
    sum = "h1:OMgl1b1MEpjFQ1m5ztEO06rz5CUd3oBv9RF7+DyvdG8=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_astequal",
    importpath = "github.com/go-toolsmith/astequal",
    sum = "h1:4zxD8j3JRFNyLN46lodQuqz3xdKSrur7U/sr0SDS/gQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_astfmt",
    importpath = "github.com/go-toolsmith/astfmt",
    sum = "h1:A0vDDXt+vsvLEdbMFJAUBI/uTbRw1ffOPnxsILnFL6k=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_astinfo",
    importpath = "github.com/go-toolsmith/astinfo",
    sum = "h1:wP6mXeB2V/d1P1K7bZ5vDUO3YqEzcvOREOxZPEu3gVI=",
    version = "v0.0.0-20180906194353-9809ff7efb21",
)

go_repository(
    name = "com_github_go_toolsmith_astp",
    importpath = "github.com/go-toolsmith/astp",
    sum = "h1:alXE75TXgcmupDsMK1fRAy0YUzLzqPVvBKoyWV+KPXg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_pkgload",
    importpath = "github.com/go-toolsmith/pkgload",
    sum = "h1:4DFWWMXVfbcN5So1sBNW9+yeiMqLFGl1wFLTL5R0Tgg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_strparse",
    importpath = "github.com/go-toolsmith/strparse",
    sum = "h1:Vcw78DnpCAKlM20kSbAyO4mPfJn/lyYA4BJUDxe2Jb4=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_go_toolsmith_typep",
    importpath = "github.com/go-toolsmith/typep",
    sum = "h1:zKymWyA1TRYvqYrYDrfEMZULyrhcnGY3x7LDKU2XQaA=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_gobwas_glob",
    importpath = "github.com/gobwas/glob",
    sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
    version = "v0.2.3",
)

go_repository(
    name = "com_github_gocql_gocql",
    importpath = "github.com/gocql/gocql",
    sum = "h1:vF83LI8tAakwEwvWZtrIEx7pOySacl2TOxx6eXk4ePo=",
    version = "v0.0.0-20190301043612-f6df8288f9b4",
)

go_repository(
    name = "com_github_gogo_protobuf",
    importpath = "github.com/gogo/protobuf",
    sum = "h1:G8O7TerXerS4F6sx9OV7/nRfJdnXgHZu/S/7F2SN+UE=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_golang_gddo",
    importpath = "github.com/golang/gddo",
    sum = "h1:xisWqjiKEff2B0KfFYGpCqc3M3zdTz+OHQHRc09FeYk=",
    version = "v0.0.0-20190904175337-72a348e765d2",
)

go_repository(
    name = "com_github_golang_glog",
    importpath = "github.com/golang/glog",
    sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
    version = "v0.0.0-20160126235308-23def4e6c14b",
)

go_repository(
    name = "com_github_golang_groupcache",
    importpath = "github.com/golang/groupcache",
    sum = "h1:55rEp52jU6bkyslZ1+C/7NGfpQsEc6pxGLAGDOctqbw=",
    version = "v0.0.0-20191002201903-404acd9df4cc",
)

go_repository(
    name = "com_github_golang_migrate_migrate_v4",
    importpath = "github.com/golang-migrate/migrate/v4",
    sum = "h1:LDDOHo/q1W5UDj6PbkxdCv7lv9yunyZHXvxuwDkGo3k=",
    version = "v4.6.2",
)

go_repository(
    name = "com_github_golang_mock",
    importpath = "github.com/golang/mock",
    sum = "h1:qGJ6qTW+x6xX/my+8YUVl4WNpX9B7+/l2tRsHGZ7f2s=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    sum = "h1:6nsPYzhq5kReh6QImI3k5qWzO4PEbvbIW2cwSfR/6xs=",
    version = "v1.3.2",
)

go_repository(
    name = "com_github_golang_snappy",
    importpath = "github.com/golang/snappy",
    sum = "h1:Qgr9rKW7uDUkrbSmQeiDsGa8SjGyCOGtuasMWwvp2P4=",
    version = "v0.0.1",
)

go_repository(
    name = "com_github_golangci_check",
    importpath = "github.com/golangci/check",
    sum = "h1:23T5iq8rbUYlhpt5DB4XJkc6BU31uODLD1o1gKvZmD0=",
    version = "v0.0.0-20180506172741-cfe4005ccda2",
)

go_repository(
    name = "com_github_golangci_dupl",
    importpath = "github.com/golangci/dupl",
    sum = "h1:w8hkcTqaFpzKqonE9uMCefW1WDie15eSP/4MssdenaM=",
    version = "v0.0.0-20180902072040-3e9179ac440a",
)

go_repository(
    name = "com_github_golangci_errcheck",
    importpath = "github.com/golangci/errcheck",
    sum = "h1:YYWNAGTKWhKpcLLt7aSj/odlKrSrelQwlovBpDuf19w=",
    version = "v0.0.0-20181223084120-ef45e06d44b6",
)

go_repository(
    name = "com_github_golangci_go_misc",
    importpath = "github.com/golangci/go-misc",
    sum = "h1:9kfjN3AdxcbsZBf8NjltjWihK2QfBBBZuv91cMFfDHw=",
    version = "v0.0.0-20180628070357-927a3d87b613",
)

go_repository(
    name = "com_github_golangci_goconst",
    importpath = "github.com/golangci/goconst",
    sum = "h1:pe9JHs3cHHDQgOFXJJdYkK6fLz2PWyYtP4hthoCMvs8=",
    version = "v0.0.0-20180610141641-041c5f2b40f3",
)

go_repository(
    name = "com_github_golangci_gocyclo",
    importpath = "github.com/golangci/gocyclo",
    sum = "h1:pXTK/gkVNs7Zyy7WKgLXmpQ5bHTrq5GDsp8R9Qs67g0=",
    version = "v0.0.0-20180528144436-0a533e8fa43d",
)

go_repository(
    name = "com_github_golangci_gofmt",
    importpath = "github.com/golangci/gofmt",
    sum = "h1:iR3fYXUjHCR97qWS8ch1y9zPNsgXThGwjKPrYfqMPks=",
    version = "v0.0.0-20190930125516-244bba706f1a",
)

go_repository(
    name = "com_github_golangci_golangci_lint",
    importpath = "github.com/golangci/golangci-lint",
    sum = "h1:bokjFHRgut7TT9YiKSVWVscWfZNs1fQgSz94geGPHJw=",
    version = "v1.20.0",
)

go_repository(
    name = "com_github_golangci_ineffassign",
    importpath = "github.com/golangci/ineffassign",
    sum = "h1:gLLhTLMk2/SutryVJ6D4VZCU3CUqr8YloG7FPIBWFpI=",
    version = "v0.0.0-20190609212857-42439a7714cc",
)

go_repository(
    name = "com_github_golangci_lint_1",
    importpath = "github.com/golangci/lint-1",
    sum = "h1:664ewjIQUXDvinFMbAsoH2V2Yvaro/X8BoYpIMTWGXI=",
    version = "v0.0.0-20190930103755-fad67e08aa89",
)

go_repository(
    name = "com_github_golangci_maligned",
    importpath = "github.com/golangci/maligned",
    sum = "h1:kNY3/svz5T29MYHubXix4aDDuE3RWHkPvopM/EDv/MA=",
    version = "v0.0.0-20180506175553-b1d89398deca",
)

go_repository(
    name = "com_github_golangci_misspell",
    importpath = "github.com/golangci/misspell",
    sum = "h1:EL/O5HGrF7Jaq0yNhBLucz9hTuRzj2LdwGBOaENgxIk=",
    version = "v0.0.0-20180809174111-950f5d19e770",
)

go_repository(
    name = "com_github_golangci_prealloc",
    importpath = "github.com/golangci/prealloc",
    sum = "h1:leSNB7iYzLYSSx3J/s5sVf4Drkc68W2wm4Ixh/mr0us=",
    version = "v0.0.0-20180630174525-215b22d4de21",
)

go_repository(
    name = "com_github_golangci_revgrep",
    importpath = "github.com/golangci/revgrep",
    sum = "h1:XQKc8IYQOeRwVs36tDrEmTgDgP88d5iEURwpmtiAlOM=",
    version = "v0.0.0-20180812185044-276a5c0a1039",
)

go_repository(
    name = "com_github_golangci_unconvert",
    importpath = "github.com/golangci/unconvert",
    sum = "h1:zwtduBRr5SSWhqsYNgcuWO2kFlpdOZbP0+yRjmvPGys=",
    version = "v0.0.0-20180507085042-28b1c447d1f4",
)

go_repository(
    name = "com_github_golangplus_bytes",
    importpath = "github.com/golangplus/bytes",
    sum = "h1:7xqw01UYS+KCI25bMrPxwNYkSns2Db1ziQPpVq99FpE=",
    version = "v0.0.0-20160111154220-45c989fe5450",
)

go_repository(
    name = "com_github_golangplus_fmt",
    importpath = "github.com/golangplus/fmt",
    sum = "h1:f5gsjBiF9tRRVomCvrkGMMWI8W1f2OBFar2c5oakAP0=",
    version = "v0.0.0-20150411045040-2a5d6d7d2995",
)

go_repository(
    name = "com_github_golangplus_testing",
    importpath = "github.com/golangplus/testing",
    sum = "h1:KhcknUwkWHKZPbFy2P7jH5LKJ3La+0ZeknkkmrSgqb0=",
    version = "v0.0.0-20180327235837-af21d9c3145e",
)

go_repository(
    name = "com_github_gomodule_oauth1",
    importpath = "github.com/gomodule/oauth1",
    sum = "h1:NlNEdePx7QY9Z4rds4EIe1dvUT8Ao1PZgLS80S5YTbU=",
    version = "v0.0.0-20181215000758-9a59ed3b0a84",
)

go_repository(
    name = "com_github_gomodule_redigo",
    importpath = "github.com/gomodule/redigo",
    sum = "h1:K/R+8tc58AaqLkqG2Ol3Qk+DR/TlNuhuh457pBFPtt0=",
    version = "v2.0.0+incompatible",
)

go_repository(
    name = "com_github_google_btree",
    importpath = "github.com/google/btree",
    sum = "h1:0udJVsspx3VBr5FwtLhQQtuAsVc79tTq0ocGIPAU6qo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_go_cmp",
    importpath = "github.com/google/go-cmp",
    sum = "h1:Xye71clBPdm5HgqGwUkwhbynsUJZhDbS20FvLhQ2izg=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_google_go_github",
    importpath = "github.com/google/go-github",
    sum = "h1:N0LgJ1j65A7kfXrZnUDaYCs/Sf4rEjNlfyDHW9dolSY=",
    version = "v17.0.0+incompatible",
)

go_repository(
    name = "com_github_google_go_github_v27",
    importpath = "github.com/google/go-github/v27",
    sum = "h1:oiOZuBmGHvrGM1X9uNUAUlLgp5r1UUO/M/KnbHnLRlQ=",
    version = "v27.0.6",
)

go_repository(
    name = "com_github_google_go_github_v28",
    importpath = "github.com/google/go-github/v28",
    sum = "h1:kORf5ekX5qwXO2mGzXXOjMe/g6ap8ahVe0sBEulhSxo=",
    version = "v28.1.1",
)

go_repository(
    name = "com_github_google_go_querystring",
    importpath = "github.com/google/go-querystring",
    sum = "h1:Xkwi/a1rcvNg1PPYe5vI8GbeBY/jrVuDX5ASuANWTrk=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_gofuzz",
    importpath = "github.com/google/gofuzz",
    sum = "h1:A8PeW59pxE9IoFRqBp37U+mSNaQoZ46F1f0f863XSXw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_google_martian",
    importpath = "github.com/google/martian",
    sum = "h1:/CP5g8u/VJHijgedC/Legn3BAbAaWPgecwXBIDzw5no=",
    version = "v2.1.0+incompatible",
)

go_repository(
    name = "com_github_google_pprof",
    importpath = "github.com/google/pprof",
    sum = "h1:Jnx61latede7zDD3DiiP4gmNz33uK0U5HDUaF0a/HVQ=",
    version = "v0.0.0-20190515194954-54271f7e092f",
)

go_repository(
    name = "com_github_google_renameio",
    importpath = "github.com/google/renameio",
    sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_google_slothfs",
    importpath = "github.com/google/slothfs",
    sum = "h1:iBLrJ79cF90CZmpskySqhPvzrWr9njBYEsOZubXLZlc=",
    version = "v0.0.0-20190417171004-6b42407d9230",
)

go_repository(
    name = "com_github_google_uuid",
    importpath = "github.com/google/uuid",
    sum = "h1:Gkbcsh/GbpXz7lPftLA3P6TYMwjCLYm83jiFQZF/3gY=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_google_zoekt",
    importpath = "github.com/google/zoekt",
    replace = "github.com/sourcegraph/zoekt",
    sum = "h1:8ggy5qE3+5d3521NPAQ+0AjGuwSJZSgUvnMNQbpn3Yc=",
    version = "v0.0.0-20191018064203-90218876a705",
)

go_repository(
    name = "com_github_googleapis_gax_go_v2",
    importpath = "github.com/googleapis/gax-go/v2",
    sum = "h1:sjZBwGj9Jlw33ImPtvFviGYvseOtDM7hkSKB7+Tv3SM=",
    version = "v2.0.5",
)

go_repository(
    name = "com_github_gopherjs_gopherjs",
    importpath = "github.com/gopherjs/gopherjs",
    sum = "h1:TyqzGm2z1h3AGhjOoRYyeLcW4WlW81MDQkWa+rx/000=",
    version = "v0.0.0-20190915194858-d3ddacdb130f",
)

go_repository(
    name = "com_github_gorilla_context",
    importpath = "github.com/gorilla/context",
    sum = "h1:AWwleXJkX/nhcU9bZSnZoi3h/qGYqQAGhq6zZe/aQW8=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_csrf",
    importpath = "github.com/gorilla/csrf",
    sum = "h1:wua1OxOTarfqtUVfiSvzs2zTr3qV57cXVGclVETJXXc=",
    version = "v1.6.1",
)

go_repository(
    name = "com_github_gorilla_handlers",
    importpath = "github.com/gorilla/handlers",
    sum = "h1:0QniY0USkHQ1RGCLfKxeNHK9bkDHGRYGNDFBCS+YARg=",
    version = "v1.4.2",
)

go_repository(
    name = "com_github_gorilla_mux",
    importpath = "github.com/gorilla/mux",
    sum = "h1:gnP5JzjVOuiZD07fKKToCAOjS0yOpj/qPETTXCCS6hw=",
    version = "v1.7.3",
)

go_repository(
    name = "com_github_gorilla_schema",
    importpath = "github.com/gorilla/schema",
    sum = "h1:CamqUDOFUBqzrvxuz2vEwo8+SUdwsluFh7IlzJh30LY=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_gorilla_securecookie",
    importpath = "github.com/gorilla/securecookie",
    sum = "h1:miw7JPhV+b/lAHSXz4qd/nN9jRiAFV5FwjeKyCS8BvQ=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_gorilla_sessions",
    importpath = "github.com/gorilla/sessions",
    sum = "h1:S7P+1Hm5V/AT9cjEcUD5uDaQSX0OE577aCXgoaKpYbQ=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_gorilla_websocket",
    importpath = "github.com/gorilla/websocket",
    sum = "h1:WDFjx/TMzVgy9VdMMQi2K2Emtwi2QcUQsztZ/zLaH/Q=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_gostaticanalysis_analysisutil",
    importpath = "github.com/gostaticanalysis/analysisutil",
    sum = "h1:iwp+5/UAyzQSFgQ4uR2sni99sJ8Eo9DEacKWM5pekIg=",
    version = "v0.0.3",
)

go_repository(
    name = "com_github_goware_urlx",
    importpath = "github.com/goware/urlx",
    sum = "h1:BbvKl8oiXtJAzOzMqAQ0GfIhf96fKeNEZfm9ocNSUBI=",
    version = "v0.3.1",
)

go_repository(
    name = "com_github_graph_gophers_graphql_go",
    importpath = "github.com/graph-gophers/graphql-go",
    sum = "h1:HYKt6iF2RNYqWWA8e58qT4DLo5COH0YJjUzpCBMkNO0=",
    version = "v0.0.0-20190917030536-38a077bc812d",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    importpath = "github.com/gregjones/httpcache",
    sum = "h1:+ngKgrYPPJrOjhax5N+uePQ0Fh1Z7PheYoUI/0nzkPA=",
    version = "v0.0.0-20190611155906-901d90724c79",
)

go_repository(
    name = "com_github_grpc_ecosystem_go_grpc_middleware",
    importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
    sum = "h1:Iju5GlWwrvL6UBg4zJJt3btmonfrMlCDdsejg4CZE7c=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_grpc_ecosystem_go_grpc_prometheus",
    importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
    sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_grpc_ecosystem_grpc_gateway",
    importpath = "github.com/grpc-ecosystem/grpc-gateway",
    sum = "h1:bM6ZAFZmc/wPFaRDi0d5L7hGEZEx/2u+Tmr2evNHDiI=",
    version = "v1.9.0",
)

go_repository(
    name = "com_github_hailocab_go_hostpool",
    importpath = "github.com/hailocab/go-hostpool",
    sum = "h1:5upAirOpQc1Q53c0bnx2ufif5kANL7bfZWcc6VJWJd8=",
    version = "v0.0.0-20160125115350-e80d13ce29ed",
)

go_repository(
    name = "com_github_hashicorp_errwrap",
    importpath = "github.com/hashicorp/errwrap",
    sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_cleanhttp",
    importpath = "github.com/hashicorp/go-cleanhttp",
    sum = "h1:dH3aiDG9Jvb5r5+bYHsikaOUIpcM0xvgMXVoDkXMzJM=",
    version = "v0.5.1",
)

go_repository(
    name = "com_github_hashicorp_go_hclog",
    importpath = "github.com/hashicorp/go-hclog",
    sum = "h1:CG6TE5H9/JXsFWJCfoIVpKFIkFe6ysEuHirp4DxCsHI=",
    version = "v0.9.2",
)

go_repository(
    name = "com_github_hashicorp_go_multierror",
    importpath = "github.com/hashicorp/go-multierror",
    sum = "h1:iVjPR7a6H0tWELX5NxNe7bYopibicUzc7uPribsnS6o=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_hashicorp_go_retryablehttp",
    importpath = "github.com/hashicorp/go-retryablehttp",
    sum = "h1:bHM2aVXwBtBJWxHtkSrWuI4umABCUczs52eiUS9nSiw=",
    version = "v0.6.2",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    importpath = "github.com/hashicorp/golang-lru",
    sum = "h1:0hERBMJE1eitiLkihrMvRVBYAkpHzc/J3QdDN+dAcgU=",
    version = "v0.5.1",
)

go_repository(
    name = "com_github_hashicorp_hcl",
    importpath = "github.com/hashicorp/hcl",
    sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_honeycombio_libhoney_go",
    importpath = "github.com/honeycombio/libhoney-go",
    sum = "h1:BUMNgFFoA7Rq+7+kwwxQInsb7LO2/we4JAxJvZ/9aBE=",
    version = "v1.12.1",
)

go_repository(
    name = "com_github_hpcloud_tail",
    importpath = "github.com/hpcloud/tail",
    sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    importpath = "github.com/inconshreveable/mousetrap",
    sum = "h1:Z8tu5sraLXCXIcARxBp/8cbvlwVa7Z1NHg9XEKhtSvM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_jackc_fake",
    importpath = "github.com/jackc/fake",
    sum = "h1:vr3AYkKovP8uR8AvSGGUK1IDqRa5lAAvEkZG1LKaCRc=",
    version = "v0.0.0-20150926172116-812a484cc733",
)

go_repository(
    name = "com_github_jackc_pgx",
    importpath = "github.com/jackc/pgx",
    sum = "h1:0Vihzu20St42/UDsvZGdNE6jak7oi/UOeMzwMPHkgFY=",
    version = "v3.2.0+incompatible",
)

go_repository(
    name = "com_github_jbenet_go_context",
    importpath = "github.com/jbenet/go-context",
    sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
    version = "v0.0.0-20150711004518-d14ea06fba99",
)

go_repository(
    name = "com_github_jessevdk_go_flags",
    importpath = "github.com/jessevdk/go-flags",
    sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    importpath = "github.com/jmespath/go-jmespath",
    sum = "h1:pmfjZENx5imkbgOkpRUYLnmbU7UEFbjtDA2hxJ1ichM=",
    version = "v0.0.0-20180206201540-c2b33e8439af",
)

go_repository(
    name = "com_github_jmoiron_sqlx",
    importpath = "github.com/jmoiron/sqlx",
    sum = "h1:41Ip0zITnmWNR/vHV+S4m+VoUivnWY5E4OJfLZjCJMA=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_joho_godotenv",
    importpath = "github.com/joho/godotenv",
    sum = "h1:Zjp+RcGpHhGlrMbJzXTrZZPrWj+1vfm90La1wgB6Bhc=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_jonboulle_clockwork",
    importpath = "github.com/jonboulle/clockwork",
    sum = "h1:VKV+ZcuP6l3yW9doeqz6ziZGgcynBVQO+obU0+0hcPo=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_json_iterator_go",
    importpath = "github.com/json-iterator/go",
    sum = "h1:KfgG9LzI+pYjr4xvmz/5H4FXjokeP+rlHLhv3iH62Fo=",
    version = "v1.1.7",
)

go_repository(
    name = "com_github_jstemmer_go_junit_report",
    importpath = "github.com/jstemmer/go-junit-report",
    sum = "h1:6QPYqodiu3GuPL+7mfx+NwDdp2eTkp9IfEUpgAwUN0o=",
    version = "v0.9.1",
)

go_repository(
    name = "com_github_jtolds_gls",
    importpath = "github.com/jtolds/gls",
    sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
    version = "v4.20.0+incompatible",
)

go_repository(
    name = "com_github_julienschmidt_httprouter",
    importpath = "github.com/julienschmidt/httprouter",
    sum = "h1:TDTW5Yz1mjftljbcKqRcrYhd4XeOoI98t+9HbQbYf7g=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_kardianos_osext",
    importpath = "github.com/kardianos/osext",
    sum = "h1:iQTw/8FWTuc7uiaSepXwyf3o52HaUYcV+Tu66S3F5GA=",
    version = "v0.0.0-20190222173326-2bc1f35cddc0",
)

go_repository(
    name = "com_github_karlseguin_expect",
    importpath = "github.com/karlseguin/expect",
    sum = "h1:z4wy4npwwHSWKjGWH85WNJO42VQhovxTCZDSzhjo8hY=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_karlseguin_typed",
    importpath = "github.com/karlseguin/typed",
    sum = "h1:R8JGxdS0PNWWPKnHg4zh/n8NDFTEa6oIUcxjrMO4c6w=",
    version = "v1.1.7",
)

go_repository(
    name = "com_github_karrick_godirwalk",
    importpath = "github.com/karrick/godirwalk",
    sum = "h1:nkS4xxsjiZMvVlazd0mFyiwD4BR9f3m6LXGhM2TUx3Y=",
    version = "v1.12.0",
)

go_repository(
    name = "com_github_karrick_tparse",
    importpath = "github.com/karrick/tparse",
    sum = "h1:+cW306qKAzrASC5XieHkgN7/vPaGKIuK62Q7nI7DIRc=",
    version = "v2.4.2+incompatible",
)

go_repository(
    name = "com_github_keegancsmith_rpc",
    importpath = "github.com/keegancsmith/rpc",
    sum = "h1:bXVRk3EzbtrEegTGKxNTc+St1lR7t/Z1PAO8misBnCc=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_keegancsmith_sqlf",
    importpath = "github.com/keegancsmith/sqlf",
    sum = "h1:lQ2YFRqdEyjDZ/OHsh417gs/2Fz+Jc7qEM7kRrc3ofw=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_keegancsmith_tmpfriend",
    importpath = "github.com/keegancsmith/tmpfriend",
    sum = "h1:xa9SZfAid/jlS3kjwAvVDQFpe6t8SiS0Vl/H51BZYww=",
    version = "v0.0.0-20180423180255-86e88902a513",
)

go_repository(
    name = "com_github_kevinburke_go_bindata",
    importpath = "github.com/kevinburke/go-bindata",
    sum = "h1:b7htuEw9tZW6lYhES775BTH/KRsqKzmU2RockgdunKw=",
    version = "v3.14.0+incompatible",
)

go_repository(
    name = "com_github_kevinburke_ssh_config",
    importpath = "github.com/kevinburke/ssh_config",
    sum = "h1:Coekwdh0v2wtGp9Gmz1Ze3eVRAWJMLokvN3QjdzCHLY=",
    version = "v0.0.0-20190725054713-01f96b0aa0cd",
)

go_repository(
    name = "com_github_kisielk_errcheck",
    importpath = "github.com/kisielk/errcheck",
    sum = "h1:reN85Pxc5larApoH1keMBiu2GWtPqXQ1nc9gx+jOU+E=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_kisielk_gotool",
    importpath = "github.com/kisielk/gotool",
    sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_klauspost_compress",
    importpath = "github.com/klauspost/compress",
    sum = "h1:970MQcQdxX7hfgc/aqmB4a3grW0ivUVV6i1TLkP8CiE=",
    version = "v1.8.6",
)

go_repository(
    name = "com_github_klauspost_cpuid",
    importpath = "github.com/klauspost/cpuid",
    sum = "h1:vJi+O/nMdFt0vqm8NZBI6wzALWdA2X+egi0ogNyrC/w=",
    version = "v1.2.1",
)

go_repository(
    name = "com_github_konsorten_go_windows_terminal_sequences",
    importpath = "github.com/konsorten/go-windows-terminal-sequences",
    sum = "h1:DB17ag19krx9CFsz4o3enTrPXyIXCl+2iCXH/aMAp9s=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_kr_logfmt",
    importpath = "github.com/kr/logfmt",
    sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
    version = "v0.0.0-20140226030751-b84e30acd515",
)

go_repository(
    name = "com_github_kr_pretty",
    importpath = "github.com/kr/pretty",
    sum = "h1:L/CwN0zerZDmRFUapSPitk6f+Q3+0za1rQkzVuMiMFI=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_kr_pty",
    importpath = "github.com/kr/pty",
    sum = "h1:AkaSdXYQOWeaO3neb8EM634ahkXXe3jYbVh/F9lq+GI=",
    version = "v1.1.8",
)

go_repository(
    name = "com_github_kr_text",
    importpath = "github.com/kr/text",
    sum = "h1:45sCR5RtlFHMR4UwH9sdQ5TC8v0qDQCHnXt+kaKSTVE=",
    version = "v0.1.0",
)

go_repository(
    name = "com_github_kshvakov_clickhouse",
    importpath = "github.com/kshvakov/clickhouse",
    sum = "h1:PDTYk9VYgbjPAWry3AoDREeMgOVUFij6bh6IjlloHL0=",
    version = "v1.3.5",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    importpath = "github.com/kylelemons/godebug",
    sum = "h1:RPNrshWIDI6G2gRW9EHilWtl7Z6Sb1BR0xunSBf0SNc=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_leanovate_gopter",
    importpath = "github.com/leanovate/gopter",
    sum = "h1:U4YLBggDFhJdqQsG4Na2zX7joVTky9vHaj/AGEwSuXU=",
    version = "v0.2.4",
)

go_repository(
    name = "com_github_lib_pq",
    importpath = "github.com/lib/pq",
    sum = "h1:LXpIM/LZ5xGFhOpXAQUIMM1HdyqzVYM13zNdjCEEcA0=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_lightstep_lightstep_tracer_common_golang_gogo",
    importpath = "github.com/lightstep/lightstep-tracer-common/golang/gogo",
    sum = "h1:143Bb8f8DuGWck/xpNUOckBVYfFbBTnLevfRZ1aVVqo=",
    version = "v0.0.0-20190605223551-bc2310a04743",
)

go_repository(
    name = "com_github_lightstep_lightstep_tracer_go",
    importpath = "github.com/lightstep/lightstep-tracer-go",
    sum = "h1:ZzGERV0m7ROSKtwODRStAve3kQaVeCLD5JGNcQP3+88=",
    version = "v0.17.0",
)

go_repository(
    name = "com_github_logrusorgru_aurora",
    importpath = "github.com/logrusorgru/aurora",
    sum = "h1:9MlwzLdW7QSDrhDjFlsEYmxpFyIoXmYRon3dt0io31k=",
    version = "v0.0.0-20181002194514-a7b3b318ed4e",
)

go_repository(
    name = "com_github_magiconair_properties",
    importpath = "github.com/magiconair/properties",
    sum = "h1:ZC2Vc7/ZFkGmsVC9KvOjumD+G5lXy2RtTKyzRKO2BQ4=",
    version = "v1.8.1",
)

go_repository(
    name = "com_github_martinlindhe_base36",
    importpath = "github.com/martinlindhe/base36",
    sum = "h1:CVsnY46BCLkX9XOhALJ/S7yb9ayc4eqjXSXO3tyB66A=",
    version = "v0.0.0-20190418230009-7c6542dfbb41",
)

go_repository(
    name = "com_github_matoous_godox",
    importpath = "github.com/matoous/godox",
    sum = "h1:RHba4YImhrUVQDHUCe2BNSOz4tVy2yGyXhvYDvxGgeE=",
    version = "v0.0.0-20190911065817-5d6d842e92eb",
)

go_repository(
    name = "com_github_mattn_go_colorable",
    importpath = "github.com/mattn/go-colorable",
    sum = "h1:snbPLB8fVfU9iwbbo30TPtbLRzwWu6aJS6Xh4eaaviA=",
    version = "v0.1.4",
)

go_repository(
    name = "com_github_mattn_go_isatty",
    importpath = "github.com/mattn/go-isatty",
    sum = "h1:qxFzApOv4WsAL965uUPIsXzAKCZxN2p9UqdhFS4ZW10=",
    version = "v0.0.10",
)

go_repository(
    name = "com_github_mattn_go_runewidth",
    importpath = "github.com/mattn/go-runewidth",
    sum = "h1:2BvfKmzob6Bmd4YsL0zygOqfdFnK7GR4QL06Do4/p7Y=",
    version = "v0.0.4",
)

go_repository(
    name = "com_github_mattn_go_sqlite3",
    importpath = "github.com/mattn/go-sqlite3",
    sum = "h1:LDdKkqtYlom37fkvqs8rMPFKAMe8+SgjbwZ6ex1/A/Q=",
    version = "v1.11.0",
)

go_repository(
    name = "com_github_mattn_goreman",
    importpath = "github.com/mattn/goreman",
    replace = "github.com/sourcegraph/goreman",
    sum = "h1:FxF0pen6r1WOqbm4kVDR3JV078HfPz65inWcMh1aVQI=",
    version = "v0.1.2-0.20180928223752-6e9a2beb830d",
)

go_repository(
    name = "com_github_mattn_goveralls",
    importpath = "github.com/mattn/goveralls",
    sum = "h1:7eJB6EqsPhRVxvwEXGnqdO2sJI0PTsrWoTMXEk9/OQc=",
    version = "v0.0.2",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
    sum = "h1:4hp9jkHxhMHkqkrB3Ix0jegS5sx/RkqARlsWZ6pIwiU=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_mcuadros_go_version",
    importpath = "github.com/mcuadros/go-version",
    sum = "h1:YocNLcTBdEdvY3iDK6jfWXvEaM5OCKkjxPKoJRdB3Gg=",
    version = "v0.0.0-20190830083331-035f6764e8d2",
)

go_repository(
    name = "com_github_microcosm_cc_bluemonday",
    importpath = "github.com/microcosm-cc/bluemonday",
    sum = "h1:5lPfLTTAvAbtS0VqT+94yOtFnGfUWYyx0+iToC3Os3s=",
    version = "v1.0.2",
)

go_repository(
    name = "com_github_microsoft_go_winio",
    importpath = "github.com/Microsoft/go-winio",
    sum = "h1:+hMXMk01us9KgxGb7ftKQt2Xpf5hH/yky+TDA+qxleU=",
    version = "v0.4.14",
)

go_repository(
    name = "com_github_mitchellh_go_homedir",
    importpath = "github.com/mitchellh/go-homedir",
    sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_mitchellh_go_ps",
    importpath = "github.com/mitchellh/go-ps",
    sum = "h1:9+ke9YJ9KGWw5ANXK6ozjoK47uI3uNbXv4YVINBnGm8=",
    version = "v0.0.0-20190716172923-621e5597135b",
)

go_repository(
    name = "com_github_mitchellh_go_wordwrap",
    importpath = "github.com/mitchellh/go-wordwrap",
    sum = "h1:6GlHJ/LTGMrIJbwgdqdl2eEH8o+Exx/0m8ir9Gns0u4=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_mitchellh_mapstructure",
    importpath = "github.com/mitchellh/mapstructure",
    sum = "h1:fmNYVwqnSfB9mZU6OS2O6GsXM+wcskZDuKQzvN1EDeE=",
    version = "v1.1.2",
)

go_repository(
    name = "com_github_modern_go_concurrent",
    importpath = "github.com/modern-go/concurrent",
    sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
    version = "v0.0.0-20180306012644-bacd9c7ef1dd",
)

go_repository(
    name = "com_github_modern_go_reflect2",
    importpath = "github.com/modern-go/reflect2",
    sum = "h1:9f412s+6RmYXLWZSEzVVgPGK7C2PphHj5RJrvfx9AWI=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_morikuni_aec",
    importpath = "github.com/morikuni/aec",
    sum = "h1:nP9CBfwrvYnBRgY6qfDQkygYDmYwOilePFkwzv4dU8A=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_mozilla_tls_observatory",
    importpath = "github.com/mozilla/tls-observatory",
    sum = "h1:Av0AX0PnAlPZ3AY2rQUobGFaZfE4KHVRdKWIEPvsCWY=",
    version = "v0.0.0-20190404164649-a3c1b6cfecfd",
)

go_repository(
    name = "com_github_mozillazg_go_slugify",
    importpath = "github.com/mozillazg/go-slugify",
    sum = "h1:SIhqDlnJWZH8OdiTmQgeXR28AOnypmAXPeOTcG7b9lk=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_mozillazg_go_unidecode",
    importpath = "github.com/mozillazg/go-unidecode",
    sum = "h1:uiRy1s4TUqLbcROUrnCN/V85Jlli2AmDF6EeAXOeMHE=",
    version = "v0.1.1",
)

go_repository(
    name = "com_github_mschoch_smat",
    importpath = "github.com/mschoch/smat",
    sum = "h1:VeRdUYdCw49yizlSbMEn2SZ+gT+3IUKx8BqxyQdz+BY=",
    version = "v0.0.0-20160514031455-90eadee771ae",
)

go_repository(
    name = "com_github_mwitkow_go_conntrack",
    importpath = "github.com/mwitkow/go-conntrack",
    sum = "h1:F9x/1yl3T2AeKLr2AMdilSD8+f9bvMnNN8VS5iDtovc=",
    version = "v0.0.0-20161129095857-cc309e4a2223",
)

go_repository(
    name = "com_github_mxk_go_flowrate",
    importpath = "github.com/mxk/go-flowrate",
    sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
    version = "v0.0.0-20140419014527-cca7078d478f",
)

go_repository(
    name = "com_github_nakagami_firebirdsql",
    importpath = "github.com/nakagami/firebirdsql",
    sum = "h1:P48LjvUQpTReR3TQRbxSeSBsMXzfK0uol7eRcr7VBYQ=",
    version = "v0.0.0-20190310045651-3c02a58cfed8",
)

go_repository(
    name = "com_github_nbutton23_zxcvbn_go",
    importpath = "github.com/nbutton23/zxcvbn-go",
    sum = "h1:AREM5mwr4u1ORQBMvzfzBgpsctsbQikCVpvC+tX285E=",
    version = "v0.0.0-20180912185939-ae427f1e4c1d",
)

go_repository(
    name = "com_github_neelance_parallel",
    importpath = "github.com/neelance/parallel",
    sum = "h1:NZOii9TDGRAfCS5VM16XnF4K7afoLQmIiZX8EkKnxtE=",
    version = "v0.0.0-20160708114440-4de9ce63d14c",
)

go_repository(
    name = "com_github_nkovacs_streamquote",
    importpath = "github.com/nkovacs/streamquote",
    sum = "h1:E2B8qYyeSgv5MXpmzZXRNp8IAQ4vjxIjhpAf5hv/tAg=",
    version = "v0.0.0-20170412213628-49af9bddb229",
)

go_repository(
    name = "com_github_nvveen_gotty",
    importpath = "github.com/Nvveen/Gotty",
    sum = "h1:TngWCqHvy9oXAN6lEVMRuU21PR1EtLVZJmdB18Gu3Rw=",
    version = "v0.0.0-20120604004816-cd527374f1e5",
)

go_repository(
    name = "com_github_nytimes_gziphandler",
    importpath = "github.com/NYTimes/gziphandler",
    sum = "h1:ZUDjpQae29j0ryrS0u/B8HZfJBtBQHjqw2rQ2cqUQ3I=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_oklog_ulid",
    importpath = "github.com/oklog/ulid",
    sum = "h1:EGfNDEx6MqHz8B3uNV6QAib1UR2Lm97sHi3ocA6ESJ4=",
    version = "v1.3.1",
)

go_repository(
    name = "com_github_oneofone_xxhash",
    importpath = "github.com/OneOfOne/xxhash",
    sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
    version = "v1.2.2",
)

go_repository(
    name = "com_github_onsi_ginkgo",
    importpath = "github.com/onsi/ginkgo",
    sum = "h1:uqH7bpe+ERSiDa34FDOF7RikN6RzXgduUF8yarlZp94=",
    version = "v1.10.2",
)

go_repository(
    name = "com_github_onsi_gomega",
    importpath = "github.com/onsi/gomega",
    sum = "h1:XPnZz8VVBHjVsy1vzJmRwIcSwiUO+JFfrv/xGiigmME=",
    version = "v1.7.0",
)

go_repository(
    name = "com_github_opencontainers_go_digest",
    importpath = "github.com/opencontainers/go-digest",
    sum = "h1:WzifXhOVOEOuFYOJAW6aQqW0TooG2iki3E3Ii+WN7gQ=",
    version = "v1.0.0-rc1",
)

go_repository(
    name = "com_github_opencontainers_image_spec",
    importpath = "github.com/opencontainers/image-spec",
    sum = "h1:JMemWkRwHx4Zj+fVxWoMCFm/8sYGGrUVojFA6h/TRcI=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_openpeedeep_depguard",
    importpath = "github.com/OpenPeeDeeP/depguard",
    sum = "h1:VlW4R6jmBIv3/u1JNlawEvJMM4J+dPORPaZasQee8Us=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_opentracing_contrib_go_stdlib",
    importpath = "github.com/opentracing-contrib/go-stdlib",
    sum = "h1:QsgXACQhd9QJhEmRumbsMQQvBtmdS0mafoVEBplWXEg=",
    version = "v0.0.0-20190519235532-cf7a6c988dc9",
)

go_repository(
    name = "com_github_opentracing_opentracing_go",
    importpath = "github.com/opentracing/opentracing-go",
    sum = "h1:pWlfV3Bxv7k65HYwkikxat0+s3pV4bsqf19k25Ur8rU=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_openzipkin_zipkin_go",
    importpath = "github.com/openzipkin/zipkin-go",
    sum = "h1:yXiysv1CSK7Q5yjGy1710zZGnsbMUIjluWBxtLXHPBo=",
    version = "v0.1.6",
)

go_repository(
    name = "com_github_pelletier_go_buffruneio",
    importpath = "github.com/pelletier/go-buffruneio",
    sum = "h1:U4t4R6YkofJ5xHm3dJzuRpPZ0mr5MMCoAWooScCR7aA=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_pelletier_go_toml",
    importpath = "github.com/pelletier/go-toml",
    sum = "h1:5BakdOZdtKJ1FFk6QdL8iSGrMWsXgchNJcrnarjbmJQ=",
    version = "v1.5.0",
)

go_repository(
    name = "com_github_peterbourgon_ff",
    importpath = "github.com/peterbourgon/ff",
    sum = "h1:xax7BA50iaQoo6EnC7EhJlGsioTuo/e2FhMGq44iOHE=",
    version = "v1.6.1-0.20190916204019-6cd704ec2eeb",
)

go_repository(
    name = "com_github_peterh_liner",
    importpath = "github.com/peterh/liner",
    sum = "h1:f+aAedNJA6uk7+6rXsYBnhdo4Xux7ESLe+kcuVUF5os=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_peterhellberg_link",
    importpath = "github.com/peterhellberg/link",
    sum = "h1:mUWkiegowUXEcmlb+ybF75Q/8D2Y0BjZtR8cxoKhaQo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_philhofer_fwd",
    importpath = "github.com/philhofer/fwd",
    sum = "h1:UbZqGr5Y38ApvM/V/jEljVxwocdweyH+vmYvRPBnbqQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_pierrec_lz4",
    importpath = "github.com/pierrec/lz4",
    sum = "h1:2xWsjqPFWcplujydGg4WmhC/6fZqK42wMM8aXeqhl0I=",
    version = "v2.0.5+incompatible",
)

go_repository(
    name = "com_github_pkg_errors",
    importpath = "github.com/pkg/errors",
    sum = "h1:iURUrRGxPUNPdy5/HRSm+Yj6okJ6UtLINN0Q9M4+h3I=",
    version = "v0.8.1",
)

go_repository(
    name = "com_github_pkg_profile",
    importpath = "github.com/pkg/profile",
    sum = "h1:OQIvuDgm00gWVWGTf4m4mCt6W1/0YqU7Ntg0mySWgaI=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    importpath = "github.com/pmezard/go-difflib",
    sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_pquerna_cachecontrol",
    importpath = "github.com/pquerna/cachecontrol",
    sum = "h1:J9b7z+QKAmPf4YLrFg6oQUotqHQeUNWwkvo7jZp1GLU=",
    version = "v0.0.0-20180517163645-1555304b9b35",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    importpath = "github.com/prometheus/client_golang",
    sum = "h1:BQ53HtBmfOitExawJ6LokA4x8ov/z0SYYb0+HxJfRI8=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_prometheus_client_model",
    importpath = "github.com/prometheus/client_model",
    sum = "h1:gQz4mCbXsO+nc9n1hCxHcGA3Zx3Eo+UHZoInFGUIXNM=",
    version = "v0.0.0-20190812154241-14fe0d1b01d4",
)

go_repository(
    name = "com_github_prometheus_common",
    importpath = "github.com/prometheus/common",
    sum = "h1:L+1lyG48J1zAQXA3RBX/nG/B3gjlHq0zTt2tlbJLyCY=",
    version = "v0.7.0",
)

go_repository(
    name = "com_github_prometheus_procfs",
    importpath = "github.com/prometheus/procfs",
    sum = "h1:3+auTFlqw+ZaQYJARz6ArODtkaIwtvBTx3N2NehQlL8=",
    version = "v0.0.5",
)

go_repository(
    name = "com_github_prometheus_tsdb",
    importpath = "github.com/prometheus/tsdb",
    sum = "h1:YZcsG11NqnK4czYLrWd9mpEuAJIHVQLwdrleYfszMAA=",
    version = "v0.7.1",
)

go_repository(
    name = "com_github_puerkitobio_purell",
    importpath = "github.com/PuerkitoBio/purell",
    sum = "h1:WEQqlqaGbrPkxLJWfBwQmfEAE1Z7ONdDLqrN38tNFfI=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_puerkitobio_urlesc",
    importpath = "github.com/PuerkitoBio/urlesc",
    sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
    version = "v0.0.0-20170810143723-de5bf2ad4578",
)

go_repository(
    name = "com_github_quasilyte_go_consistent",
    importpath = "github.com/quasilyte/go-consistent",
    sum = "h1:JoUA0uz9U0FVFq5p4LjEq4C0VgQ0El320s3Ms0V4eww=",
    version = "v0.0.0-20190521200055-c6f3937de18c",
)

go_repository(
    name = "com_github_rcrowley_go_metrics",
    importpath = "github.com/rcrowley/go-metrics",
    sum = "h1:9ZKAASQSHhDYGoxY8uLVpewe1GDZ2vu2Tr/vTdVAkFQ=",
    version = "v0.0.0-20181016184325-3113b8401b8a",
)

go_repository(
    name = "com_github_roaringbitmap_roaring",
    importpath = "github.com/RoaringBitmap/roaring",
    sum = "h1:WJ/zIlNX4wQZ9x8Ey33O1UaD9TCTakYsdLFSBcTwH+8=",
    version = "v0.4.21",
)

go_repository(
    name = "com_github_rogpeppe_fastuuid",
    importpath = "github.com/rogpeppe/fastuuid",
    sum = "h1:gu+uRPtBe88sKxUCEXRoeCvVG90TJmwhiqRpvdhQFng=",
    version = "v0.0.0-20150106093220-6724a57986af",
)

go_repository(
    name = "com_github_rogpeppe_go_internal",
    importpath = "github.com/rogpeppe/go-internal",
    sum = "h1:XU784Pr0wdahMY2bYcyK6N1KuaRAdLtqD4qd8D18Bfs=",
    version = "v1.3.2",
)

go_repository(
    name = "com_github_russellhaering_gosaml2",
    importpath = "github.com/russellhaering/gosaml2",
    replace = "github.com/sourcegraph/gosaml2",
    sum = "h1:tXj2rcIvqVGM6SOt6DkGVxcjVE8c03hEethti8D77O8=",
    version = "v0.0.0-20190712190530-f05918046bab",
)

go_repository(
    name = "com_github_russellhaering_goxmldsig",
    importpath = "github.com/russellhaering/goxmldsig",
    sum = "h1:J4AOUcOh/t1XbQcJfkEqhzgvMJ2tDxdCVvmHxW5QXao=",
    version = "v0.0.0-20180430223755-7acd5e4a6ef7",
)

go_repository(
    name = "com_github_russross_blackfriday",
    importpath = "github.com/russross/blackfriday",
    replace = "github.com/russross/blackfriday",
    sum = "h1:HyvC0ARfnZBqnXwABFeSZHpKvJHJJfPz81GNueLj0oo=",
    version = "v1.5.2",
)

go_repository(
    name = "com_github_satori_go_uuid",
    importpath = "github.com/satori/go.uuid",
    sum = "h1:0uYX9dsZ2yD7q2RtLRtPSdGDWzjeM3TbMJP9utgA0ww=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_securego_gosec",
    importpath = "github.com/securego/gosec",
    sum = "h1:qlgneduvrO2HHdK1CcP+P1vlTAkT3J64A+1qd+PobNg=",
    version = "v0.0.0-20191008095658-28c1128b7336",
)

go_repository(
    name = "com_github_segmentio_fasthash",
    importpath = "github.com/segmentio/fasthash",
    sum = "h1:U+9f+rh5LxMOquTrEKNw1Z3JgsBlms9QoReNfUo+fws=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_sergi_go_diff",
    importpath = "github.com/sergi/go-diff",
    sum = "h1:Kpca3qRNrduNnOQeazBd0ysaKrUJiIuISHxogkT9RPQ=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_shirou_gopsutil",
    importpath = "github.com/shirou/gopsutil",
    sum = "h1:WokF3GuxBeL+n4Lk4Fa8v9mbdjlrl7bHuneF4N1bk2I=",
    version = "v0.0.0-20190901111213-e4ec7b275ada",
)

go_repository(
    name = "com_github_shirou_w32",
    importpath = "github.com/shirou/w32",
    sum = "h1:udFKJ0aHUL60LboW/A+DfgoHVedieIzIXE8uylPue0U=",
    version = "v0.0.0-20160930032740-bb4de0191aa4",
)

go_repository(
    name = "com_github_shopify_sarama",
    importpath = "github.com/Shopify/sarama",
    sum = "h1:9oksLxC6uxVPHPVYUmq6xhr1BOF/hHobWH2UzO67z1s=",
    version = "v1.19.0",
)

go_repository(
    name = "com_github_shopify_toxiproxy",
    importpath = "github.com/Shopify/toxiproxy",
    sum = "h1:TKdv8HiTLgE5wdJuEML90aBgNWsokNbMijUGhmcoBJc=",
    version = "v2.1.4+incompatible",
)

go_repository(
    name = "com_github_shopspring_decimal",
    importpath = "github.com/shopspring/decimal",
    sum = "h1:pntxY8Ary0t43dCZ5dqY4YTJCObLY1kIXl0uzMv+7DE=",
    version = "v0.0.0-20180709203117-cd690d0c9e24",
)

go_repository(
    name = "com_github_shurcool_github_flavored_markdown",
    importpath = "github.com/shurcooL/github_flavored_markdown",
    sum = "h1:qb9IthCFBmROJ6YBS31BEMeSYjOscSiG+EO+JVNTz64=",
    version = "v0.0.0-20181002035957-2122de532470",
)

go_repository(
    name = "com_github_shurcool_go",
    importpath = "github.com/shurcooL/go",
    sum = "h1:SpaoQDTgpo2YZkvmr2mtgloFFfPTjtLMlZkQtNAPQik=",
    version = "v0.0.0-20190704215121-7189cc372560",
)

go_repository(
    name = "com_github_shurcool_go_goon",
    importpath = "github.com/shurcooL/go-goon",
    sum = "h1:llrF3Fs4018ePo4+G/HV/uQUqEI1HMDjCeOf2V6puPc=",
    version = "v0.0.0-20170922171312-37c2f522c041",
)

go_repository(
    name = "com_github_shurcool_highlight_diff",
    importpath = "github.com/shurcooL/highlight_diff",
    sum = "h1:KaKXZldeYH73dpQL+Nr38j1r5BgpAYQjYvENOUpIZDQ=",
    version = "v0.0.0-20181222201841-111da2e7d480",
)

go_repository(
    name = "com_github_shurcool_highlight_go",
    importpath = "github.com/shurcooL/highlight_go",
    sum = "h1:a6a6gGfBoO2ty+yyHNd7M6gkp37EwE3GIoycUnLo1Oo=",
    version = "v0.0.0-20181215221002-9d8641ddf2e1",
)

go_repository(
    name = "com_github_shurcool_httpfs",
    importpath = "github.com/shurcooL/httpfs",
    sum = "h1:bUGsEnyNbVPw06Bs80sCeARAlK8lhwqGyi6UT8ymuGk=",
    version = "v0.0.0-20190707220628-8d4bc4ba7749",
)

go_repository(
    name = "com_github_shurcool_httpgzip",
    importpath = "github.com/shurcooL/httpgzip",
    sum = "h1:mj/nMDAwTBiaCqMEs4cYCqF7pO6Np7vhy1D1wcQGz+E=",
    version = "v0.0.0-20190720172056-320755c1c1b0",
)

go_repository(
    name = "com_github_shurcool_octicon",
    importpath = "github.com/shurcooL/octicon",
    sum = "h1:C2+alklsN4yRHXaOX3v9TuCGlTSwZQjSnN88nLGVhg8=",
    version = "v0.0.0-20190930024621-43309dfb482e",
)

go_repository(
    name = "com_github_shurcool_sanitized_anchor_name",
    importpath = "github.com/shurcooL/sanitized_anchor_name",
    sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_shurcool_vfsgen",
    importpath = "github.com/shurcooL/vfsgen",
    sum = "h1:ug7PpSOB5RBPK1Kg6qskGBoP3Vnj/aNYFTznWvlkGo0=",
    version = "v0.0.0-20181202132449-6a9ea43bcacd",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    importpath = "github.com/sirupsen/logrus",
    sum = "h1:SPIRibHv4MatM3XXNO2BJeFLZwZ2LvZgfQ5+UNI2im4=",
    version = "v1.4.2",
)

go_repository(
    name = "com_github_sloonz_go_qprintable",
    importpath = "github.com/sloonz/go-qprintable",
    sum = "h1:kr3of2TY0avjMhryOvEOUExpDF5yYRs3PFUGpwsawUw=",
    version = "v0.0.0-20160203160305-775b3a4592d5",
)

go_repository(
    name = "com_github_soheilhy_cmux",
    importpath = "github.com/soheilhy/cmux",
    sum = "h1:0HKaf1o97UwFjHH9o5XsHUOF+tqmdA7KEzXLpiyaw0E=",
    version = "v0.1.4",
)

go_repository(
    name = "com_github_sourcegraph_annotate",
    importpath = "github.com/sourcegraph/annotate",
    sum = "h1:yKm7XZV6j9Ev6lojP2XaIshpT4ymkqhMeSghO5Ps00E=",
    version = "v0.0.0-20160123013949-f4cad6c6324d",
)

go_repository(
    name = "com_github_sourcegraph_ctxvfs",
    importpath = "github.com/sourcegraph/ctxvfs",
    sum = "h1:v4/JVxZSPWifxmICRqgXK7khThjw03RfdGhyeA2S4EQ=",
    version = "v0.0.0-20180418081416-2b65f1b1ea81",
)

go_repository(
    name = "com_github_sourcegraph_docsite",
    importpath = "github.com/sourcegraph/docsite",
    sum = "h1:FyxKjkXE1GblbpF0Ugj9ZYEE4fI5Dxk+su81QgMjiFI=",
    version = "v0.0.0-20190808202107-429b97fc9faf",
)

go_repository(
    name = "com_github_sourcegraph_go_diff",
    importpath = "github.com/sourcegraph/go-diff",
    sum = "h1:gO6i5zugwzo1RVTvgvfwCOSVegNuvnNi6bAD1QCmkHs=",
    version = "v0.5.1",
)

go_repository(
    name = "com_github_sourcegraph_go_jsonschema",
    importpath = "github.com/sourcegraph/go-jsonschema",
    sum = "h1:a/t5nL3D1Varb1mLdcOGcG6MI1+9envp19Bwb3M4AAs=",
    version = "v0.0.0-20191016093209-4dfde5805930",
)

go_repository(
    name = "com_github_sourcegraph_go_langserver",
    importpath = "github.com/sourcegraph/go-langserver",
    sum = "h1:EX1bSaIbVia2U/Pt/2Z62y8wJS4Z4iNiK5VCeUKuBx8=",
    version = "v2.0.1-0.20181108233942-4a51fa2e1238+incompatible",
)

go_repository(
    name = "com_github_sourcegraph_go_lsp",
    importpath = "github.com/sourcegraph/go-lsp",
    sum = "h1:O1d7nVzpGmP5pGAZBSlp9TSpjNwwI0xThxhPd9oVJuU=",
    version = "v0.0.0-20181119182933-0c7d621186c1",
)

go_repository(
    name = "com_github_sourcegraph_gosyntect",
    importpath = "github.com/sourcegraph/gosyntect",
    sum = "h1:bbs7cDhEjewPB6aicTDJ15tseQAsl8bSGDVsnIuG0c0=",
    version = "v0.0.0-20191003053245-e91d603ba4eb",
)

go_repository(
    name = "com_github_sourcegraph_jsonschemadoc",
    importpath = "github.com/sourcegraph/jsonschemadoc",
    sum = "h1:MXlcJZ1VL5nNGkCj6ZTT71P4pImPkeG2lvzcJYzGvU4=",
    version = "v0.0.0-20190214000648-1850b818f08c",
)

go_repository(
    name = "com_github_sourcegraph_jsonx",
    importpath = "github.com/sourcegraph/jsonx",
    sum = "h1:MrlKMpoGse4bCneDoK/c+ZbPGqOP5Hme5ulatc8smbQ=",
    version = "v0.0.0-20190114210550-ba8cb36a8614",
)

go_repository(
    name = "com_github_sourcegraph_syntaxhighlight",
    importpath = "github.com/sourcegraph/syntaxhighlight",
    sum = "h1:qpG93cPwA5f7s/ZPBJnGOYQNK/vKsaDaseuKT5Asee8=",
    version = "v0.0.0-20170531221838-bd320f5d308e",
)

go_repository(
    name = "com_github_spaolacci_murmur3",
    importpath = "github.com/spaolacci/murmur3",
    sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
    version = "v0.0.0-20180118202830-f09979ecbc72",
)

go_repository(
    name = "com_github_spf13_afero",
    importpath = "github.com/spf13/afero",
    sum = "h1:5jhuqJyZCZf2JRofRvN/nIFgIWNzPa3/Vz8mYylgbWc=",
    version = "v1.2.2",
)

go_repository(
    name = "com_github_spf13_cast",
    importpath = "github.com/spf13/cast",
    sum = "h1:oget//CVOEoFewqQxwr0Ej5yjygnqGkvggSE/gB35Q8=",
    version = "v1.3.0",
)

go_repository(
    name = "com_github_spf13_cobra",
    importpath = "github.com/spf13/cobra",
    sum = "h1:f0B+LkLX6DtmRH1isoNA9VTtNUK9K8xYd28JNNfOv/s=",
    version = "v0.0.5",
)

go_repository(
    name = "com_github_spf13_jwalterweatherman",
    importpath = "github.com/spf13/jwalterweatherman",
    sum = "h1:ue6voC5bR5F8YxI5S67j9i582FU4Qvo2bmqnqMYADFk=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_spf13_pflag",
    importpath = "github.com/spf13/pflag",
    sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
    version = "v1.0.5",
)

go_repository(
    name = "com_github_spf13_viper",
    importpath = "github.com/spf13/viper",
    sum = "h1:yXHLWeravcrgGyFSyCgdYpXQ9dR9c/WED3pg1RhxqEU=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_sqs_httpgzip",
    importpath = "github.com/sqs/httpgzip",
    sum = "h1:35Om/vTf7BiYaXjy+V6enYxBRzWOhoFI0g/Uzc0QL8Y=",
    version = "v0.0.0-20180622165210-91da61ed4dff",
)

go_repository(
    name = "com_github_src_d_enry_v2",
    importpath = "github.com/src-d/enry/v2",
    sum = "h1:z1L8t+B8bh3mmjPkJrgOTnVRpFGmTPJsplHX9wAn6BI=",
    version = "v2.1.0",
)

go_repository(
    name = "com_github_src_d_gcfg",
    importpath = "github.com/src-d/gcfg",
    sum = "h1:xXbNR5AlLSA315x2UO+fTSSAXCDf+Ar38/6oyGbDKQ4=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_src_d_go_oniguruma",
    importpath = "github.com/src-d/go-oniguruma",
    sum = "h1:EG+Nm5n2JqWUaCjtM0NtutPxU7ZN5Tp50GWrrV8bTww=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_stackexchange_wmi",
    importpath = "github.com/StackExchange/wmi",
    sum = "h1:fLjPD/aNc3UIOA6tDi6QXUemppXK3P9BI7mr2hd6gx8=",
    version = "v0.0.0-20180116203802-5d049714c4a6",
)

go_repository(
    name = "com_github_stretchr_objx",
    importpath = "github.com/stretchr/objx",
    sum = "h1:Hbg2NidpLE8veEBkEZTL3CvlkUIVzuU9jDplZO54c48=",
    version = "v0.2.0",
)

go_repository(
    name = "com_github_stretchr_testify",
    importpath = "github.com/stretchr/testify",
    sum = "h1:2E4SXV/wtOkTonXsotYi4li6zVWxYlZuYNCXe9XRJyk=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_stripe_stripe_go",
    importpath = "github.com/stripe/stripe-go",
    sum = "h1:OD5y3FppSAVDs0f9q9FTfcnYWWqsPsKKV9b5kV8po7o=",
    version = "v65.1.0+incompatible",
)

go_repository(
    name = "com_github_stvp_tempredis",
    importpath = "github.com/stvp/tempredis",
    sum = "h1:QVqDTf3h2WHt08YuiTGPZLls0Wq99X9bWd0Q5ZSBesM=",
    version = "v0.0.0-20181119212430-b82af8480203",
)

go_repository(
    name = "com_github_temoto_robotstxt",
    importpath = "github.com/temoto/robotstxt",
    sum = "h1:Gh8RCs8ouX3hRSxxK7B1mO5RFByQ4CmJZDwgom++JaA=",
    version = "v1.1.1",
)

go_repository(
    name = "com_github_tidwall_pretty",
    importpath = "github.com/tidwall/pretty",
    sum = "h1:BP2bjP495BBPaBcS5rmqviTfrOkN5rO5ceKAMRZCRFc=",
    version = "v0.0.0-20180105212114-65a9db5fad51",
)

go_repository(
    name = "com_github_timakin_bodyclose",
    importpath = "github.com/timakin/bodyclose",
    sum = "h1:RumXZ56IrCj4CL+g1b9OL/oH0QnsF976bC8xQFYUD5Q=",
    version = "v0.0.0-20190930140734-f7f2e9bca95e",
)

go_repository(
    name = "com_github_tinylib_msgp",
    importpath = "github.com/tinylib/msgp",
    sum = "h1:9fQd+ICuRIu/ue4vxJZu6/LzxN0HwMds2nq/0cFvxHU=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_tmc_grpc_websocket_proxy",
    importpath = "github.com/tmc/grpc-websocket-proxy",
    sum = "h1:LnC5Kc/wtumK+WB441p7ynQJzVuNRJiqddSIE3IlSEQ=",
    version = "v0.0.0-20190109142713-0ad062ec5ee5",
)

go_repository(
    name = "com_github_toqueteos_trie",
    importpath = "github.com/toqueteos/trie",
    sum = "h1:8i6pXxNUXNRAqP246iibb7w/pSFquNTQ+uNfriG7vlk=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_uber_go_atomic",
    importpath = "github.com/uber-go/atomic",
    sum = "h1:yOuPqEq4ovnhEjpHmfFwsqBXDYbQeT6Nb0bwD6XnD5o=",
    version = "v1.4.0",
)

go_repository(
    name = "com_github_uber_gonduit",
    importpath = "github.com/uber/gonduit",
    replace = "github.com/sourcegraph/gonduit",
    sum = "h1:0BTiXaxNto/Jm0uvd/LdnZZr6JOkZz+VqPulU+YoB34=",
    version = "v0.4.0",
)

go_repository(
    name = "com_github_uber_jaeger_client_go",
    importpath = "github.com/uber/jaeger-client-go",
    sum = "h1:pbwbYfHUoaase0oPQOdZ1GcaUjImYGimUXSQ/+8+Z8Q=",
    version = "v2.19.0+incompatible",
)

go_repository(
    name = "com_github_uber_jaeger_lib",
    importpath = "github.com/uber/jaeger-lib",
    sum = "h1:MxZXOiR2JuoANZ3J6DE/U0kSFv/eJ/GfSYVCjK7dyaw=",
    version = "v2.2.0+incompatible",
)

go_repository(
    name = "com_github_ugorji_go",
    importpath = "github.com/ugorji/go",
    sum = "h1:/68gy2h+1mWMrwZFeD1kQialdSzAb432dtpeJ42ovdo=",
    version = "v1.1.7",
)

go_repository(
    name = "com_github_ugorji_go_codec",
    importpath = "github.com/ugorji/go/codec",
    sum = "h1:2SvQaVZ1ouYrrKKwoSk2pzd4A9evlKJb9oTL+OaLUSs=",
    version = "v1.1.7",
)

go_repository(
    name = "com_github_ultraware_funlen",
    importpath = "github.com/ultraware/funlen",
    sum = "h1:Av96YVBwwNSe4MLR7iI/BIa3VyI7/djnto/pK3Uxbdo=",
    version = "v0.0.2",
)

go_repository(
    name = "com_github_ultraware_whitespace",
    importpath = "github.com/ultraware/whitespace",
    sum = "h1:If7Va4cM03mpgrNH9k49/VOicWpGoG70XPBFFODYDsg=",
    version = "v0.0.4",
)

go_repository(
    name = "com_github_uudashr_gocognit",
    importpath = "github.com/uudashr/gocognit",
    sum = "h1:NST9SQhYHpFiyKdXn8UAiogdB1xkt5RnA1/rR/oBcGw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_valyala_bytebufferpool",
    importpath = "github.com/valyala/bytebufferpool",
    sum = "h1:GqA5TC/0021Y/b9FG4Oi9Mr3q7XYx6KllzawFIhcdPw=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_valyala_fasthttp",
    importpath = "github.com/valyala/fasthttp",
    sum = "h1:dzZJf2IuMiclVjdw0kkT+f9u4YdrapbNyGAN47E/qnk=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_valyala_fasttemplate",
    importpath = "github.com/valyala/fasttemplate",
    sum = "h1:tY9CJiPnMXf1ERmG2EyK7gNUd+c6RKGD0IfU8WdUSz8=",
    version = "v1.0.1",
)

go_repository(
    name = "com_github_valyala_quicktemplate",
    importpath = "github.com/valyala/quicktemplate",
    sum = "h1:BaO1nHTkspYzmAjPXj0QiDJxai96tlcZyKcI9dyEGvM=",
    version = "v1.2.0",
)

go_repository(
    name = "com_github_valyala_tcplisten",
    importpath = "github.com/valyala/tcplisten",
    sum = "h1:0R4NLDRDZX6JcmhJgXi5E4b8Wg84ihbmUKp/GvSPEzc=",
    version = "v0.0.0-20161114210144-ceec8f93295a",
)

go_repository(
    name = "com_github_vmihailenco_msgpack",
    importpath = "github.com/vmihailenco/msgpack",
    sum = "h1:dSLoQfGFAo3F6OoNhwUmLwVgaUXK79GlxNBwueZn0xI=",
    version = "v4.0.4+incompatible",
)

go_repository(
    name = "com_github_willf_bitset",
    importpath = "github.com/willf/bitset",
    sum = "h1:NotGKqX0KwQ72NUzqrjZq5ipPNDQex9lo3WpaS8L2sc=",
    version = "v1.1.10",
)

go_repository(
    name = "com_github_wsxiaoys_terminal",
    importpath = "github.com/wsxiaoys/terminal",
    sum = "h1:3UeQBvD0TFrlVjOeLOBz+CPAI8dnbqNSVwUwRrkp7vQ=",
    version = "v0.0.0-20160513160801-0940f3fc43a0",
)

go_repository(
    name = "com_github_xanzy_go_gitlab",
    importpath = "github.com/xanzy/go-gitlab",
    sum = "h1:+1BWDry84G5PzsnzG9DI4YjPbHeWKyouM0q0gfDPKgY=",
    version = "v0.20.1",
)

go_repository(
    name = "com_github_xanzy_ssh_agent",
    importpath = "github.com/xanzy/ssh-agent",
    sum = "h1:TCbipTQL2JiiCprBWx9frJ2eJlCYT00NmctrHxVAr70=",
    version = "v0.2.1",
)

go_repository(
    name = "com_github_xdg_scram",
    importpath = "github.com/xdg/scram",
    sum = "h1:u40Z8hqBAAQyv+vATcGgV0YCnDjqSL7/q/JyPhhJSPk=",
    version = "v0.0.0-20180814205039-7eeb5667e42c",
)

go_repository(
    name = "com_github_xdg_stringprep",
    importpath = "github.com/xdg/stringprep",
    sum = "h1:d9X0esnoa3dFsV0FG35rAT0RIhYFlPq7MiP+DW89La0=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_xeipuuv_gojsonpointer",
    importpath = "github.com/xeipuuv/gojsonpointer",
    sum = "h1:zGWFAtiMcyryUHoUjUJX0/lt1H2+i2Ka2n+D3DImSNo=",
    version = "v0.0.0-20190905194746-02993c407bfb",
)

go_repository(
    name = "com_github_xeipuuv_gojsonreference",
    importpath = "github.com/xeipuuv/gojsonreference",
    sum = "h1:EzJWgHovont7NscjpAxXsDA8S8BMYve8Y5+7cuRE7R0=",
    version = "v0.0.0-20180127040603-bd5ef7bd5415",
)

go_repository(
    name = "com_github_xeipuuv_gojsonschema",
    importpath = "github.com/xeipuuv/gojsonschema",
    sum = "h1:ngVtJC9TY/lg0AA/1k48FYhBrhRoFlEmWzsehpNAaZg=",
    version = "v1.1.0",
)

go_repository(
    name = "com_github_xeonx_timeago",
    importpath = "github.com/xeonx/timeago",
    sum = "h1:9rRzv48GlJC0vm+iBpLcWAr8YbETyN9Vij+7h2ammz4=",
    version = "v1.0.0-rc4",
)

go_repository(
    name = "com_github_xiang90_probing",
    importpath = "github.com/xiang90/probing",
    sum = "h1:eY9dn8+vbi4tKz5Qo6v2eYzo7kUS51QINcR5jNpbZS8=",
    version = "v0.0.0-20190116061207-43a291ad63a2",
)

go_repository(
    name = "com_github_xordataexchange_crypt",
    importpath = "github.com/xordataexchange/crypt",
    sum = "h1:ESFSdwYZvkeru3RtdrYueztKhOBCSAAzS4Gf+k0tEow=",
    version = "v0.0.3-0.20170626215501-b2862e3d0a77",
)

go_repository(
    name = "com_github_zenazn_goji",
    importpath = "github.com/zenazn/goji",
    sum = "h1:RSQQAbXGArQ0dIDEq+PI6WqN6if+5KHu6x2Cx/GXLTQ=",
    version = "v0.9.0",
)

go_repository(
    name = "com_gitlab_nyarla_go_crypt",
    importpath = "gitlab.com/nyarla/go-crypt",
    sum = "h1:7gd+rd8P3bqcn/96gOZa3F5dpJr/vEiDQYlNb/y2uNs=",
    version = "v0.0.0-20160106005555-d9a5dc2b789b",
)

go_repository(
    name = "com_google_cloud_go",
    importpath = "cloud.google.com/go",
    sum = "h1:AVXDdKsrtX33oR9fbCMu/+c1o8Ofjq6Ku/MInaLVg5Y=",
    version = "v0.46.3",
)

go_repository(
    name = "com_google_cloud_go_bigquery",
    importpath = "cloud.google.com/go/bigquery",
    sum = "h1:Fm4frQlg3F76NQKi1VHb+PskluqTvlBpzdETxNTlbyw=",
    version = "v1.1.0",
)

go_repository(
    name = "com_google_cloud_go_datastore",
    importpath = "cloud.google.com/go/datastore",
    sum = "h1:Kt+gOPPp2LEPWp8CSfxhsM8ik9CcyE/gYu+0r+RnZvM=",
    version = "v1.0.0",
)

go_repository(
    name = "com_google_cloud_go_pubsub",
    importpath = "cloud.google.com/go/pubsub",
    sum = "h1:W9tAK3E57P75u0XLLR82LZyw8VpAnhmyTOxW9qzmyj8=",
    version = "v1.0.1",
)

go_repository(
    name = "com_google_cloud_go_storage",
    importpath = "cloud.google.com/go/storage",
    sum = "h1:KYV0dnmEcOuxTd8YLiuQqfx8PzSwDeuDvYGoa5+DbDI=",
    version = "v1.1.0",
)

go_repository(
    name = "com_shuralyov_dmitri_gpu_mtl",
    importpath = "dmitri.shuralyov.com/gpu/mtl",
    sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
    version = "v0.0.0-20190408044501-666a987793e9",
)

go_repository(
    name = "com_sourcegraph_sqs_pbtypes",
    importpath = "sourcegraph.com/sqs/pbtypes",
    sum = "h1:f7lAwqviDEGvON4kRv0o5V7FT/IQK+tbkF664XMbP3o=",
    version = "v1.0.0",
)

go_repository(
    name = "in_gopkg_airbrake_gobrake_v2",
    importpath = "gopkg.in/airbrake/gobrake.v2",
    sum = "h1:7z2uVWwn7oVeeugY1DtlPAy5H+KYgB1KeKTnqjNatLo=",
    version = "v2.0.9",
)

go_repository(
    name = "in_gopkg_alecthomas_kingpin_v2",
    importpath = "gopkg.in/alecthomas/kingpin.v2",
    sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
    version = "v2.2.6",
)

go_repository(
    name = "in_gopkg_alexcesaro_statsd_v2",
    importpath = "gopkg.in/alexcesaro/statsd.v2",
    sum = "h1:FXkZSCZIH17vLCO5sO2UucTHsH9pc+17F6pl3JVCwMc=",
    version = "v2.0.0",
)

go_repository(
    name = "in_gopkg_check_v1",
    importpath = "gopkg.in/check.v1",
    sum = "h1:YR8cESwS4TdDjEe65xsg0ogRM/Nc3DYOhEAlW+xobZo=",
    version = "v1.0.0-20190902080502-41f04d3bba15",
)

go_repository(
    name = "in_gopkg_errgo_v2",
    importpath = "gopkg.in/errgo.v2",
    sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
    version = "v2.1.0",
)

go_repository(
    name = "in_gopkg_fsnotify_v1",
    importpath = "gopkg.in/fsnotify.v1",
    sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
    version = "v1.4.7",
)

go_repository(
    name = "in_gopkg_gemnasium_logrus_airbrake_hook_v2",
    importpath = "gopkg.in/gemnasium/logrus-airbrake-hook.v2",
    sum = "h1:OAj3g0cR6Dx/R07QgQe8wkA9RNjB2u4i700xBkIT4e0=",
    version = "v2.1.2",
)

go_repository(
    name = "in_gopkg_go_playground_assert_v1",
    importpath = "gopkg.in/go-playground/assert.v1",
    sum = "h1:xoYuJVE7KT85PYWrN730RguIQO0ePzVRfFMXadIrXTM=",
    version = "v1.2.1",
)

go_repository(
    name = "in_gopkg_go_playground_validator_v8",
    importpath = "gopkg.in/go-playground/validator.v8",
    sum = "h1:lFB4DoMU6B626w8ny76MV7VX6W2VHct2GVOI3xgiMrQ=",
    version = "v8.18.2",
)

go_repository(
    name = "in_gopkg_inconshreveable_log15_v2",
    importpath = "gopkg.in/inconshreveable/log15.v2",
    sum = "h1:RlWgLqCMMIYYEVcAR5MDsuHlVkaIPDAF+5Dehzg8L5A=",
    version = "v2.0.0-20180818164646-67afb5ed74ec",
)

go_repository(
    name = "in_gopkg_inf_v0",
    importpath = "gopkg.in/inf.v0",
    sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
    version = "v0.9.1",
)

go_repository(
    name = "in_gopkg_jpoehls_gophermail_v0",
    importpath = "gopkg.in/jpoehls/gophermail.v0",
    sum = "h1:Un0HKXHsvpUSZPX77tzIBx2Qdrd0bst8wE0Jh00hovk=",
    version = "v0.0.0-20160410235621-62941eab772c",
)

go_repository(
    name = "in_gopkg_karlseguin_expect_v1",
    importpath = "gopkg.in/karlseguin/expect.v1",
    sum = "h1:9u0iUltnhFbJTHaSIH0EP+cuTU5rafIgmcsEsg2JQFw=",
    version = "v1.0.1",
)

go_repository(
    name = "in_gopkg_resty_v1",
    importpath = "gopkg.in/resty.v1",
    sum = "h1:CuXP0Pjfw9rOuY6EP+UvtNvt5DSqHpIxILZKT/quCZI=",
    version = "v1.12.0",
)

go_repository(
    name = "in_gopkg_russross_blackfriday_v2",
    importpath = "gopkg.in/russross/blackfriday.v2",
    replace = "github.com/russross/blackfriday/v2",
    sum = "h1:lPqVAte+HuHNfhJ/0LC98ESWRz8afy9tM/0RK8m9o+Q=",
    version = "v2.0.1",
)

go_repository(
    name = "in_gopkg_square_go_jose_v2",
    importpath = "gopkg.in/square/go-jose.v2",
    sum = "h1:SK5KegNXmKmqE342YYN2qPHEnUYeoMiXXl1poUlI+o4=",
    version = "v2.3.1",
)

go_repository(
    name = "in_gopkg_src_d_go_billy_v4",
    importpath = "gopkg.in/src-d/go-billy.v4",
    sum = "h1:0SQA1pRztfTFx2miS8sA97XvooFeNOmvUenF4o0EcVg=",
    version = "v4.3.2",
)

go_repository(
    name = "in_gopkg_src_d_go_git_fixtures_v3",
    importpath = "gopkg.in/src-d/go-git-fixtures.v3",
    sum = "h1:ivZFOIltbce2Mo8IjzUHAFoq/IylO9WHhNOAJK+LsJg=",
    version = "v3.5.0",
)

go_repository(
    name = "in_gopkg_src_d_go_git_v4",
    importpath = "gopkg.in/src-d/go-git.v4",
    sum = "h1:SRtFyV8Kxc0UP7aCHcijOMQGPxHSmMOPrzulQWolkYE=",
    version = "v4.13.1",
)

go_repository(
    name = "in_gopkg_tomb_v1",
    importpath = "gopkg.in/tomb.v1",
    sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
    version = "v1.0.0-20141024135613-dd632973f1e7",
)

go_repository(
    name = "in_gopkg_toqueteos_substring_v1",
    importpath = "gopkg.in/toqueteos/substring.v1",
    sum = "h1:urLqCeMm6x/eTuQa1oZerNw8N1KNOIp5hD5kGL7lFsE=",
    version = "v1.0.2",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    importpath = "gopkg.in/warnings.v0",
    sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
    version = "v0.1.2",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    importpath = "gopkg.in/yaml.v2",
    sum = "h1:/eiJrUcujPVeJ3xlSWaiNi3uSVmDGBK1pDHUHAnao1I=",
    version = "v2.2.4",
)

go_repository(
    name = "io_etcd_go_bbolt",
    importpath = "go.etcd.io/bbolt",
    sum = "h1:Z/90sZLPOeCy2PwprqkFa25PdkusRzaj9P8zm/KNyvk=",
    version = "v1.3.2",
)

go_repository(
    name = "io_opencensus_go",
    importpath = "go.opencensus.io",
    sum = "h1:8dP3SGL7MPB94crU3bEPplMPe83FI4EouesJUeFHv50=",
    version = "v0.22.1",
)

go_repository(
    name = "io_rsc_binaryregexp",
    importpath = "rsc.io/binaryregexp",
    sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
    version = "v0.2.0",
)

go_repository(
    name = "io_rsc_pdf",
    importpath = "rsc.io/pdf",
    sum = "h1:k1MczvYDUvJBe93bYd7wrZLLUEcLZAuF824/I4e5Xr4=",
    version = "v0.1.1",
)

go_repository(
    name = "net_starlark_go",
    importpath = "go.starlark.net",
    sum = "h1:WR7X1xgXJlXhQBdorVc9Db3RhwG+J/kp6bLuMyJjfVw=",
    version = "v0.0.0-20190919145610-979af19b165c",
)

go_repository(
    name = "org_golang_google_api",
    importpath = "google.golang.org/api",
    sum = "h1:n/qM3q0/rV2F0pox7o0CvNhlPvZAo7pLbef122cbLJ0=",
    version = "v0.11.0",
)

go_repository(
    name = "org_golang_google_appengine",
    importpath = "google.golang.org/appengine",
    sum = "h1:tycE03LOZYQNhDpS27tcQdAzLCVMaj7QT2SXxebnpCM=",
    version = "v1.6.5",
)

go_repository(
    name = "org_golang_google_genproto",
    importpath = "google.golang.org/genproto",
    sum = "h1:4HYDjxeNXAOTv3o1N2tjo8UUSlhQgAD52FVkwxnWgM8=",
    version = "v0.0.0-20191009194640-548a555dbc03",
)

go_repository(
    name = "org_golang_google_grpc",
    importpath = "google.golang.org/grpc",
    sum = "h1:vb/1TCsVn3DcJlQ0Gs1yB1pKI6Do2/QNwxdKqmc/b0s=",
    version = "v1.24.0",
)

go_repository(
    name = "org_golang_x_arch",
    importpath = "golang.org/x/arch",
    sum = "h1:QlVATYS7JBoZMVaf+cNjb90WD/beKVHnIxFKT4QaHVI=",
    version = "v0.0.0-20190927153633-4e8777c89be4",
)

go_repository(
    name = "org_golang_x_crypto",
    importpath = "golang.org/x/crypto",
    sum = "h1:/b4NMZurYfBIQyRMqaPGMDeUrSW6gU7/7Hv6owY1Vjk=",
    version = "v0.0.0-20191010185427-af544f31c8ac",
)

go_repository(
    name = "org_golang_x_exp",
    importpath = "golang.org/x/exp",
    sum = "h1:n9HxLrNxWWtEb1cA950nuEEj3QnKbtsCJ6KjcgisNUs=",
    version = "v0.0.0-20191002040644-a1355ae1e2c3",
)

go_repository(
    name = "org_golang_x_image",
    importpath = "golang.org/x/image",
    sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
    version = "v0.0.0-20190802002840-cff245a6509b",
)

go_repository(
    name = "org_golang_x_lint",
    importpath = "golang.org/x/lint",
    sum = "h1:5hukYrvBGR8/eNkX5mdUezrA6JiaEZDtJb9Ei+1LlBs=",
    version = "v0.0.0-20190930215403-16217165b5de",
)

go_repository(
    name = "org_golang_x_mobile",
    importpath = "golang.org/x/mobile",
    sum = "h1:4+4C/Iv2U4fMZBiMCc98MG1In4gJY5YRhtpDNeDeHWs=",
    version = "v0.0.0-20190719004257-d2bd2a29d028",
)

go_repository(
    name = "org_golang_x_mod",
    importpath = "golang.org/x/mod",
    sum = "h1:sfUMP1Gu8qASkorDVjnMuvgJzwFbTZSeXFiGBYAVdl4=",
    version = "v0.1.0",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:TR699M2v0qoKTOHxeLgp6zPqaQNs74f01a/ob9W0qko=",
    version = "v0.0.0-20191009170851-d66e71096ffb",
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    sum = "h1:SVwTIAaPC2U/AvvLNZ2a7OVsmBpC8L5BlwK1whH3hm0=",
    version = "v0.0.0-20190604053449-0f29369cfe45",
)

go_repository(
    name = "org_golang_x_sync",
    importpath = "golang.org/x/sync",
    sum = "h1:vcxGaoTs7kV8m5Np9uUNQin4BrLOthgV7252N8V+FwY=",
    version = "v0.0.0-20190911185100-cd5d95a43a6e",
)

go_repository(
    name = "org_golang_x_sys",
    importpath = "golang.org/x/sys",
    sum = "h1:/XfQ9z7ib8eEJX2hdgFTZJ/ntt0swNk5oYBziWeTCvY=",
    version = "v0.0.0-20191010194322-b09406accb47",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:tW2bmiBqwgJj/UpqtC8EpXEZVYOwU0yG4iWbprSVAcs=",
    version = "v0.3.2",
)

go_repository(
    name = "org_golang_x_time",
    importpath = "golang.org/x/time",
    sum = "h1:xQwXv67TxFo9nC1GJFyab5eq/5B590r6RlnL/G8Sz7w=",
    version = "v0.0.0-20190921001708-c4c64cad1fd0",
)

go_repository(
    name = "org_golang_x_tools",
    importpath = "golang.org/x/tools",
    sum = "h1:Cgj5oVkw7Gktu56MAiU0r1u0jyuT6jmtOzcAJwLj89c=",
    version = "v0.0.0-20191010201905-e5ffc44a6fee",
)

go_repository(
    name = "org_golang_x_xerrors",
    importpath = "golang.org/x/xerrors",
    sum = "h1:9zdDQZ7Thm29KFXgAX/+yaf3eVbP7djjWp/dXAppNCc=",
    version = "v0.0.0-20190717185122-a985d3407aa7",
)

go_repository(
    name = "org_mongodb_go_mongo_driver",
    importpath = "go.mongodb.org/mongo-driver",
    sum = "h1:aeOqSrhl9eDRAap/3T5pCfMBEBxZ0vuXBP+RMtp2KX8=",
    version = "v1.1.0",
)

go_repository(
    name = "org_uber_go_atomic",
    importpath = "go.uber.org/atomic",
    sum = "h1:cxzIVoETapQEqDhQu3QfnvXAV4AlzcvUCxkVUFw3+EU=",
    version = "v1.4.0",
)

go_repository(
    name = "org_uber_go_automaxprocs",
    importpath = "go.uber.org/automaxprocs",
    sum = "h1:+RUihKM+nmYUoB9w0D0Ov5TJ2PpFO2FgenTxMJiZBZA=",
    version = "v1.2.0",
)

go_repository(
    name = "org_uber_go_multierr",
    importpath = "go.uber.org/multierr",
    sum = "h1:HoEmRHQPVSqub6w2z2d2EOVs2fjyFRGyofhKuyDq0QI=",
    version = "v1.1.0",
)

go_repository(
    name = "org_uber_go_zap",
    importpath = "go.uber.org/zap",
    sum = "h1:ORx85nbTijNz8ljznvCMR1ZBIPKFn3jQrag10X2AsuM=",
    version = "v1.10.0",
)

go_repository(
    name = "tools_gotest",
    importpath = "gotest.tools",
    sum = "h1:VsBPFP1AI068pPrMxtb/S8Zkgf9xEmTLJjfM+P5UIEo=",
    version = "v2.2.0+incompatible",
)
