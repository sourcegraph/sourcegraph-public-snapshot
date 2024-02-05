"""
Bazel aspect to run scip-java against a Java Bazel codebase.

You can optionally commit this file into your git repository, gitignore it, or
just delete it. When you run `scip-java index` in a Bazel codebase, this file
will get re-created and the command will error if the file already exists but with
different contents.

This aspect is needed for scip-java to inspect the structure of the Bazel build
and register actions to index all java_library/java_test/java_binary targets.
The result of running this aspect is that your bazel-bin/ directory will contain
many *.scip (https://github.com/sourcegraph/scip) and
*.semanticdb (https://scalameta.org/docs/semanticdb/specification.html) files.
These files encode information about which symbols are referenced from which
locations in your source code.

This aspect only works on Linux when using the `local` spawn strategy because
the `run_shell` action writes SemanticDB and SCIP files to the provided
--targetroot argument. It should be possible to avoid this requirement
in the future if there's a strong desire to make the aspect work with the
default (sandboxed) spawn strategy.

Use the command below to merge all of these SCIP files into a single index:

    find bazel-bin/ -type f -name '*.scip' | xargs cat > index.scip

Use `src code-intel upload` to upload the unified SCIP file to Sourcegraph:

    npm install -g @sourcegraph/src
    export SRC_ENDPOINT=SOURCEGRAPH_URL
    export SRC_ACCESS_TOKEN=TOKEN_VALUE
    src login # confirm you are correctly authenticated
    src code-intel upload -file=index.scip

Example command to run this aspect directly:

    bazel build //... --spawn_strategy=local  --aspects path/to/scip_java.bzl%scip_java_aspect --output_groups=scip --define=sourceroot=$(pwd) --define=scip_java_binary=$(which scip-java) --define=java_home=$JAVA_HOME

To learn more about aspects: https://bazel.build/extending/aspects
"""

def _scip_java(target, ctx):
    if JavaInfo not in target or not hasattr(ctx.rule.attr, "srcs"):
        return None

    javac_action = None
    for a in target.actions:
        if a.mnemonic == "Javac":
            javac_action = a
            break

    if not javac_action:
        return None

    info = target[JavaInfo]
    compilation = info.compilation_info
    annotations = info.annotation_processing

    source_files = []
    for src in ctx.rule.files.srcs:
        source_files.append(src.path)
    if len(source_files) == 0:
        return None

    classpath = [j.path for j in compilation.compilation_classpath.to_list()]
    bootclasspath = [j.path for j in compilation.boot_classpath]

    processorpath = []
    processors = []
    if annotations and annotations.enabled:
        processorpath += [j.path for j in annotations.processor_classpath.to_list()]
        processors = annotations.processor_classnames

    build_config = struct(**{
        "javaHome": ctx.var["java_home"],
        "classpath": classpath,
        "sourceFiles": source_files,
        "javacOptions": compilation.javac_options,
        "processors": processors,
        "processorpath": processorpath,
        "bootclasspath": bootclasspath,
        "reportWarningOnEmptyIndex": False,
    })
    build_config_path = ctx.actions.declare_file(ctx.label.name + ".scip.json")

    scip_output = ctx.actions.declare_file(ctx.label.name + ".scip")
    targetroot = ctx.actions.declare_directory(ctx.label.name + ".semanticdb")
    ctx.actions.write(
        output = build_config_path,
        content = build_config.to_json(),
    )

    deps = [javac_action.inputs, annotations.processor_classpath]
    ctx.actions.run_shell(
        command = "\"{}\" index --no-cleanup --index-semanticdb.allow-empty-index --cwd \"{}\" --targetroot {} --scip-config \"{}\" --output \"{}\"".format(
            ctx.var["scip_java_binary"],
            ctx.var["sourceroot"],
            targetroot.path,
            build_config_path.path,
            scip_output.path,
        ),
        env = {
            "JAVA_HOME": ctx.var["java_home"],
            "NO_PROGRESS_BAR": "true",
        },
        mnemonic = "ScipJavaIndex",
        inputs = depset([build_config_path], transitive = deps),
        outputs = [scip_output, targetroot],
    )

    return scip_output

def _scip_java_aspect(target, ctx):
    scip = _scip_java(target, ctx)
    if not scip:
        return struct()
    return [OutputGroupInfo(scip = [scip])]

scip_java_aspect = aspect(
    _scip_java_aspect,
)

def _scip_java_impl(ctx):
    output = ctx.attr.compilation[OutputGroupInfo]
    return [
        OutputGroupInfo(scip = output.scip),
        DefaultInfo(files = output.scip),
    ]

scip_java = rule(
    implementation = _scip_java_impl,
    attrs = {"compilation": attr.label(aspects = [scip_java_aspect])},
)
