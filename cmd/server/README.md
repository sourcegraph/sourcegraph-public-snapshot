# Sourcegraph Server - Docker images

The user facing documentation lives in
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
1.  Update every old version number in the documentation to be the version number you are releasing. [Use search to do this](https://sourcegraph.sgdev.org/search?q=repo:%5Egithub%5C.com/sourcegraph/website%24+server%5C:2).
1.  Regenerate the site settings docs by running the last two commands mentioned under https://github.com/sourcegraph/website#documentation-pages
1.  Create the PR on the website repository, but do not merge it yet.

#### (2) Build a Sourcegraph Server Docker image

1.  Checkout the `master` branch in the [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) repository.
1.  Update `../cmd/frontend/internal/app/pkg/updatecheck/handler.go`'s `ProductVersion` to the
    semver version string of the new version (**DO NOT update `latestReleaseVersion` yet**).
1.  Commit and `git push` this change directly to the `master` branch.
1.  `git push origin -f origin/master:docker-images/server`

#### (3) Test the Sourcegraph Server Docker image

1.  Wait for the build to complete [buildkite docker-images/server](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=docker-images%2Fserver)
1.  `gcloud docker -- pull us.gcr.io/sourcegraph-dev/server:${CI_VERSION}`.
    You can find it on the build output CI page in the last Docker build step, it should look something like
    `08248_2017-12-14_8dad5ab`. Important: The version number must come from the [docker-images/server](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=docker-images%2Fserver) branch, not `master`. Make sure you are on the right buildkite page.
1.  Run through the [https://about.sourcegraph.com/docs/server/], but using the
    image you just pulled instead of the dockerhub image. Do this for both the
    old and new instructions, to ensure we don't make any bad backwards
    incompatible changes. In future this will be more automated. The `docker run` command you will use will look like:

```bash
gcloud docker -- run \
 --publish 7080:7080 --rm \
 --volume $HOME/.sourcegraph/config:/etc/sourcegraph \
 --volume $HOME/.sourcegraph/data:/var/opt/sourcegraph \
 us.gcr.io/sourcegraph-dev/server:08248_2017-12-14_8dad5ab
```

At this point if you've discovered an issue and plan to stop the release, you should inform everyone that there is an issue and not to do a release temporarily (e.g. in #dev-announce). You are responsible for completing the next release following these steps where you left off, or stating clearly to others where you left off in this process so that someone else can confidently continue.

#### (4) Completing the release

It is important that the following steps be ran closely together, otherwise we will end up in an incomplete release state. DO NOT pause or otherwise stop once you begin the following steps.

1.  `docker tag us.gcr.io/sourcegraph-dev/server:CI_VERSION sourcegraph/server:VERSION`
1.  `docker tag sourcegraph/server:VERSION sourcegraph/server:latest`
1.  `docker push sourcegraph/server:VERSION`
1.  `docker push sourcegraph/server:latest`
1.  Merge the PR that you previously prepared to the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
1.  Checkout the `master` branch in the [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) repository.
1.  Update [CHANGELOG](../../CHANGELOG.md) and move any `Unreleased changes` under their own section for the new `VERSION` you have just released.
1.  Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `latestReleaseBuild` to the
    semver version string of the new version.
1.  Commit and `git push` this change directly to the `master` branch.

You are done! Sourcegraph Server version `VERSION` has been released!

## Publishing new code intelligence images

1.  Ensure that the latest version of the language server is uploaded
    to `us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION`. For most
    languages, this can be accomplished by pushing to the branch
    `xlang-$LANG`. Pull the latest version locally.
1.  `docker tag us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION sourcegraph/codeintel-$LANG:latest`
1.  `docker push sourcegraph/codeintel-$LANG:latest`
