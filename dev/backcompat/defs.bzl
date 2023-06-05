"""Database schema backward compatibility definitions.

Clones and patches Sourcegraph git repository under @sourcegraph_back_compat so we can run tests targets
from that particular release against the new database schema.

Flakes can be defined to skip known problematic tests which are either flaky or simply cannot run against
the new schema in case of a breaking change. See //dev/backcompat:flakes.bzl for more details about
how to define them.

The final result is the definition of a @sourcegraph_back_compat target, whose test targets are exactly
the same as back then, but with instead a new schema.

Example: `bazel test @sourcegraph_back_compat//enterprise/internal/batches/...`.

See https://sourcegraph.com/search?q=context:global+dev/backcompat/patches/back_compat_migrations.patch+repo:github.com/sourcegraph/sourcegraph+lang:Go&patternType=standard&sm=0&groupBy=repo
for the command generating the mandatory patch file in CI for these tests to run.

If the patch file were to be missing, a placeholder diff is in place to make it explicit (in the
eventuality of someone running those locally).
"""

load("test_release_version.bzl", "MINIMUM_UPGRADEABLE_VERSION", "MINIMUM_UPGRADEABLE_VERSION_REF")
load("flakes.bzl", "FLAKES")

load("@bazel_gazelle//:deps.bzl", "go_repository")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

# Shell snippet to disable a test on the fly. Needs to be formatted before being used.
#
# OSX ships BSD sed and the GNU sed that is traditionally available on Linux is named gsed instead.
PATCH_GO_TEST = """_sed_binary="sed"
if [ "$(uname)" == "Darwin" ]; then
    _sed_binary="gsed"
fi
$_sed_binary -i "s/func {}/func _{}/g" {}/*.go
"""

# Assemble go test patches, based on the currently defined version.
# See //dev/backcompat:test_release_version.bzl for the version definition.
PATCH_GO_TEST_CMDS = [
    PATCH_GO_TEST.format(test["prefix"], test["prefix"], test["path"], test["prefix"], test["reason"])
    for test in FLAKES[MINIMUM_UPGRADEABLE_VERSION]
]

# Join all individual go test patches into a single shell snippet.
PATCH_ALL_GO_TESTS_CMD = "\n".join(PATCH_GO_TEST_CMDS)

# Replaces all occurences of @com_github_sourcegraph_(scip|conc) by @back_compat_com_github_sourcegraph_(scip|conc).
PATCH_BUILD_FIXES_CMD = """_sed_binary="sed"
if [ "$(uname)" == "Darwin" ]; then
    _sed_binary="gsed"
fi
find . -type f -name "*.bazel" -exec $_sed_binary -i 's|@com_github_sourcegraph_conc|@back_compat_com_github_sourcegraph_conc|g' {} +
find . -type f -name "*.bazel" -exec $_sed_binary -i 's|@com_github_sourcegraph_scip|@back_compat_com_github_sourcegraph_scip|g' {} +
"""

def back_compat_defs():
    # github.com/sourcegraph/scip and github.com/sourcegraph/conc both rely on a few
    # internal libraries from github.com/sourcegraph/sourcegraph/lib and their
    # respective go_repository rules are annoted with build directives for Gazelle
    # that fixes package resolution.
    #
    # When we're cloning the git repository of sourcegraph/sourcegraph on the release version
    # we're testing against, those directives are not working as intended, as they appear to be
    # evaluate target resolution in the global scope instead of @sourcegraph_back_compat, leading
    # to compiling these two packages with the right targets, but linking them against the ones
    # from our root workspace.
    #
    # So to work around that, we introduce two additional go_repository rules, which are the one that
    # @sourcegraph_back_compat declares, but we rewrite the directives to resolve toward
    # @sourcegraph_back_compat explicitly. The version/sum are exactly the same ones as defined
    # by @sourcegraph_back_compat on the commit tagged by the release.
    #
    # And to make this work, we patch on the fly buildfiles in @sourcegraph_back_compat to reference
    # the newly declared repos, by replacing all occurences of @com_github_sourcegraph_scip by
    # @back_compat_com_github_sourcegraph_scip (and same thing for sourcegraph/conc).
    go_repository(
        name = "back_compat_com_github_sourcegraph_scip",
        # This fixes the build for sourcegraph/scip which depends on sourcegraph/sourcegraph/lib but
        # gazelle doesn't know how to resolve those packages from within sourcegraph/scip.
        build_directives = [
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/errors @sourcegraph_back_compat//lib/errors",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol @sourcegraph_back_compat//lib/codeintel/lsif/protocol",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader @sourcegraph_back_compat//lib/codeintel/lsif/protocol/reader",
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/writer @sourcegraph_back_compat//lib/codeintel/lsif/protocol/writer",
        ],
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sourcegraph/scip",
        sum = "h1:fWPxLkDObzzKTGe9vb6wpzK0FYkwcfSxmxUBvAOc8aw=", # Need to be manually updated when bumping the back compat release target.
        version = "v0.2.4-0.20221213205653-aa0e511dcfef", # Need to be manually updated when bumping the back compat release target.
    )

    # Same logic for this repository.
    go_repository(
        name = "back_compat_com_github_sourcegraph_conc",
        build_directives = [
            "gazelle:resolve go github.com/sourcegraph/sourcegraph/lib/errors @sourcegraph_back_compat//lib/errors",
        ],
        build_file_proto_mode = "disable_global",
        importpath = "github.com/sourcegraph/conc",
        sum = "h1:96VpOCAtXDCQ8Oycz0ftHqdPyMi8w12ltN4L2noYg7s=", # Need to be manually updated when bumping the back compat release target.
        version = "v0.2.0", # Need to be manually updated when bumping the back compat release target.
    )


    # Now that we have declared a replacement for the two problematic go packages that
    # @sourcegraph_back_compat depends on, we can define the repository itself. Because it
    # comes with its Bazel rules (logical, that's just the current repository but with a different
    # commit), we can simply use git_repository to fetch it and apply patches on the fly to
    # inject migrations and build fixes.
    git_repository(
        name = "sourcegraph_back_compat",
        remote = "https://github.com/sourcegraph/sourcegraph.git",
        patches = ["//dev/backcompat/patches:back_compat_migrations.patch"],
        patch_args = ["-p1"],
        commit = MINIMUM_UPGRADEABLE_VERSION_REF,
        patch_cmds = [
            # webpack rules are complaining about a missing entry point, which is irrelevant as we're
            # simply running go tests only. Therefore, we can simply drop the client folder.
            #
            # Because the target release at the time of writing this comment is 5.0.0 which was
            # before we fully switched to Bazel, it's not exactly in a buildable state, and we're using
            # a backported fix  to make it buildable. That fix is merely about running `bazel configure`
            # and dropping the client folder.
            #
            # "rm -Rf client",
            PATCH_ALL_GO_TESTS_CMD,
            PATCH_BUILD_FIXES_CMD,
            # Seems to be affecting the root workspace somehow.
            # TODO(JH) Look into bzlmod.
            "rm .bazelversion",
        ],
    )
