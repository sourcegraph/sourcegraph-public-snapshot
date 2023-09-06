# @jhchabran: I attempted making it work with templating and handling the docker run seperately, but that runs
# gets cached, which is absolutely not what we want to be doing. I tried setting execution requirements to avoid
# that but it still got cached.
#
# So in the end, the following is not very pretty, but it works as intended, so that'll do for now.
WRAP_POSTGRES_UTILS = """
port_offset=$(($$ % 1000))
pg_port=$(($port_offset + 5432 + 1))

echo "---"
env
echo "---"

docker load --input {pg_image}
container_id=$(docker run --rm --detach --platform linux/amd64 -p $pg_port:5432 postgres-12:candidate)
trap "docker kill $container_id" EXIT

# Let's wrap a few postgres utilities
mkdir -p pg_wrapper_bin

## createdb is straightforward, just need to set the PGUSER
echo -e "#!/usr/bin/env bash\ndocker exec -e PGUSER=sg $container_id createdb \\$@" > pg_wrapper_bin/createdb
chmod +x pg_wrapper_bin/createdb

## pg_dump is trickier, sg calls it with the full POSTGRESDSN, so we replace the dynamic port on the fly, to make it work
## because it's run inside the container, which has the default port.
echo -e "#!/usr/bin/env bash\ndocker exec -e PGUSER=sg $container_id pg_dump \\$(echo \\$@ | sed \"s/$pg_port/5432/\")" > pg_wrapper_bin/pg_dump
chmod +x pg_wrapper_bin/pg_dump

export PATH="$(pwd)/pg_wrapper_bin:$PATH"

RETRIES=10
until docker exec $container_id psql -h localhost -U sg -d sg -c "select 1" > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
echo "Waiting for postgres server, $((RETRIES--)) remaining attempts..."
sleep 1
done

export PGUSER=sg
export PGPORT=$pg_port
export PGHOST=localhost
"""

def _migration_impl(ctx):
    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [ctx.outputs.out],
        progress_message = "Running squash migration for %s" % ctx.attr.db,
        command = """{wrap_postgres_script}
        export HOME=$(pwd)
        export SG_FORCE_REPO_ROOT=$(pwd)

        {sg} migration squash-all -skip-teardown -db {db} -f {output_file}
        """.format(
            wrap_postgres_script = WRAP_POSTGRES_UTILS.format(pg_image = ctx.file._pg_image.path),
            sg = ctx.executable._sg.path,
            db = ctx.attr.db,
            output_file = ctx.outputs.out.path,
        ),
        tools = [ctx.attr._sg[DefaultInfo].default_runfiles.files, ctx.file._pg_image],
    )

migration = rule(
    implementation = _migration_impl,
    attrs = {
        "srcs": attr.label_list(allow_files= True, mandatory= True),
        "db": attr.string(mandatory = True),
        "out": attr.output(mandatory = True),
        "_sg": attr.label(executable = True, default = "//dev/sg:sg", cfg = "exec"),
        "_pg_image": attr.label(allow_single_file = True, default = "//docker-images/postgres-12-alpine:image_tarball")
    },
)

def _describe_impl(ctx):
    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [ctx.outputs.out],
        progress_message = "Running describe migration for %s" % ctx.attr.db,
        command = """
        {wrap_postgres_script}

        export HOME=$(pwd)
        export SG_FORCE_REPO_ROOT=$(pwd)

        {sg} migration describe -db {db} --format={format} -force -out {output_file}
        """.format(
            wrap_postgres_script = WRAP_POSTGRES_UTILS.format(pg_image = ctx.file._pg_image.path),
            sg = ctx.executable._sg.path,
            db = ctx.attr.db,
            format = ctx.attr.format,
            output_file = ctx.outputs.out.path
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
        "_pg_image": attr.label(allow_single_file = True, default = "//docker-images/postgres-12-alpine:image_tarball")
    },
)

