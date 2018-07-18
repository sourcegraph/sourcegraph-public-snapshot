# Sourcegraph Server release process

The user-facing documentation lives in
[sourcegraph/website](https://github.com/sourcegraph/website). Quick links:

* https://github.com/sourcegraph/website/blob/master/data/docs/server/docs.md
* https://about.sourcegraph.com/docs/server/

This file documents information relevant to developers at Sourcegraph.

## Publishing a new version

This process is quite manual still, since we want to ensure each release is
high quality. As we get used to releasing Sourcegraph Server more and more
parts will be automated. You will need to complete four main steps.

#### (1) Prepare a PR to the [sourcegraph/website](https://github.com/sourcegraph/website) repository

1.  Check out a new branch in the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
1.  Ensure documentation is up-to-date with everything listed under the `Unreleased changes` section in the [CHANGELOG](../../CHANGELOG.md). Do not edit the `CHANGELOG.md` file yet.
1.  Update the old version number in the documentation to be the version number you are releasing [by editing this React component](https://github.com/sourcegraph/website/blob/master/src/components/ServerVersionNumber.tsx).
1.  Regenerate the site settings docs by running the last two commands mentioned under https://github.com/sourcegraph/website#documentation-pages
1.  Create the PR on the website repository, but do not merge it yet.

#### (2) Test

1.  Run the latest container from a clean state:

    ```
    CLEAN=true ./dev/run-server-image.sh
    ```

1.  Do some manual testing:
    * Create admin account
    * Add a repo
    * Do a search
    * Open a code file and see code intel working
1.  Run the previous minor version (e.g. if releasing 2.9.0 then install 2.8.0) from a clean state:

    ```
    CLEAN=true IMAGE=sourcegraph/server:X.Y.0 ./dev/run-server-image.sh
    ```

1.  Do some manual testing to create state that we can ensure persists after the update:
    * Create admin account
    * Add a repo
    * Do a search
    * Open a code file and see code intel working
1.  Stop the old container and run the new container without cleaning the data:

    ```
    CLEAN=false ./dev/run-server-image.sh
    ```

1.  Verify that
    * You don't need to recreate an admin account.
    * The repo you added is still added.
    * Search works.
    * Code intelligence works.

#### (3) Release the Sourcegraph Server Docker image

```
git fetch
git tag vX.Y.Z origin/master
git push --tags
```

#### (4) Update the public documentation

1.  Merge the PR that you previously prepared to the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
1.  Checkout the `master` branch in the [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) repository.
1.  Update [CHANGELOG](../../CHANGELOG.md) and move any `Unreleased changes` under their own section for the new `VERSION` you have just released.
1.  Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `latestReleaseServerBuild` to the
    semver version string of the new version.
1.  Commit and `git push` this change directly to the `master` branch.

You are done! Sourcegraph Server version `VERSION` has been released!

## Publishing new code intelligence images

1.  Ensure that the latest version of the language server is uploaded
    to `us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION`.

    * Java: `sourcegraph/sourcegraph/cmd/xlang-java/build-and-upload.sh`
    * Java (skinny): `sourcegraph/sourcegraph/cmd/xlang-java-skinny/build-and-upload.sh`
    * JS/TS: Automatically build and uploaded in CI when you commit to master in `sourcegraph/javascript-typescript-buildserver`
    * PHP: Automatically build and uploaded in CI when you commit to master in `sourcegraph/php-buildserver`
    * Python: `git push origin master:docker-images/xlang-python`
    * Go: `git push origin master:docker-images/xlang-go`

1.  `./cmd/server/release-codeintel.sh $LANG $VERSION` e.g. `./cmd/server/release-codeintel.sh go 16903_2018-06-13_060942e`
