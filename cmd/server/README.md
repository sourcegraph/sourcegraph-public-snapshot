# Sourcegraph Server release process

The user-facing documentation lives in
[sourcegraph/website](https://github.com/sourcegraph/website). Quick links:

- https://github.com/sourcegraph/website/blob/master/data/docs/server/docs.md
- https://about.sourcegraph.com/docs/server/

This file documents information relevant to developers at Sourcegraph.

## Publishing a new version

This process is quite manual still, since we want to ensure each release is
high quality. As we get used to releasing Sourcegraph Server more and more
parts will be automated. You will need to complete four main steps.

#### (0) Check out the proper branch

New major/minor releases should be cut from `master`.

Patch releases should be cut from the branch that corresponds to the major/minor version of the
patch, e.g., `2.9` or `2.10`. This will typically mean cherry-picking commits from `master` onto the
version branch. If no such branch yet exists, one should be created off the tag of the first release
with that major/minor version, e.g., `v2.9.0` or `v2.10.0`.

Check out the appropriate branch and proceed with the next step.

#### (1) Create the release candidate

1.  Update [CHANGELOG](../../CHANGELOG.md) and move any `Unreleased changes` under their own section for the new `VERSION` you are about to release.
2.  Commit and push this change. Remember the commit hash. This is the commit you will tag as the official release if testing goes well.

#### (2) Test

1.  Look in Buildkite to find the image hash (e.g. `sha256:043dce9761dd4b48f1211e7b69c0eeb2a4ee89d5a35f889fbdaea2492fb70f5d`) for the `sourcegraph/server` docker image step ([example](https://buildkite.com/sourcegraph/sourcegraph/builds/18738#eca69bac-2efd-4e99-82bd-99e9edd986f9)) of the build you just created in (1.2).

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

#### (3) Release the Sourcegraph Server Docker image

```bash
git fetch
git tag v$VERSION ${COMMIT_FROM_STEP_1.2}
git push --tags
```

#### (4) Update the public documentation

1.  Check out a new branch in the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
1.  Update the old version number in the documentation to be the version number you are releasing [by editing this React component](https://github.com/sourcegraph/website/blob/master/src/components/ServerVersionNumber.tsx).
1.  Copy over the new changelog entries from [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).
1.  Copy the schema from the released git commit to the [sourcegraph/website](https://github.com/sourcegraph/website) repo:
    ```bash
    cd $WEBSITE_REPO
    cp $GOPATH/src/github.com/sourcegraph/sourcegraph/schema/site.schema.json utils/
    ```
1.  Create the PR on the website repository and merge it.

#### (5) Notify existing instances that an update is available

1.  Checkout the `master` branch in the [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) repository.
1.  Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `latestReleaseServerBuild` to the
    semver version string of the new version.
1.  Commit and `git push` this change directly to the `master` branch.

Sourcegraph Server version `VERSION` has been released!

You should also [release Data Center](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/README.dev.md#cutting-a-release).

## Publishing new code intelligence images

1.  Ensure that the latest version of the language server is uploaded
    to `us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION`.

    - Java: `sourcegraph/sourcegraph/cmd/xlang-java/build-and-upload.sh`
    - Java (skinny): `sourcegraph/sourcegraph/cmd/xlang-java-skinny/build-and-upload.sh`
    - JS/TS: Automatically build and uploaded in CI when you commit to master in `sourcegraph/javascript-typescript-buildserver`
    - PHP: Automatically build and uploaded in CI when you commit to master in `sourcegraph/php-buildserver`
    - Python: `git push origin master:docker-images/xlang-python`
    - Go: `git push origin master:docker-images/xlang-go`

1.  `./cmd/server/release-codeintel.sh $LANG $VERSION` e.g. `./cmd/server/release-codeintel.sh go 16903_2018-06-13_060942e`
