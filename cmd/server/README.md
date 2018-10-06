# `sourcegraph/server` release process

The user-facing documentation lives in
[sourcegraph/website](https://github.com/sourcegraph/website). Quick links:

- https://github.com/sourcegraph/website/blob/master/data/docs/server/docs.md
- https://about.sourcegraph.com/docs/server/

This file documents information relevant to developers at Sourcegraph.

## Publishing a new version

#### (0) Update the CHANGELOG

1.  Update [CHANGELOG](../../CHANGELOG.md) on `master` (even if the release will be cut from a different branch).

    1.  If this is a major or minor release, move any `Unreleased changes` under their own section
        for the new `VERSION` you are releasing.
    1.  If this is a patch release, create a new section for the `VERSION` you are releasing and
        selectively move changelog items from `Unreleased changes` to this section.

    **Remove any empty sections in the changelog**.

1.  Commit and `git push` this change directly to upstream `master`.

#### (1) Use the proper release branch

- New major releases (e.g. `v3.0.0`) should be tagged from `master`. Once you have tagged the release, create a long lived branch from that tag for patches to that release (e.g. `3.0`).
- New minor releases (e.g. `v3.1.0`) should be tagged from `master`. Once you have tagged the release, create a long lived branch from that tag for patches to that release (e.g. `3.1`).
- New patch releases (e.g. `v3.1.1`) should be tagged from the existing patch branch for the release being patched (e.g. `3.1`).
  - Commits necessary for a patch release are cherry-picked from `master` into the patch branch
    before tagging a release.
  - Do NOT commit any fixes or features into a patch release that are not yet in
    `master`. Otherwise, you will end up breaking the validity of the changelog, as such orphaned
    updates will be silently regressed in the next major/minor release.

To avoid confusion between tags and branches:

- Tags are always the full semantic version with a leading `v` (e.g. `v2.10.0`)
- Branches are always the dot-separated major/minor versions with no leading `v` (e.g. `2.10`).

#### (2) Create the release candidate

Push your release branch upstream. Remember the commit hash. This is the commit you will tag as the official release if testing goes well.

If you are creating a patch release, you will want to push to `docker-images-patch/server` to create a docker image that you can test.

#### (3) Test

1.  Look in Buildkite to find the image hash (e.g. `sha256:043dce9761dd4b48f1211e7b69c0eeb2a4ee89d5a35f889fbdaea2492fb70f5d`) for the `sourcegraph/server` docker image step ([example](https://buildkite.com/sourcegraph/sourcegraph/builds/18738#eca69bac-2efd-4e99-82bd-99e9edd986f9)) of the release candidate build you just created in step 2.

1.  Run the latest container from a clean state:

    ```bash
    CLEAN=true IMAGE=sourcegraph/server@$IMAGE_HASH ./dev/run-server-image.sh
    ```

1.  Do some manual testing:
    - Create admin account
    - Add a repo
    - Do a search
    - Open a code file and see code intel working
1.  Run the previous minor version (e.g. if releasing 2.9.0 then install 2.8.0) from a clean state:

    ```bash
    CLEAN=true IMAGE=sourcegraph/server:$PREVIOUS_VERSION ./dev/run-server-image.sh
    ```

1.  Do some manual testing to create state that we can ensure persists after the update:
    - Create admin account
    - Add a repo
    - Do a search
    - Open a code file and see code intel working
1.  Stop the old container and run the new container without cleaning the data:

    ```bash
    CLEAN=false IMAGE=sourcegraph/server@$IMAGE_HASH ./dev/run-server-image.sh
    ```

1.  Verify that
    - You don't need to recreate an admin account.
    - The repo you added is still added.
    - Search works.
    - Code intelligence works.

#### (4) Release the `sourcegraph/server` Docker image

```bash
git fetch
git tag v$VERSION ${COMMIT_FROM_STEP_2}
git push --tags
```

#### (5) Update the public documentation

1.  Check out a new branch in the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
1.  Update the old version number in the documentation to be the version number you are releasing [by editing this React component](https://github.com/sourcegraph/website/blob/master/src/components/ServerVersionNumber.tsx).
1.  Copy over the new changelog entries from [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).
1.  Copy the schema from the released git commit to the [sourcegraph/website](https://github.com/sourcegraph/website) repo:
    ```bash
    cd $WEBSITE_REPO
    cp $GOPATH/src/github.com/sourcegraph/sourcegraph/schema/site.schema.json utils/
    ```
1.  Create the PR on the website repository and merge it.

#### (6) Notify existing instances that an update is available

1.  Checkout the `master` branch in the [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) repository.
1.  Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `latestReleaseDockerServerImageBuild` to the
    semver version string of the new version.
1.  Commit and `git push` this change directly to the `master` branch.

`sourcegraph/server` version `VERSION` has been released!

You should also [release the new version for Sourcegraph cluster deployment](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#cutting-a-release).

## Publishing new code intelligence images

1.  Ensure that the latest version of the language server is uploaded
    to `us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION`.
1.  `./cmd/server/release-codeintel.sh $LANG $VERSION` e.g. `./cmd/server/release-codeintel.sh go 16903_2018-06-13_060942e`
