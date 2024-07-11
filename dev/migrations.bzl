# Because we have a bunch of ${} in there, which clash with the interpolation for "".format(...),
# it's simpler to have this in a var and inject it as is, rather than having to escape everything.
CMD_PREAMBLE = """set -e

export HOME=$(pwd)
export SG_FORCE_REPO_ROOT=$(pwd)

if [ -z "$PGUSER" ]; then
    export PGUSER="sourcegraph"
fi

if [ -z "$CODEINTEL_PGUSER" ]; then
    export CODEINTEL_PGUSER="$PGUSER"
fi

if [ -z "$CODEINSIGHTS_PGUSER" ]; then
    export CODEINSIGHTS_PGUSER="$PGUSER"
fi

set +e -x
test -d "$PGHOST"
IS_UNIX_POSTGRES=$?
set -ex

if [ $IS_UNIX_POSTGRES -eq 0 ]; then
    if [ -z "$PGDATASOURCE" ]; then
        echo "\\$PGDATASOURCE expected to be set when \\$PGHOST points to the filesystem."
        exit 1
    fi
    PGDATASOURCE_BASE=$PGDATASOURCE
    export PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-frontend"
    export CODEINTEL_PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-codeintel"
    export CODEINSIGHTS_PGDATASOURCE="${PGDATASOURCE_BASE}&dbname=sg-squasher-codeinsights"
else
    export PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-frontend"
    export CODEINTEL_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeintel"
    export CODEINSIGHTS_PGDATASOURCE="postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/sg-squasher-codeinsights"
fi
"""

# Dumping the schema requires running the squash operation first, as it reuses the database, so we do all of those operations
# in a single step.
def _generate_schemas_impl(ctx):
    pgutils_path = ctx.attr._pg_utils[DefaultInfo].files.to_list()[0].path.rpartition("/")[0]
    runfiles = depset(direct = ctx.attr._sg[DefaultInfo].default_runfiles.files.to_list() + ctx.attr._pg_utils[DefaultInfo].default_runfiles.files.to_list())

    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        tools = runfiles,
        outputs = [
            ctx.outputs.out_frontend_squash,
            ctx.outputs.out_codeintel_squash,
            ctx.outputs.out_codeinsights_squash,
            ctx.outputs.out_frontend_schema,
            ctx.outputs.out_codeintel_schema,
            ctx.outputs.out_codeinsights_schema,
            ctx.outputs.out_frontend_schema_md,
            ctx.outputs.out_codeintel_schema_md,
            ctx.outputs.out_codeinsights_schema_md,
        ],
        progress_message = "Running sg migration ...",
        use_default_shell_env = True,
        execution_requirements = {"requires-network": "1"},
        env = {
            # needed because of https://github.com/golang/go/issues/53962
            "GODEBUG": "execerrdot=0",
            # blank out PATH so that we don't pick up host binaries if we end up using more than what
            # //dev:pg_utils filegroup provides.
            "PATH": "",
        },
        command = """{cmd_preamble}

        export PATH="{pgutils_path}"

        trap "dropdb --if-exists sg-squasher-frontend && echo 'temp db sg-squasher-frontend dropped'" EXIT
        trap "dropdb --if-exists sg-squasher-codeintel && echo 'temp db sg-squasher-codeintel dropped'" EXIT
        trap "dropdb --if-exists sg-squasher-codeinsights && echo 'temp db sg-squasher-codeinsights dropped'" EXIT

        {sg} migration squash-all -skip-teardown -db frontend -f {out_frontend_squash}
        {sg} migration squash-all -skip-teardown -db codeintel -f {out_codeintel_squash}
        {sg} migration squash-all -skip-teardown -db codeinsights -f {out_codeinsights_squash}

        {sg} migration describe -db frontend --format=json -force -out {out_frontend_schema}
        {sg} migration describe -db codeintel --format=json -force -out {out_codeintel_schema}
        {sg} migration describe -db codeinsights --format=json -force -out {out_codeinsights_schema}

        {sg} migration describe -db frontend --format=psql -force -out {out_frontend_schema_md}
        {sg} migration describe -db codeintel --format=psql -force -out {out_codeintel_schema_md}
        {sg} migration describe -db codeinsights --format=psql -force -out {out_codeinsights_schema_md}

        """.format(
            cmd_preamble = CMD_PREAMBLE,
            sg = ctx.executable._sg.path,
            pgutils_path = pgutils_path,
            out_frontend_squash = ctx.outputs.out_frontend_squash.path,
            out_codeintel_squash = ctx.outputs.out_codeintel_squash.path,
            out_codeinsights_squash = ctx.outputs.out_codeinsights_squash.path,
            out_frontend_schema = ctx.outputs.out_frontend_schema.path,
            out_codeintel_schema = ctx.outputs.out_codeintel_schema.path,
            out_codeinsights_schema = ctx.outputs.out_codeinsights_schema.path,
            out_frontend_schema_md = ctx.outputs.out_frontend_schema_md.path,
            out_codeintel_schema_md = ctx.outputs.out_codeintel_schema_md.path,
            out_codeinsights_schema_md = ctx.outputs.out_codeinsights_schema_md.path,
        ),
    )

    return [
        DefaultInfo(
            files = depset([
                ctx.outputs.out_frontend_squash,
                ctx.outputs.out_codeintel_squash,
                ctx.outputs.out_codeinsights_squash,
                ctx.outputs.out_frontend_schema,
                ctx.outputs.out_codeintel_schema,
                ctx.outputs.out_codeinsights_schema,
                ctx.outputs.out_frontend_schema_md,
                ctx.outputs.out_codeintel_schema_md,
                ctx.outputs.out_codeinsights_schema_md,
            ]),
        ),
        OutputGroupInfo(
            frontend_squash = depset([ctx.outputs.out_frontend_squash]),
            codeintel_squash = depset([ctx.outputs.out_codeintel_squash]),
            codeinsights_squash = depset([ctx.outputs.out_codeinsights_squash]),
            schemas = depset([
                ctx.outputs.out_frontend_schema,
                ctx.outputs.out_codeintel_schema,
                ctx.outputs.out_codeinsights_schema,
                ctx.outputs.out_frontend_schema_md,
                ctx.outputs.out_codeintel_schema_md,
                ctx.outputs.out_codeinsights_schema_md,
            ]),
        ),
    ]

generate_schemas = rule(
    implementation = _generate_schemas_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "out_frontend_squash": attr.output(mandatory = True),
        "out_codeintel_squash": attr.output(mandatory = True),
        "out_codeinsights_squash": attr.output(mandatory = True),
        "out_frontend_schema": attr.output(mandatory = True),
        "out_codeintel_schema": attr.output(mandatory = True),
        "out_codeinsights_schema": attr.output(mandatory = True),
        "out_frontend_schema_md": attr.output(mandatory = True),
        "out_codeintel_schema_md": attr.output(mandatory = True),
        "out_codeinsights_schema_md": attr.output(mandatory = True),
        "_sg": attr.label(executable = True, default = "//dev/sg:sg", cfg = "exec"),
        "_pg_utils": attr.label(default = ":pg_utils", cfg = "exec"),
    },
)
