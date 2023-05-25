"""`js_binary` helper functions. Copied from rules_js internals.
"""

load("@aspect_rules_js//js:providers.bzl", "JsInfo")
load("@aspect_rules_js//npm:providers.bzl", "NpmPackageStoreInfo")

def gather_files_from_js_providers(
        targets,
        include_sources,
        include_transitive_sources,
        include_declarations,
        include_npm_linked_packages):
    """Gathers files from JsInfo and NpmPackageStoreInfo providers.

    Args:
        targets: list of target to gather from
        include_sources: see js_filegroup documentation
        include_transitive_sources: see js_filegroup documentation
        include_declarations: see js_filegroup documentation
        include_npm_linked_packages: see js_filegroup documentation

    Returns:
        A depset of files
    """

    files_depsets = []

    files_depsets.extend([
        target[DefaultInfo].default_runfiles.files
        for target in targets
        if DefaultInfo in target and hasattr(target[DefaultInfo], "default_runfiles")
    ])

    if include_sources:
        files_depsets.extend([
            target[JsInfo].sources
            for target in targets
            if JsInfo in target and hasattr(target[JsInfo], "sources")
        ])

    if include_transitive_sources:
        files_depsets.extend([
            target[JsInfo].transitive_sources
            for target in targets
            if JsInfo in target and hasattr(target[JsInfo], "transitive_sources")
        ])

    if include_declarations:
        files_depsets.extend([
            target[JsInfo].transitive_declarations
            for target in targets
            if JsInfo in target and hasattr(target[JsInfo], "transitive_declarations")
        ])

    if include_npm_linked_packages:
        files_depsets.extend([
            target[JsInfo].transitive_npm_linked_package_files
            for target in targets
            if JsInfo in target and hasattr(target[JsInfo], "transitive_npm_linked_package_files")
        ])
        files_depsets.extend([
            target[NpmPackageStoreInfo].transitive_files
            for target in targets
            if NpmPackageStoreInfo in target and hasattr(target[NpmPackageStoreInfo], "transitive_files")
        ])

    # print(files_depsets)

    return depset([], transitive = files_depsets)

def gather_runfiles(ctx, sources, data, deps):
    """Creates a runfiles object containing files in `sources`, default outputs from `data` and transitive runfiles from `data` & `deps`.

    As a defense in depth against `data` & `deps` targets not supplying all required runfiles, also
    gathers the transitive sources & transitive npm linked packages from the `JsInfo` &
    `NpmPackageStoreInfo` providers of `data` & `deps` targets.

    See https://bazel.build/extending/rules#runfiles for more info on providing runfiles in build rules.

    Args:
        ctx: the rule context

        sources: list or depset of files which should be included in runfiles

        data: list of data targets; default outputs and transitive runfiles are gather from these targets

            See https://bazel.build/reference/be/common-definitions#typical.data and
            https://bazel.build/concepts/dependencies#data-dependencies for more info and guidance
            on common usage of the `data` attribute in build rules.

        deps: list of dependency targets; only transitive runfiles are gather from these targets

    Returns:
        A [runfiles](https://bazel.build/rules/lib/runfiles) object created with [ctx.runfiles](https://bazel.build/rules/lib/ctx#runfiles).
    """

    # Includes sources
    if type(sources) == "list":
        sources = depset(sources)
    transitive_files_depsets = [sources]

    # Gather the default outputs of data targets
    transitive_files_depsets.extend([
        target[DefaultInfo].files
        for target in data
    ])

    # Gather the transitive sources & transitive npm linked packages from the JsInfo &
    # NpmPackageStoreInfo providers of data & deps targets.
    transitive_files_depsets.append(gather_files_from_js_providers(
        targets = data + deps,
        include_sources = True,
        include_transitive_sources = True,
        include_declarations = False,
        include_npm_linked_packages = True,
    ))

    # Merge the above with the transitive runfiles of data & deps.
    runfiles = ctx.runfiles(
        transitive_files = depset(transitive = transitive_files_depsets),
    ).merge_all([
        target[DefaultInfo].default_runfiles
        for target in data + deps
    ])

    return runfiles
