load("@npm//client/web-sveltekit:vite/package_json.bzl", vite_bin = "bin")
load("@bazel_skylib//lib:dicts.bzl", "dicts")

def compile_app(name, env = {}, test_overrides = {}, **kwargs):
    """compile_app produces a test and production build of the given arguments,

    where the test build has a "TEST": "1" env var set.

    Args:
        name: the name of the production build
        env: a dictionary of environment variables to set for the production build
        test_overrides: a dictionary of vite_bin.vite arguments to override for the test build
        **kwargs: additional key-value arguments to pass to vite_bin.vite
    """
    test_name = test_overrides.pop("name", None)
    if test_name == None:
        test_name = name + "_test"

    # The production build is the default target
    vite_bin.vite(
        name = name,
        env = env,
        **kwargs
    )

    test_env = dicts.add(
        env,
        test_overrides.pop("env", {}),
        {
            "E2E_BUILD": "1",
        },
    )

    # Also produces a test build construct with the E2E_BUILD env var set
    vite_bin.vite(
        name = test_name,
        env = test_env,
        **dicts.add(kwargs, test_overrides)
    )
