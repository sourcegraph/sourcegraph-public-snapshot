# Sourcegraph Server - Docker images

The user facing documentation lives in
[sourcegraph/website](https://github.com/sourcegraph/website). Quick links:

* https://github.com/sourcegraph/website/blob/master/data/docs/server/docs.md
* https://about.sourcegraph.com/docs/server/

This file documents information relevant to developers at Sourcegraph.

## Publishing a new version

This process is quite manual still, since we want to ensure each release is
high quality. As we get used to releasing Sourcegraph Server more and more
parts will be automated.

#### Prepare a PR to the [sourcegraph/website](https://github.com/sourcegraph/website) repository

1. Check out a new branch in the [sourcegraph/website](https://github.com/sourcegraph/website) repository.
2. Ensure documentation is up-to-date with everything listed under the `Coming Soon` section in the [CHANGELOG](../../CHANGELOG.md). Do not edit the `CHANGELOG.md` file yet.
3. Update every old version number in the documentation to be the version number you are releasing. [Use search to do this](https://sourcegraph.sgdev.org/search?q=repo:%5Egithub%5C.com/sourcegraph/website%24+server%5C:2).
4. Regenerate the site settings docs by running the last two commands mentioned under https://github.com/sourcegraph/website#documentation-pages
5. Create the PR on the website repository, but do not merge it yet.

---

1. Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `ProductVersion` to the
   semver version string of the new version.
1. `git push origin -f origin/master:docker-images/server`
1. Wait for the build to complete [buildkite docker-images/server](https://buildkite.com/sourcegraph/sourcegraph/builds?branch=docker-images%2Fserver)
1. `gcloud docker -- pull us.gcr.io/sourcegraph-dev/server:${CI_VERSION}`.
   You can find it on the build output, it should look something like
   `08248_2017-12-14_8dad5ab`.
1. Run through the [https://about.sourcegraph.com/docs/server/], but using the
   image you just pulled instead of the dockerhub image. Do this for both the
   old and new instructions, to ensure we don't make any bad backwards
   incompatible changes. In future this will be more automated.
1. Update `CHANGELOG` and renaming Unreleased to the new `VERSION`.
1. `docker tag us.gcr.io/sourcegraph-dev/server:CI_VERSION sourcegraph/server:VERSION`
1. `docker tag sourcegraph/server:VERSION sourcegraph/server:latest`
1. `docker push sourcegraph/server:VERSION`
1. `docker push sourcegraph/server:latest`
1. Update ../cmd/frontend/internal/app/pkg/updatecheck/handler.go's `latestReleaseBuild` to the
   timestamp and semver version string of the new version.

## Publishing new code intelligence images

1. Ensure that the latest version of the language server is uploaded
   to `us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION`. For most
   languages, this can be accomplished by pushing to the branch
   `xlang-$LANG`. Pull the latest version locally.
1. `docker tag us.gcr.io/sourcegraph-dev/xlang-$LANG:$VERSION sourcegraph/codeintel-$LANG:latest`
1. `docker push sourcegraph/codeintel-$LANG:latest`
