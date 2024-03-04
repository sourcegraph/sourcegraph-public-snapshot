def _wolfi_lockfiles(rctx):
    translates_contents = """
load("@rules_apko//apko:translate_lock.bzl", "translate_apko_lock")

def apko_translate_locks():
"""

    repo_loads = ""
    repo_contents = """
def apko_repositories():
"""

    result = rctx.execute(["ls", str(rctx.workspace_root) + "/wolfi-images"])
    if result.return_code != 0:
        fail("failed to list wolfi-images:", result.stderr)

    for file in result.stdout.split("\n"):
        if not file.endswith(".lock.json"):
            continue

        lockname = file.partition(".")[0].replace("-", "_")

        translates_contents += """
    translate_apko_lock(
        name = "{}_apko_lock",
        lock = "@//wolfi-images:{}",
        visibility = ["//visibility:public"],
    )
""".format(lockname, file)

        repo_loads += """load("@{}_apko_lock//:repositories.bzl", {}_apko_repositories = "apko_repositories")\n""".format(lockname, lockname)
        repo_contents += """    {}_apko_repositories()\n""".format(lockname)

    rctx.file("BUILD.bazel")
    rctx.file("translates.bzl", content = translates_contents)
    rctx.file("repositories.bzl", content = repo_loads + repo_contents)

wolfi_lockfiles = repository_rule(
    implementation = _wolfi_lockfiles,
)
