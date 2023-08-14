def _impl(ctx):
    print(ctx.configuration.default_shell_env)
    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [ctx.outputs.out],
        progress_message = "Running squash migration for %s" % ctx.attr.db,
        use_default_shell_env = True,
        command = """
        env
        echo $PATH
        export HOME=$(pwd)
        export SG_FORCE_REPO_ROOT=$(pwd)
        {sg} migration squash-all -skip-teardown -db {db} -f {output_file}
        """.format(sg = ctx.executable._sg.path, db = ctx.attr.db, output_file = ctx.outputs.out.path),
        tools = ctx.attr._sg[DefaultInfo].default_runfiles.files,
    )

migration = rule(
    implementation = _impl,
    attrs = {
        "srcs": attr.label_list(allow_files= True, mandatory= True),
        "db": attr.string(mandatory = True),
        "out": attr.output(mandatory = True),
        "_sg": attr.label(executable = True, default = "//dev/sg:sg", cfg = "exec"),
    },
)
