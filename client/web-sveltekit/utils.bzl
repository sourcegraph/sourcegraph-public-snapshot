load("@npm//client/web-sveltekit:vite/package_json.bzl", vite_bin = "bin")
load("@bazel_skylib//lib:dicts.bzl", "dicts")

def compile_app(name, test_name = None, env = {}, **kwargs):
    """compile_app produces a test and production build of the given arguments,

    where the test build has a "TEST": "1" env var set.

    Args:
        name: the name of the production build
        test_name: the name of the test build. If not provided, it defaults to name + "_test"
        env: a dictionary of environment variables to set for the build
        **kwargs: additional key-value arguments to pass to vite_bin.vite
    """
    if test_name == None:
        test_name = name + "_test"

    vite_bin.vite(
        name = name,
        env = env,
        **kwargs
    )

    vite_bin.vite(
        name = test_name,
        env = dicts.add(
            env,
            {
                "E2E_BUILD": "1",
            },
        ),
        **kwargs
    )
