# How to cache CI artefacts

This guide documents how to cache build artefacts in order to speed up build times in Sourcegraph's [Buildkite CI pipelines](../background-information/ci/index.md#buildkite-pipelines).

> NOTE: Before getting started, we recommend familiarize yourself with [Pipeline development](../background-information/ci/development.md) and [Buildkite infrastructure](../background-information/ci/development.md#buildkite-infrastructure).

## The need for caching

Because [Buildkite agents are stateless](../background-information/ci/development.md#buildkite-infrastructure) and start with a blank slate, this means that all dependencies and binaries have to rebuild on each job. This is the price to pay for complete isolation from one job to the other.

A common strategy to address this problem of having to rebuild everything is to store objects that are commonly reused accross jobs and to download them again rather than rebuilding everything from scratch. 

Cached artefacts *are automatically expired after 30 days* (by an object lifecycle policy on the bucket).

## What and when to cache?

In order to determine what we can cache and when to do it, we need to make sure to cover the following questions:

1. Will it be faster to download a cached version rather than building it?
1. Can we come up with an identifier that represents accurately what version of the cache is needed, before building anything? 
  - Ex: the checksum of `yarn.lock` accurately tells us which version of the dependencies cache needs to be downloaded if it exists. 
1. Is what is being cached, not the end result of that test? 
  - If it were the case, it means we would be caching the result of the test, which is defeating the purpose of a continuous integration process. 

## How to write a step that caches an artefact?

In the [CI pipeline generator](../background-information/ci/development.md), when defining a step you can use the `buildkite.Cache()` function to define what needs to be cached and under which key to store it. 

For example: we want to cache the `node_modules` folder to avoid dowloading again all dependencies for the front-end. 

```go
// Browser extension unit tests
pipeline.AddStep(":jest::chrome: Test browser extension",
  bk.Cache(&buildkite.CacheOptions{
		ID:          "node_modules",
		Key:         "cache-node_modules-{{ checksum 'yarn.lock' }}",
		RestoreKeys: []string{"cache-node_modules-{{ checksum 'yarn.lock' }}"},
		Paths:       []string{"node_modules", "client/extension-api/node_modules", "client/eslint-plugin-sourcegraph/node_modules"},
	})
  bk.Cmd("dev/ci/yarn-test.sh client/browser"),
  bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
```

The important part here are: 

- The `Key` which defines how the cached version is named, so we can find it again on ulterior builds. 
  - It includes a `{{ checkusm 'yarn.lock' }}` segment in its name, which means that any changes in the `yarn.lock` will be reflected in the key name. 
  - It means that if the `yarn.lock` checksum would change because dependencies have changed, it should use a different version of the cached dependencies. If those were to not be present on the cache, it will simply rebuild them and upload the result to the cache. 
- The `RestoreKeys` lists the keys we can use to know if there is a cached version or not available. 99% of the time, that's the same exact thing as the `Key`.
- The `Paths` lists the path to the files that needs to be cached. They **must be within the repository**, not outside.

Please note that from the perspective of the commands ran in that step, it's not possible to know if the cache was hit or missed. So the code using what has been cached must be able to understand it on its own (example, checking for the presence of folders defined in the `Paths` keys). Most build systems will handle that on their own, but if you were to cache test data you'll probably have to handle that yourself. 

## How to make sure that caching is faster?

When a build is finished and successful, you will find an annotation with a link to the trace of that build on HoneyComb. You can compare the timings of your build before and after adding the cache (with a cache hit of course, otherwise you would compare no cache with a cache miss, which will often be longer because there is time spent uploading the cache result). 

## How to purge the cache from a given key?

If you accidently cached incorrect results, the simplest way to purge it is to use `gsutil`:

```
# List everything
gsutil ls -al gs://sourcegraph_buildkite_cache/sourcegraph/sourcegraph/
# Purge the key you want to delete
gsutil rm gs://sourcegraph_buildkite_cache/sourcegraph/sourcegraph/[MY-KEY].tar.gz
```

## When is the cache purged?

The cached is purged every TODO
