# Release tooling

## Usage

### Generating a database schemas tarball

Generating a database schemas tarball is achieved by downloading all known schemas plus the current schema. There are two options for the current version database schemas: either we are cutting a new release and we need to inject the current one, or we are regenerating the tarball to fix a problem.

To control which approach we take, we use the second parameter of the following command:

```
bazel run //tools/release:generate_schemas_archive -- vX.Y.Z [ACTION] $HOME/[PATH-TO-YOUR-REPO]
```

If ran with `fetch-current-schemas`, the script will ensure that the schemas archive in the bucket correctly
contains the given version database schemas. It will also prompt the user for confirmation if the associated
tarball with that version exists in the bucket.

If ran with `inject-current-schemas`, the script will ensure that the schemas archive in the bucket doesn't
contain the schemas for the new version and will instead create them by injecting the `internal/database/schemas*.json` schemas into the tarball, properly renamed to the expected convention.

Finally, in both cases, the tarball will be uploaded in the bucket, and the third party dependency, located in
`tools/release/schema_deps.bzl` will be updated accordingly, allowing builds past that point to use those schemas.

### Uploading the current database schemas

Once a release is considered to be correct (upcoming in RFC 795) the release tooling runs another command
to store the current database schemas in the bucket, under the `schemas` folder, to capture how the database
looks at that point.

This enables to build migrator binaries that will be able to use that particular release as a migration point.

```
bazel run //tools/release:upload_current_schemas -- vX.Y.Z
```

The script will ensure that there are no existing database schemas for that version before uploading anything. This way
we prevent accidentally breaking previously generated database schemas.

## Database schemas

Database schemas are necessary for Multi-Version Upgrades, so we need to populate
them when building and cutting new releases.

The following diagram provides an overview of how it works.

```
 ┌─────────────────────────────────────────┐
 │ GCS Bucket                              │
 │                                         │
 │  ┌───────────────────────────────────┐  │
 │  │ schemas/                          │  │
 │  │  v3.1.4-(...).schema.json         │  │
 │  │  ...                              │  │
 │  │  v5.2.1234-(...).schema.json ◄────┼──┼───────  Uploaded on a successful
 │  │                                   │  │         new release build
 │  │                                   │  │
 │  │                                   │  │
 │  └───────────────────────────────────┘  │
 │                                         │
 │  ┌───────────────────────────────────┐  │
 │  │ dist/                             │  │
 │  │  schemas-v5.2.1093.tar.gz         │  │
 │  │  schemas-v5.2.1234.tar.gz ◄───────┼──┼──────  Uploaded at the beginning of
 │  │             ▲                     │  │        of a new release build.
 │  │             │                     │  │
 │  │             │                     │  │        Release build automatically
 │  │             │                     │  │        update the Bazel reference
 │  │             │                     │  │        to that file.
 │  └─────────────┼─────────────────────┘  │
 │                │                        │        Gets deleted if the release
 │                │                        │        build fails.
 │                │                        │
 └────────────────┼────────────────────────┘
                  │
                  │
                  │
              referenced by Bazel and used to
              populate the schemas when building
              the cmd/migrator Docker container.
```

There are two distinct scenarios:

1. Normal builds
2. Release builds

When doing a normal build, we simply use the schemas tarball that has been previously
set by the last release build. It contains all knowns schema descriptions that existed
at that time.

Now when doing a release build, we need to refresh the schema descriptions, because patch releases
might have been publicly released, meaning those schemas now exist in the wild, on customer deployments
or cloud.

Let's use a concrete example:

1. t=0 5.1.0 has been released publicly
   - `main` branch is now considered to be 5.2.0
   - `5.1` branch is the target for PRs for backports and bug fixes.
1. t=10 5.1.2222 has been released publicly
   - `5.1` branch is from where this release was cut.
1. t=20 5.2.0 has been released publicly
   - `main` branch is now considered to be 5.3.0
   - `5.2` branch is the target for PRs for backports and bug fixes.
1. t=30 5.1.3333 has been released publicly
   - `5.1` branch is from where this release was cut.

So with that scenario, when 5.1.3333 has been released, we introduced a new version that the _migrator_ must be aware of, on both `main` and the `5.1` branch. Previously, this required us to make a PR to port to main, the 5.1 branch references
to the new 5.1.3333 schemas. See [this PR for a real example](https://github.com/sourcegraph/sourcegraph/pull/56405/files#diff-38f26d6e9cb950b24ced060cd86effd4363b313d880d1afad1850887eabaf238R79).

Failing to do this, would mean the _migrator_ we're going to ship on the next 5.2 release will not cover the migration path from 5.1.3333 when doing multi-version upgrades.

Ultimately, this means that when a release cut is at any point in time, you need to be aware of all previously released
version, even if they were released on the previous minor release. Instead of having to remember to enact those changes,
we can take a different approach.

The GCS bucket has two folders: `schemas/` and `dist/`. `schemas/` is the source of truth for all known schemas up until now, regardless of the current version. Whenever a new release is cut, the new schemas are added in that folder. Therefore, when doing the next release cut, we will use that folder to populate all the schemas that _migrator_ needs to be aware of, without having to make any manual change in the code.

Now, when building the _migrator_, we can't directly use the GCS bucket. Bazel wants a deterministic set of inputs and "all content from the bucket" is not deterministic.
To satisfy Bazel, we need a fixed input, checksumed, to guarantee that the build is stable. So when we're creating a new release, we simply regenerate that
tarball based on the schemas we find in the bucket, under `schemas/` and upload it under `dist/`.

Step by step process (work-in-progress):

1. We want to create a new release, which is materialized by a pull-request automatically created by `sg`
1. `sg release create ...` runs `bazel run //tools/release:generate_schemas_archive -- v5.3.4444`
1. it fetches all schemas whose version are below 5.3.4444
1. it copies the current `schema.*.json` files to `v5.3.4444-internal_database.schema.*.json`, to match the convention of the other schemas.
1. it creates a tarball named `schemas-v5.3.4444.tar.gz`
1. it uploads it under the `dist/` folder in the bucket.
1. it updates `tools/release/schema_deps.bzl` with the new tarball URL and its checksum.
1. CI builds the new release.
1. At the end of the build:

- If green
  - the schemas `v5.3.4444-internal_database.schema.*.json` are uploaded to the `schemas/` folder.
- If red
  - the schemas `v5.3.4444-internal_database.schema.*.json` are _NOT_ uploaded to the `schemas/` folder.
    - that's because if the release build failed, it means that it never existed, so there is no need to capture its existence as nobody will migrate from that version number.
  - the `schemas-v5.3.4444.tar.gz` tarball is removed from the `dist/` folder in the bucket. This is perfectly fine that there is no revision apart from the current PR that references it.

1. PR driving the release build is merged back in the base branch

- the updated buildfiles will now use that uploaded `schemas-v5.3.4444.tar.gz` tarball from now on, eliminating the need to fetch anything from GCS apart the tarball (until it's cached by Bazel).

## Q&A

> What happens if two release builds are built at the same time?

If two builds are going on at the same time, they won't interfere with each other, because the only artifacts that can be removed without notice are the schemas tarballs, which are
only referenced by each individual release build. As for the schemas, the only time they get created is when the internal release is finally green and ready to be merged. If one of the two builds end
up referencing the schemas from the other, it means they didn't happen at the same time, but instead that they happened sequentially. That's because GCS is guaranteeing us that file uploads are
transactional, i.e it's not possible to list a file until it's fully uploaded.

> What happens if a release build fails. Can it mess with ulterior release builds?

It cannot, because the only time the schemas are finally added to `schemas/` is when the release build succeeds. This is why when we're regenerating the tarball, we are fetching
all the known schemas _and_ adding the new one from the source tree at that point. Had we uploaded the new schemas at the beginning of the build instead, to then fetch everything to
build the tarball, including the new one, we would have had the problem.

> How do we ensure that the `schema.*.json` in the source, at the revision we're cutting the release are correct?

This is covered by Bazel. These files are generated through `bazel run //dev:write_all_generated` which comes with automatically generated `diff_test` rules, which are comparing
the files on disk, with the files it would generate. Therefore, if someone pushes code without updating the current schemas in the code, Bazel will fail the build. And if on that
precise commit we would try to cut a release, that same exact test would run again and fail.

Therefore, we can safely use the current schemas when cutting a release.

> What happens if the _migrator_ is built with newer schemas, like 5.1.3333 that contains schemas for 5.2.4444?

The script that populates the schemas, when regenerating the tarball, in that case, would exclude all schemas above 5.1.X, so it won't happen.

> How does this work until we revamp the release process to match RFC 795?

The initial state has been created manually on the bucket, and there won't be any issues until we create a new release, which is at the time of writing this doc
a periodic event, manually driven by the release captain. We can keep building the patch releases for 5.2.X with the old method, we just have to upload the
new schemas to the bucket to ensure that the next release from `main`, i.e 5.3.X will be correct.

> How is that better than the previous flow?

- Before
  - Cutting a new release
    - Required to port the new schemas to `main` manually on each release.
    - Required Bazel to perform 280 individual HTTP requests sequentially to GitHub and GCS to fetch the schemas.
  - When building normally
    - Schemas are fully cached if the Bazel cache is warm. Otherwise, we go back to the 280 requests.
- After
  - Cutting a new release
    - Schemas are always up to date when cutting a new release. No need to port changes to other release branches or `main`.
    - Schemas are downloaded concurrently - only takes a few second to grab all of them.
  - When building normally
    - Schemas are cached if the Bazel cache is warm. Otherwise, we download a single tarball of a few mbs.

> How do I see which schemas where used to build the _migrator_ container.

`tar tf $(bazel cquery //cmd/migrator:tar_schema_descriptions --output=files)` will show the content the container layer used
to inject the schemas in the final _migrator_ container image.
