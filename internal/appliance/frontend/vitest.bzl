"Wrapper around vitest_test"

load("@npm//:vitest/package_json.bzl", "bin")

def vitest(name, config, deps, **kwargs):
    bin.vitest_test(
        name = name,
        # Perform a single run without watch mode. If we want to watch we will use ibazel.
        args = ["run"],
        # Paths in the configuration file are relative to its folder, so we must use that as the
        # working directory since vite doesn't handle this itself.
        chdir = Label(config).package,
        data = [config] + deps + [
            "//:node_modules/jsdom",
            "//:node_modules/vitest",
        ],
        **kwargs
    )
