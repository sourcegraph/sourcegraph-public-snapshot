def server_integration_test(name, port, srcs, **kwargs):
    args = kwargs.get("args", [])
    data = kwargs.get("data", [])
    env = kwargs.get("env", {})
    env_inherit = kwargs.get("env_inherit", [])
    tags = kwargs.get("tags", [])
    deps = kwargs.get("deps", [])

    # We need the server image for the script to spawn it.
    data.append("//enterprise/cmd/server:image_tarball")

    # First two arguments are always the server image and the image name
    args = ["$(location //enterprise/cmd/server:image_tarball)", "server:candidate"] + args

    # Explicitly define the port, needs to be different for each test so we can run them concurrently.
    env.append("PORT", port)

    # These tests are making network calls to the running server image, so we need the network.
    tags.append("requires-network")

    # Finally, we depend on the integration runner script helper.
    deps.append("//testing/tools:integration_runner")

    native.sh_test(
        name = name,
        srcs = srcs,
        args = args,
        data = data,
        env = env,
        env_inherit = env_inherit
    )
