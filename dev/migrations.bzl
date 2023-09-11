CMD_PREAMBLE = """set -e

export HOME=$(pwd)
export SG_FORCE_REPO_ROOT=$(pwd)

if [ -n "$PG_UTILS_PATH" ]; then
    PATH="$PG_UTILS_PATH:$PATH"
fi

if [ -z "$PGUSER" ]; then
    export PGUSER="sourcegraph"
fi

if [ -z "$CODEINTEL_PGUSER" ]; then
    export CODEINTEL_PGUSER="$PGUSER"
fi

if [ -z "$CODEINSIGHTS_PGUSER" ]; then
    export CODEINSIGHTS_PGUSER="$PGUSER"
fi
"""

def _migration_impl(ctx):
    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [ctx.outputs.out],
        progress_message = "Running squash migration for %s" % ctx.attr.db,
        use_default_shell_env = True,
        execution_requirements = {"requires-network": "1"},
        command = """{cmd_preamble}

        trap "dropdb --if-exists sg-squasher-{db} && echo 'temp db sg-squasher-{db} dropped'" EXIT

        {sg} migration squash-all -skip-teardown -db {db} -f {output_file}
        """.format(
            cmd_preamble = CMD_PREAMBLE,
            sg = ctx.executable._sg.path,
            db = ctx.attr.db,
            output_file = ctx.outputs.out.path,
        ),
        tools = ctx.attr._sg[DefaultInfo].default_runfiles.files
    )

migration = rule(
    implementation = _migration_impl,
    attrs = {
        "srcs": attr.label_list(allow_files= True, mandatory= True),
        "db": attr.string(mandatory = True),
        "out": attr.output(mandatory = True),
        "_sg": attr.label(executable = True, default = "//dev/sg:sg", cfg = "exec"),
    },
)

def _describe_impl(ctx):
    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [ctx.outputs.out],
        progress_message = "Running describe migration for %s" % ctx.attr.db,
        use_default_shell_env = True,
        execution_requirements = {"requires-network": "1"},
        command = """{cmd_preamble}

        export PGDATABASE="_describe_{name}"
        dropdb --if-exists $PGDATABASE
        createdb "$PGDATABASE"
        trap "dropdb --if-exists $PGDATABASE" exit

        {sg} migration describe -db {db} --format={format} -force -out {output_file}
        """.format(
            cmd_preamble = CMD_PREAMBLE,
            sg = ctx.executable._sg.path,
            db = ctx.attr.db,
            format = ctx.attr.format,
            output_file = ctx.outputs.out.path,
            name = ctx.attr.name,
        ),
        tools = ctx.attr._sg[DefaultInfo].default_runfiles.files,
    )

describe = rule(
    implementation = _describe_impl,
    attrs = {
        "srcs": attr.label_list(allow_files= True, mandatory= True),
        "db": attr.string(mandatory = True),
        "format": attr.string(mandatory = True),
        "out": attr.output(mandatory = True),
        "_sg": attr.label(executable = True, default = "//dev/sg:sg", cfg = "exec"),
    },
)

