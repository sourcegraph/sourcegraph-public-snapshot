load("@npm//:eslint/package_json.bzl", eslint_bin = "bin")

def eslint_test(name, deps, config, data = [], **kwargs):
    eslint_bin.eslint_test(
        name = name,
        args = [
            "--quiet",
            "--config",
            "$(location {})".format(config),
            "--resolve-plugins-relative-to {}".format(native.package_name()),
        ] + [
            "$(location {})".format(src)
            for src in data
        ],
        data = deps + data,
        **kwargs
    )
