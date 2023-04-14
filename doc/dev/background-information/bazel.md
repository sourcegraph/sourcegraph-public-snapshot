# Bazel

Sourcegraph is currently migrating to Bazel as its build system and this page is targeted for early adopters which are helping the [#job-fair-bazel](https://sourcegraph.slack.com/archives/C03LUEB7TJS) team to test their work.

## 📅 Timeline

- 2023-04-03 Bazel steps are required to pass on all PR builds.
  - Run `bazel run :update-gazelle-repos` if you changed the `go.mod`.
  - Run `bazel configure` after making all your changes in the PR and commit the updated/added `BUILD.bazel`
  - Add `[no-bazel]` in your commit description if you want to bypass Bazel. 
    - ⚠️  see [what to do after using `[no-bazel]`](#i-just-used-no-bazel-to-merge-my-pr)
- 2023-04-06 Bazel steps are required to pass on main.
  - Run `bazel run :update-gazelle-repos` if you changed the `go.mod`.
  - Run `bazel configure` after making all your changes in the PR and commit the updated/added `BUILD.bazel`
  - Add `[no-bazel]` in your commit description if you want to bypass Bazel.
    - ⚠️  see [what to do after using `[no-bazel]`](#i-just-used-no-bazel-to-merge-my-pr)
- 2023-04-10 You cannot opt out of Bazel anymore

The above timeline affects the following CI jobs:

- Go unit tests.
- Client unit tests (Jest).
  - with the exception of Cody, which has not been ported yet to build with Bazel.

Other jobs will follow-up shortly, but as long as you're following instructions mentioned in that doc (TL;DR run `bazel configure`) this won't change anything for you.

You may find the build and tests to be slow at first, either locally or in CI. This is because to be efficient, Bazel cache needs to be warm.

- ➡️  [Cheat sheet](#bazel-cheat-sheet)
- ➡️  [FAQ](#faq)
- 📽️ [Bazel Status Update](https://go/bazel-status)

## Why do we need a build system?

Building Sourcegraph is a non-trivial task, as it not only ships a frontend and a backend, but also a variety of third parties and components that makes the building process complicated, not only locally but also in CI. Historically, this always have been solved with ad-hoc solutions, such as shell scripts, and caching in various point of the process.

But we're using languages that traditionally don't require their own build systems right? Go and Typescript have their own ecosystem and solve those problems each with their own way right? Yes indeed, but this also means they are never aware of each other and anything linking them requires to be implemented manually, which what we've done so far. Because the way our app is built, as a monolith, it's not trivial to detect things such as the need to rebuild a given Docker image because a change was made in another package, because there is nothing enforcing this structurally in the codebase. So we have to rebuild things, because there is doubt.

On top of that, whenever we attempt at building our app, we also need to fetch many third parties from various locations (GitHub releases, NPM, Go packages, ...). While most of the time, it's working well, any failure in doing so will result in failed build. This may go unnoticed when working locally, but on CI, this can prevent us to build our app for hours at times if we can't fetch the dependency we need, until the problem is resolved upstream. This makes us very dependent on the external world.

In the end, it's not what composes our application that drives us to use a build system, but instead the size it has reached after years of development. We could solve all these problems individually with custom solutions, that would enable us to deterministically say that we need to build X because Y changed. But guess what? The result would pretty much look like a build system. It's a known problem and solutions exists in the wild for us to use.

Finally, build systems provides additional benefits, especially on the security side. Because a build system is by definition aware of every little dependency, we can use that to ensure we react swiftly to CVEs (Common Vulnerabilities and Exposures) produce SBOMs (Software Bill of Materials) for our customers to speed up the upgrade process.

## Why Bazel?

Bazel is the most used build system in the world that is fully language agnostic. It means you can build whatever you want with Bazel, from a tarball containing Markdown files to an iOS app. It's like Make, but much more powerful. Because of its popularity, it means its ecosystem is the biggest and a lot have been written for us to use already.

We could have used others, but that would translate in having to write much more. Building client code for example is a delicate task, because of the complexity of targeting browsers and the pace at which its ecosystem is evolving. So we're avoiding that by using a solution that has been battle tested and proved to still work at scale hundred times bigger than ours and smaller than us.

## What is Bazel?

Bazel sits in between traditional tools that build code, and you, similarly to how Make does one could say. At it's core, Bazel enables you to describe a hierarchy of all the pieces needed to build the application, plus the steps required to build one based on the others.

### What a build system does

Let's take a simple example: we are building a small tool that is written in Go and we ship it with `jq` so our users don't need to install it to use our code. We want our release tarball to contain our app binary, `jq` and our README.

Our codebase would look like this:

```
- README.md
- main.go
- download_jq.sh
```

The result would look like this:

```
- app.tar.gz
  - ./app
  - ./jq
  - README.md
```

To built it we need to perform the following actions:

1. Build `app` with `go build ...`
1. Run `./download_jq.sh` to fetch the `jq` binary
1. Create a tarball containing `app`, `jq` and `README.md`

If we project those actions onto our filetree, it looks like this (let's call it an _action graph_):

```
- app.tar.gz
  - # tar czvf app.tar.gz .
    - ./app
      - # go build main.go -o app
    - ./jq
      - # ./download_jq.sh
    - README.md
```

We can see how we have a tree of _inputs_ forming the final _output_ which is `app.tar.gz`. If all _inputs_ of a given _output_ didn't change, we don't need to build them again right? That's exactly the question that a build system can answer, and more importantly *deterministically*. Bazel is going to store all the checksums of _inputs_ and _outputs_ and will perform only what's required to generate the final _output_.

If our Go code did not change, we're still using the same version of `jq` but the README changed, do I need to generate a new tarball? Yes because the tarball depends on the README as well. If neither changed, we can simply keep the previous tarball. If we do not have Bazel, we need to provide a way to ensure it.

As long as Bazel's cache is warm, we'll never need to run `./download_jq.sh` to download `jq` again, meaning that even if GitHub is down and we can't fetch it, we can still build our tarball.

For Go and Typescript, this means that every dependency, either a Go module or a NPM package will be cached, because Bazel is aware of it. As long as the cache is warm, it will never download it again. We can even tell Bazel to make sure that the checksum of the `jq` binary we're fetching stays the same. If someone were to maliciously swap a `jq` release with a new one, Bazel would catch it, even it was the same exact version.

### Tests are outputs too.

Tests, whether it's a unit test or an integration tests, are _outputs_ when you think about it. Instead of being a file on disk, it's just green or red. So the same exact logic can be applied to them! Do I need to run my unit tests if the code did not change? No you don't, because the _inputs_ for that test did not change.

Let's say you have integration tests driving a browser querying your HTTP API written in Go. A naive way of representing this would be to say that the _inputs_ for that e2e test are the source for the tests. A better version would be to say that the _inputs_ for this tests are also the binary powering your HTTP API. Therefore, changing the Go code would trigger the e2e tests to be ran again, because it's an _input_ and it changed again.

So, building and testing is in the end, practically the same thing.

### Why is Bazel frequently mentioned in a negative light on Reddit|HN|Twitter|... ?

Build systems are solving a complex problem. Assembling a deterministic tree of all the _inputs_ and _outputs_ is not an easy task, especially when your project is becoming less and less trivial. And to enforce it's properties, such as hermeticity and being deterministic, Bazel requires both a "boil the ocean first" approach, where you need to convert almost everything in your project to benefit from it and to learn how to operate it. That's quite the upfront cost and a lot of cognitive weight to absorb, naturally resulting in negative opinions.

In exchange for that, we get a much more robust system, resilient to some unavoidable problems that comes when building your app requires to reach the outside world.

## Bazel for teammates in a hurry

### Bazel vocabulary

- A _rule_ is a function that stitches together parts of the graph.
  - ex: build go code
- A _target_ is a named rule invocation.
  - ex: build the go code for `./app`
  - ex: run the unit tests for `./app`
- A _package_ is a a group of _targets_.
  - ex: we only have one single package in the example above, the root one.

Bazel uses two types of files to define those:

- `WORKSPACE`, which sits at the root of a project and tells Bazel where to find the rules.
  - ex: get the Go _rules_ from that repository on GitHub, in this exact version.
- `BUILD.bazel`, which sits in every folder that contains _targets_.

To reference them, the convention being used is the following: `//pkg1/pkg2:my_target` and you can say things such as `//pkg1/...` to reference all possible _targets_ in a package.

Finally, let's say we have defined in our Bazel project some third party dependencies (a NPM module or a Go package), they will be referenced using the `@` sign.

- `@com_github_keegancsmith_sqlf//:sqlf`

### Sandboxing

Bazel ensures it's running hermetically by sandboxing anything it does. It won't build your code right in your source tree. It will copy all of what's needed to build a particular _target_ in a temporary directory (and nothing more!) and then apply all the rules defined for these _targets_.

This is a *very important* difference from doing things the usual way. If you didn't tell Bazel about an _input_, it won't be built/copied in/over the sandbox. So if your tests are relying testdata for examples, Bazel must be aware of it. This means that it's not possible to change the _outputs_ by accident because you created an additional file in the source tree.

So having to make everything explicit means that the buildfiles (the `BUILD.bazel` files) need to be kept in sync all the time. Luckily, Bazel comes with a solution to automate this process for us.

### Generating buildfiles automatically

Bazel ships with a tool named `Gazelle` whose purpose is to take a look at your source tree and to update the buildfiles for you. Most of the times, it's going to do the right thing. But in some cases, you may have to manually edit the buildfiles to specify what Gazelle cannot guess for you.

Gazelle and Go: It works almost transparently with Go, it will find all your Go code and infer your dependencies from inspecting your imports. Similarly, it will inspect the `go.mod` to lock down the third parties dependencies required. Because of how well Gazelle-go works, it means that most of the time, you can still rely on your normal Go commands to work. But it's recommended to use Bazel because that's what will be used in CI to build the app and ultimately have the final word in saying if yes or no a PR can be merged. See the [cheat sheet section](#bazel-cheat-sheet) for the commands.

Gazelle and the frontend: see [Bazel for Web bundle](./bazel_web.md).

### Bazel cheat sheet

#### Keep in mind

- Do not commit file whose name include spaces, Bazel does not like it.
- Do not expect your tests to be executed inside the source tree and to be inside a git repository.
  - They will be executed in the sandbox. Instead create a temp folder and init a git repo manually over there.

#### Building and testing things

- `bazel build [path-to-target]` builds a target.
  - ex `bazel build //lib/...` will build everything under the `/lib/...` folder in the Sourcegraph repository.
- `bazel test [path-to-target]` tests a target.
  - ex `bazel test //lib/...` will run all tests under the `/lib/...` folder in the Sourcegraph repository.
- `bazel configure` automatically inspect the source tree and update the buildfiles if needed.
- `bazel run //:gazelle-update-repos` automatically inspect the `go.mod` and update the third parties dependencies if needed.

#### Debugging buildfiles

- `bazel query "//[pkg]/..."` See all subpackages of `pkg`.
- `bazel query "//[pkg]:*"` See all targets of `pkg`.
- `bazel query //[pkg] --output location` prints where the buildfile for `pkg` is.
  - ex: `bazel query @com_github_cloudflare_circl//dh/x448 --output location` which allows to inspect the autogenerated buildfile.
- `bazel query "allpaths(pkg1, pkg2)"` list all knowns connections from `pkg1` to `pkg2`
  - ex `bazel query "allpaths(//enterprise/cmd/worker, @go_googleapis//google/api)"`
  - This is very useful when you want to understand what connects a given package to another.

#### Running bazel built services locally with `sg start`

> For early adopters only.

First you need to have `bazel` installed obviously, but also `iBazel` which will watch your files and rebuild if needed. We use a tool called `bazelisk` (which is also part of Bazel) to manage the version of `bazel`. It inspects a bunch of files to determine what `bazel` version to use for your repo.

If you want the setup automated run `sg setup`, otherwise you can install it manually by executing the following commands:

- `brew install bazelisk`
- `brew install ibazel`

Then instead of running `sg start oss` you can use the `bazel` variant instead.

- `sg start oss-bazel`
- `sg start enterprise-bazel`
- `sg start codeintel-bazel`
- `sg start enterprise-codeintel-bazel`

##### How it works

When `sg start` is booting up, the standard installation process will begin as usual for commands that are not built with Bazel, but you'll also see a new program running
`[  bazel]` which will log the build process for all the targets required by your chosen commandset. Once it's done, these services will start and [`iBazel`](https://github.com/bazelbuild/bazel-watcher)
will take the relay. It will watch the files that Bazel has indentified has dependencies for your services, and rebuild them accordingly.

So when a change is detected, `iBazel` will build the affected target and it will be restarted once the build finishes.

##### Caveats

- You still need to run `bazel configure` if you add/remove files or packages.
- Error handling is not perfect, so if a build fails, that might stop the whole thing. We'll improve this in the upcoming days, as we gather feedback.

## FAQ

### General

#### I just used `[no-bazel]` to merge my PR 

While using `[no-bazel]` will enable you to get your pull request merged, the subsequent builds will be with Bazel unless they also have that flag. 

Therefore you need to follow-up quickly with a fix to ensure `main` is not broken. 

#### The analysis cache is being busted because of `--action_env`

Typically you'll see this (in CI or locally):

```
INFO: Build option --action_env has changed, discarding analysis cache.
```

- If you added a `build --action_env=VAR` to one of the `bazelrc`s, and `$VAR` is not stable across builds, it will break the analysis cache. You should never pass a variable that is not stable, otherwise, the cache being busted is totally expected and there is no way around it.
  - Use `build --action_env=VAR=123` instead to pin it down if it's not stable in your environment.
- If you added a `test --action_env=VAR`, running `bazel build [...]` will have a different `--action_env` and because the analysis cache is the same for `build` and `test` that will automatically bust the cache.
  - Use `build --test_env=VAR` instead, so that env is used only in tests, and doesn't affect builds, while avoiding to bust the cache.

#### My JetBrains IDE becomes unresponsive after Bazel builds

By default, JetBrains IDEs such as GoLand will try and index the files in your project workspace. If you run Bazel locally, the resulting artifacts will be indexed, which will likely hog the full heap size that the IDE is allocated.  

There is no reason to index these files, so you can just exclude them from indexing by right-clicking artifact directories, then choosing **Mark directory as** &rarr; **Excluded** from the context menu. A restart is required to stop the indexing process. 

#### My local `bazel configure` or `./dev/ci/bazel-configure.sh` run has diff with a result of Bazel CI step

This could happen when there are any files which are not tracked by Git. These files affect the run of `bazel configure` and typically add more items to `BUILD.bazel` file.

Solution: run `git clean -ffdx` then run `bazel configure` again. 

### Go

#### It complains about some missing symbols, but I'm sure they are there since I can see my files

```
ERROR: /Users/tech/work/sourcegraph/internal/redispool/BUILD.bazel:3:11: GoCompilePkg internal/redispool/redispool.a failed: (Exit 1): builder failed: error executing command (from target //internal/redispool:redispool) bazel-out/darwin_arm64-opt-exec-2B5CBBC6/bin/external/go_sdk/builder compilepkg -sdk external/go_sdk -installsuffix darwin_arm64 -src internal/redispool/redispool.go -src internal/redispool/sysreq.go ... (remaining 30 arguments skipped)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
internal/redispool/redispool.go:78:13: undefined: RedisKeyValue
internal/redispool/redispool.go:94:13: undefined: RedisKeyValue
```

OR

```
~/work/sourcegraph U bzl/build-go $ bazel build //dev/sg
INFO: Analyzed target //dev/sg:sg (955 packages loaded, 16719 targets configured).
INFO: Found 1 target...
ERROR: /Users/tech/work/sourcegraph/internal/conf/confdefaults/BUILD.bazel:3:11: GoCompilePkg internal/conf/confdefaults/confdefaults.a failed: (Exit 1): builder failed: error executing command (from target //internal/conf/confdefaults:confdefaults) bazel-out/darwin_arm64-opt-exec-2B5CBBC6/bin/external/go_sdk/builder compilepkg -sdk external/go_sdk -installsuffix darwin_arm64 -src internal/conf/confdefaults/confdefaults.go -embedroot '' -embedroot ... (remaining 19 arguments skipped)

Use --sandbox_debug to see verbose messages from the sandbox and retain the sandbox build root for debugging
compilepkg: missing strict dependencies:
	/private/var/tmp/_bazel_tech/3eea80c6015362974b7d423d1f30cb62/sandbox/darwin-sandbox/10/execroot/__main__/internal/conf/confdefaults/confdefaults.go: import of "github.com/russellhaering/gosaml2/uuid"
No dependencies were provided.
Check that imports in Go sources match importpath attributes in deps.
Target //dev/sg:sg failed to build
Use --verbose_failures to see the command lines of failed build steps.
INFO: Elapsed time: 11.559s, Critical Path: 2.93s
INFO: 36 processes: 2 internal, 34 darwin-sandbox.
```


Solution: run `bazel configure` to update the buildfiles automatically.


#### My go tests complains about missing testdata

In the case where your testdata lives in `../**`, Gazelle cannot see those on its own, and you need to create a filegroup manually, see https://github.com/sourcegraph/sourcegraph/pull/47605/commits/93c838aad5436dc69f6695cec933bfb84b8ba59a

#### Manually adding a `go_repository`

Sometimes Gazelle won't be able to generate a `go_repository` for your dependency and you'll need to fill in the attributes yourself. Most of the fields are easy to get, except when you need to provide values for the sum and version.

To retrieve these values:
1. Create a go.mod in the directory where the dependency is imported.
2. Run `go mod tidy`. This will populate the `go.mod` file and also generate a `go.sum` file.
3. You can then locate the version you should use for `go_repository` from the `go.mod` file and the sum from the `go.sum` file.
4. Delete the `go.mod` and `go.sum` files as they're no longer needed.

#### How to update to the latest recommended bazelrc?

```
bazel run //.aspect/bazelrc:update_aspect_bazelrc_presets
```

### Rust

#### I'm getting `Error in path: Not a regular file: docker-images/syntax-highlighter/Cargo.Bazel.lock` when I try to build `syntax-highlighter`

Below is a full example of this error:
```
ERROR: An error occurred during the fetch of repository 'crate_index':
   Traceback (most recent call last):
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/crates_repository.bzl", line 34, column 30, in _crates_repository_impl
                lockfiles = get_lockfiles(repository_ctx)
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/generate_utils.bzl", line 311, column 36, in get_lockfiles
                bazel = repository_ctx.path(repository_ctx.attr.lockfile) if repository_ctx.attr.lockfile else None,
Error in path: Not a regular file: /Users/william/code/sourcegraph/docker-images/syntax-highlighter/Cargo.Bazel.lock
ERROR: /Users/william/code/sourcegraph/WORKSPACE:197:18: fetching crates_repository rule //external:crate_index: Traceback (most recent call last):
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/crates_repository.bzl", line 34, column 30, in _crates_repository_impl
                lockfiles = get_lockfiles(repository_ctx)
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/generate_utils.bzl", line 311, column 36, in get_lockfiles
                bazel = repository_ctx.path(repository_ctx.attr.lockfile) if repository_ctx.attr.lockfile else None,
Error in path: Not a regular file: /Users/william/code/sourcegraph/docker-images/syntax-highlighter/fake.lock
ERROR: Error computing the main repository mapping: no such package '@crate_index//': Not a regular file: /Users/william/code/sourcegraph/docker-images/syntax-highlighter/Cargo.Bazel.lock
```
The error happens when the file specified in the lockfiles attribute of `crates_repository` (see WORKSPACE file for the definition) does not exist on disk. Currently this rule does not generate the file, instead it just generates the _content_ of the file. So to get passed this error you should create the file `touch docker-images/syntax-highlighter/Cargo.Bazel.lock`. With the file create it we can now populate `Cargo.Bazel.lock` with content using bazel by running `CARGO_BAZEL_REPIN=1 bazel sync --only=crate_index`.

### When I build `syntax-highlighter` it complains that the current `lockfile` is out of date
The error will look like this:
```
INFO: Repository crate_index instantiated at:
  /Users/william/code/sourcegraph/WORKSPACE:197:18: in <toplevel>
Repository rule crates_repository defined at:
  /private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/crates_repository.bzl:106:36: in <toplevel>
INFO: repository @crate_index' used the following cache hits instead of downloading the corresponding file.
 * Hash 'dc2d47b42cbe92ebdb144555603dad08eae505fc459bae5e2503647919067ac8' for https://github.com/bazelbuild/rules_rust/releases/download/0.16.1/cargo-bazel-aarch64-apple-darwin
If the definition of 'repository @crate_index' was updated, verify that the hashes were also updated.
ERROR: An error occurred during the fetch of repository 'crate_index':
   Traceback (most recent call last):
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/crates_repository.bzl", line 45, column 28, in _crates_repository_impl
                repin = determine_repin(
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/generate_utils.bzl", line 374, column 13, in determine_repin
                fail(("\n".join([
Error in fail: Digests do not match: Digest("3e9e0f927c955efa39a58c472a2eac60e3f89a7f3eafc7452e9acf23adf8ce5a") != Digest("ef858ae49063d5c22e0ee0b7632a8ced4994315395b17fb3c61f3e6bfb6deb27")

The current `lockfile` is out of date for 'crate_index'. Please re-run bazel using `CARGO_BAZEL_REPIN=true` if this is expected and the lockfile should be updated.
ERROR: /Users/william/code/sourcegraph/WORKSPACE:197:18: fetching crates_repository rule //external:crate_index: Traceback (most recent call last):
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/crates_repository.bzl", line 45, column 28, in _crates_repository_impl
                repin = determine_repin(
        File "/private/var/tmp/_bazel_william/c92ec739369034d3064b6df55c419545/external/rules_rust/crate_universe/private/generate_utils.bzl", line 374, column 13, in determine_repin
                fail(("\n".join([
Error in fail: Digests do not match: Digest("3e9e0f927c955efa39a58c472a2eac60e3f89a7f3eafc7452e9acf23adf8ce5a") != Digest("ef858ae49063d5c22e0ee0b7632a8ced4994315395b17fb3c61f3e6bfb6deb27")

The current `lockfile` is out of date for 'crate_index'. Please re-run bazel using `CARGO_BAZEL_REPIN=true` if this is expected and the lockfile should be updated.
ERROR: Error computing the main repository mapping: no such package '@crate_index//': Digests do not match: Digest("3e9e0f927c955efa39a58c472a2eac60e3f89a7f3eafc7452e9acf23adf8ce5a") != Digest("ef858ae49063d5c22e0ee0b7632a8ced4994315395b17fb3c61f3e6bfb6deb27")

The current `lockfile` is out of date for 'crate_index'. Please re-run bazel using `CARGO_BAZEL_REPIN=true` if this is expected and the lockfile should be updated.
```
Bazel uses a separate lock file to keep track of the dependencies and needs to be updated. To update the `lockfile` run `CARGO_BAZEL_REPIN=1 CARGO_BAZEL_REPIN_ONLY=crate_index bazel sync --only=crate_index`. This command takes a while to execute as it fetches all the dependencies specified in `Cargo.lock` and populates `Cargo.Bazel.lock`.

### `syntax-highlighter` fails to build and has the error `failed to resolve: use of undeclared crate or module`
The error looks something like this:
```
error[E0433]: failed to resolve: use of undeclared crate or module `scip_treesitter_languages`
  --> docker-images/syntax-highlighter/src/main.rs:56:5
   |
56 |     scip_treesitter_languages::highlights::CONFIGURATIONS
   |     ^^^^^^^^^^^^^^^^^^^^^^^^^ use of undeclared crate or module `scip_treesitter_languages`

error[E0433]: failed to resolve: use of undeclared crate or module `scip_treesitter_languages`
  --> docker-images/syntax-highlighter/src/main.rs:57:15
   |
57 |         .get(&scip_treesitter_languages::parsers::BundledParser::Go);
   |               ^^^^^^^^^^^^^^^^^^^^^^^^^ use of undeclared crate or module `scip_treesitter_languages`

error: aborting due to 2 previous errors
```
Bazel doesn't know about the module/crate being use in the rust code. If you do a git blame `Cargo.toml` you'll probably see that a new dependency has been added, but the build files were not updated. There are two ways to solve this:
1. Run `bazel configure` and `CARGO_BAZEL_REPIN=1 CARGO_BAZEL_REPIN_ONLY=crate_index bazel sync --only=crate_index`. Once the commands have completed you can check that the dependency has been picked up and syntax-highlighter can be built by running `bazel build //docker-images/syntax-highlighter/...`. **Note** this will usually work if the dependency is an *external* dependency.
2. You're going to have to update the `BUILD.bazel` file yourself. Which one you might ask? From the above error we can see the file `src/main.rs` is where the error is encountered, so we need to tell *its BUILD.bazel* about the new dependency. 
For the above dependency, the crate is defined in `docker-images/syntax-highlighter/crates`. You'll also see that each of those crates have their own `BUILD.bazel` files in them, which means we can reference them as targets! Take a peak at `scip-treesitter-languages` `BUILD.bazel` file and take note of the name - that is its target. Now that we have the name of the target we can add it as a dep to `docker-images/syntax-highlighter`. In the snippet below the `syntax-highlighter` `rust_binary` rule is updated with the `scip-treesitter-languages` dependency. Note that we need to refer to the full target path when adding it to the dep list in the `BUILD.bazel` file.
```
rust_binary(
    name = "syntect_server",
    srcs = ["src/main.rs"],
    aliases = aliases(),
    proc_macro_deps = all_crate_deps(
        proc_macro = True,
    ),
    deps = all_crate_deps(
        normal = True,
    ) + [
        "//docker-images/syntax-highlighter/crates/sg-syntax:sg-syntax",
        "//docker-images/syntax-highlighter/crates/scip-treesitter-languages:scip-treesitter-languages",
    ],
)
```

### When using nix, when protobuf gets compiled I get a C/C++ compiler issue `fatal error: 'algorithm' file not found` or some core header files are not found
Nix sets the CC environment variable to a clang version use by nix which is independent of the host system. You can verify this by running the following commands in your nix shell.
```
$ echo $CC
clang

$ which $CC
/nix/store/agjhf1m0xsvmdjkk8kc7bp3pic9lsfrb-clang-wrapper-11.1.0/bin/clang

$ cat bazel-sourcegraph/external/local_config_cc/cc_wrapper.sh | grep "# Call the C++ compiler" -A 2
# Call the C++ compiler
/nix/store/agjhf1m0xsvmdjkk8kc7bp3pic9lsfrb-clang-wrapper-11.1.0/bin/clang "$@"
```
Bazel runs a target called `locate_cc_config` which adheres to the CC environment variable. The variable defines the compiler to be used to perform C/C++ compilation. At time of writing, the compiler is incorrectly configured and the stdlib doesn't get referenced properly. Therefore, we currently recommend to unset the `CC` variable in your nix shell. The `locate_cc_config` will then find the system C/C++ compiler (which on my system resolved to `/usr/bin/gcc`) and compile protobuf.

You can also verify that the correct compiler is used by running the following command:
```
cat bazel-sourcegraph/external/local_config_cc/cc_wrapper.sh | grep "# Call the C++ compiler" -A 2
# Call the C++ compiler
/usr/bin/gcc "$@"
```

## Resources

- [Core Bazel (book)](https://www.amazon.com/Core-Bazel-Fast-Builds-People/dp/B08DVDM7BZ):
  - [Bazel User guide](https://bazel.build/docs)
- [Writing a custom rule that depends on an external dep](https://www.youtube.com/watch?v=bhirT014eCE)
- [Patching third parties when they don't build](https://rotemtam.com/2020/10/30/bazel-building-cgo-bindings/)
