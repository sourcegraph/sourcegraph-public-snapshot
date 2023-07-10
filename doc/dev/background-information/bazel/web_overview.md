# Overview of the Bazel configuration for client/*

## TL;DR

TODO

## Overview

This guide goes over the details of the Bazel integration with the client code.
Check out the **Bazel for teammates in a hurry** section [here](./index.md) first.

- [npm dependencies with Bazel](#npm-dependecies-with-bazel)
  - [running Node.js programs with `rules_js`](#running-nodejs-programs-with-rules_js)
- [WORKSPACE file, which inits the Bazel project](#workspace-file)
- [Root BUILD.bazel file, which defines many shared config targets for client packages](#root-build-bazel-file)
  - [`js_library` target definition](#js_library-target-definition)
  - [What is `js_library`](#what-is-js_library)
  - [What is build graph and info providers?](#what-is-build-graph-and-info-providers)
  - [`ts_config`](#ts_config)
  - [npm binary imports](#npm-binary-imports)
    - [Under the hood](#under-the-hood)
- [Gazelle for Javascript/Typescript code](#gazelle-for-javascript-typescript-code)
  - [Gazelle configuration](#gazelle-configuration)
- [Structure of the simplest client package](#structure-of-the-simplest-client-package)
  - [How to create nested BUILD files to leverage caching](#how-to-create-nested-build-files-to-leverage-caching)
  - [How packages depend on each other](#how-packages-depend-on-each-other)
- [Client integration tests integration with Bazel](#client-integration-tests-integration-with-bazel)
- [ESLint custom rule implementation for type-aware linting](#eslint-custom-rule-implementation-for-type-aware-linting)
- [Debugging tips](#debugging-tips)

---

## npm dependecies with Bazel

[rules_js](https://docs.aspect.build/rules/aspect_rules_js/docs) is a tool used for managing Node.js programs in Bazel. It simplifies the downloading of third-party libraries using Bazel's built-in remote downloader and "[repository cache](https://bazel.build/run/build#repository-cache)". This works by converting the dependencies from `pnpm-lock.yaml` into Bazel's internal format, allowing Bazel to handle the downloads.

### Running Node.js programs with `rules_js`

Bazel operates in a different file system layout than Node.js. The challenge is to make Bazel and Node.js work together without re-writing npm packages or disrupting Node.js resolution algorithm. `rules_js` accomplishes this by running JavaScript tools in the Bazel output tree. It uses `pnpm` to create a `node_modules` directory in `bazel-out`, enabling smooth resolution of dependencies.

This approach comes with several benefits, like solving TypeScript's `rootDirs` issue. However, it does require Bazel rules/macro authors to adjust paths and ensure sources are copied to `bazel-out` before execution. This requires setting a `BAZEL_BINDIR` environment variable for each action.

## Workspace file

The workspace file serves two main purposes:

1. **Project Identification**: The presence of a WORKSPACE file in a directory identifies that directory as the root of a Bazel workspace. It's a marker file, telling Bazel that the directory and its subdirectories contain a software project that can be built and tested. Without the WORKSPACE file, Bazel wouldn't know which directory constitutes the root of the project.

2. **Dependency Management**: The WORKSPACE file is where you can define external dependencies your project might need. You specify these dependencies with workspace rules. These rules tell Bazel where to find and how to build the dependencies.

For example:

```py
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "aspect_rules_js",
    sha256 = "0b69e0967f8eb61de60801d6c8654843076bf7ef7512894a692a47f86e84a5c2",
    strip_prefix = "rules_js-1.27.1",
    url = "https://github.com/aspect-build/rules_js/releases/download/v1.27.1/rules_js-v1.27.1.tar.gz",
)
```

1. `load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")`: loads the `http_archive` rule from the Bazel built-in tools.
2. `http_archive(...)` downloads a specific version of the `rules_js` package from the provided URL, checks that its SHA-256 checksum matches the expected value, strips the leading directory from the paths of the files in the package, and then makes the package available under the name `aspect_rules_js`. This external repository can then be used in BUILD files, which describe how individual software components are built and tested. This snippet is usually copy-pasted from the releases section of rule (e.g., [rules_js](https://github.com/aspect-build/rules_js/releases))

## Root BUILD.bazel file

This `BUILD.bazel` file primarily configures various rules for our monorepo including linking npm packages, setting up static code analysis tools like ESLint, defining TypeScript and Babel configs, along with Gazelle for dependency management and BUILD file generation.

### `js_library` target definition

One of the first Javascript related snippets in the root `BUILD.bazel` file is `js_library` target definition:

```py
load("@aspect_rules_js//js:defs.bzl", "js_library")

js_library(
    name = "postcss_config_js",
    srcs = ["postcss.config.js"],
    deps = [
        "//:node_modules/autoprefixer",
        "//:node_modules/postcss-custom-media",
        "//:node_modules/postcss-focus-visible",
        "//:node_modules/postcss-inset",
    ],
)
```

1. `load("@aspect_rules_js//js:defs.bzl", "js_library")` imports the `js_library` rule from the `defs.bzl` file located in the `js` directory of the `aspect_rules_js` package that we fetched in the WORKSPACE file above.
2. `js_library(...)` initiates the declaration of a JS library.
3. `name = "postcss_config_js"` assigns the name to the JS library. This is how other rules will refer to this library if they depend on it or how we can reference this target in your bazel commands (e.g., `bazel build //:postcss_config_js`).
4. `srcs = ["postcss.config.js"]` specifies the list of source files that are included in this library.
5. `deps = [...]` defines dependencies of this target. This may include other `js_library` targets or other targets with JS files.

### What is js_library?

`js_library` is a rule provided by [rules_js](https://docs.aspect.build/rules/aspect_rules_js). It groups together JS sources and their transitive and npm dependencies into a provided `JsInfo` added to the build graph.

See [docs](https://docs.aspect.build/rules/aspect_rules_js/docs/js_library) for more details.

### What is build graph and info providers?

In the Bazel build system, everything is represented in a structure called a build graph. Each node in the graph represents a build target (like our `js_library`), and the edges between nodes represent dependencies. When a target is built, Bazel analyses the dependencies (edges) and builds those first, ensuring that everything a target needs to build correctly is in place before the build begins.

In case of `js_library`, each node represents a collection of JavaScript sources, which might be individual JS files or packages, and these nodes are linked by their dependencies. For example, our `postcss_config_js` node would be connected to the `autoprefixer`, `postcss-custom-media`, `postcss-focus-visible`, and `postcss-inset` nodes in the build graph.

`JsInfo` is a Bazel info provider. Info Providers supply metadata or additional information about the build target. In our example, `JsInfo` is like a property or an attribute attached to each `js_library` node in the build graph. This attribute contains essential information about what files this target provides once it is built. It can include various pieces of data, such as the names and locations of the JavaScript source files, the dependencies that the target has, and any other metadata that Bazel or other parts of the build system might need to know to properly build and use the target.

### ts_config

`ts_config` is a rule provided by [rules_ts](https://docs.aspect.build/rules/aspect_rules_ts/) enables a `tsconfig.json` file to extend another one in the build graph, ensuring Bazel recognizes the extended configuration and makes it available to dependent targets.

```py
load("@aspect_rules_ts//ts:defs.bzl", "ts_config")

ts_config(
    name = "tsconfig",
    src = "tsconfig.base.json",
    deps = [
        "//:node_modules/@sourcegraph/tsconfig",
    ],
)
```

See [docs](https://docs.aspect.build/rules/aspect_rules_ts/docs/rules#ts_config) for more details.

### npm binary imports

```py
load("@npm//:eslint/package_json.bzl", eslint_bin = "bin")

eslint_bin.eslint_binary(
    name = "eslint",
    testonly = True,
    visibility = ["//visibility:public"],
)
```

The `load()` function loads the `eslint` npm package that was installed by `rules_js` based on `pnpm-lock.yaml`. The syntax `"@npm//:eslint/package_json.bzl"` is indicating the location of the `eslint` package within the Bazel-managed npm dependencies. It's loading the `bin` macro which is generated by `rules_js` for the `eslint` package and aliasing it as `eslint_bin`.

After that, `eslint_bin.eslint_binary()` is used to create a new Bazel build target. This target will execute the eslint binary when run (bazel way of doing `pnpm eslint`).

#### Under the hood

By running `bazel aquery '//:eslint' | grep Outputs:` we can find out what files `rules_js` generate as a compatibility layer between eslint binary and Bazel:

```sh
INFO: Found 1 target...
  Outputs: [bazel-out/darwin_arm64-fastbuild/bin/eslint_node_bin/node]
  Outputs: [bazel-out/darwin_arm64-fastbuild/bin/eslint.sh]
  Outputs: [bazel-out/darwin_arm64-fastbuild/bin/eslint.sh.runfiles_manifest]
  Outputs: [bazel-out/darwin_arm64-fastbuild/bin/eslint.sh.runfiles/MANIFEST]
  Outputs: [bazel-out/darwin_arm64-fastbuild/internal/_middlemen/eslint.sh-runfiles]
```

`bazel-out/darwin_arm64-fastbuild/bin/eslint.sh` is the bash script wrapper that prepares arguments for the eslint binary, executes it, and handles exit codes in a Bazel-compatible way. If we need to debug issues with the binary we're trying to run, looking at the underlying bash script and printing information to stdout helps a lot.

See **I want to execute a js binary with bazel. How do I do that?** for more details [here](./web.md)

## Gazelle for Javascript/Typescript code

[Gazelle](https://github.com/bazelbuild/bazel-gazelle) is a BUILD file generator for Bazel, a software development tool. It automates the creation and management of Bazel's BUILD files by analyzing project source code and dependencies.

In the context of TypeScript and JavaScript, Gazelle uses Tree-sitter, a parser generator tool, to interpret the code. It navigates the abstract syntax tree (AST) produced by Tree-sitter to detect dependencies, including type-only imports in TypeScript, and reflects these relationships in the generated BUILD files.

Use `bazel configure` to update source file lists for `ts_project` rules in BUILD files. This command inspects all nested folders starting from the current one and non-destructively updates `ts_project` targets. You can use a # keep directive to force the tool to leave existing BUILD contents alone.

```py
ts_project(
  srcs = [
    "index.ts",
    "manual.ts" # keep
  ]
)
```

💡 Run `bazel configure` from the specific folder (e.g., `client/common`) to update only a subset of nested BUILD files instead of crawling the whole monorepo.

### Gazelle configuration

We can configure Gazelle using directives, which are specially-formatted comments in BUILD files that govern the tool's behavior when visiting files within the Bazel package rooted at that file. See the list of available JS directives [here](https://github.com/aspect-build/aspect-cli/blob/main/docs/help/topics/directives.md).

By default Gazelle for JS produces two targets: sources and tests based on globs used in config. We can use this [undocumented](https://github.com/aspect-build/aspect-cli/blob/5.5.2/gazelle/js/tests/groups_deps/BUILD.in) directive to define custom targets:

```py
# gazelle:js_custom_files e2e *.e2e.ts
# gazelle:js_custom_files pos *.po.ts
```

In our monorepo most of the JS related Gazelle directives are defined in the root BUILD file and in `./client/BUILD.bazel`.

## Structure of the simplest client package

Based on `./client/common/BUILD.bazel` as one of the smallest client packages.

### How to create nested BUILD files to leverage caching

### How packages depend on each other

- no circular dependecies
- legit npm packages with only JS sources under the hood powered by `npm_package`

## Client integration tests integration with Bazel

- esbuild bundling
- environment variables
- mocha integration

### Percy configuration and stamping

- Percy builds need to know the current commit and branch name which busts Bazel cache because it depends on target inputs
- Stamping to the rescue.
- Stamping is only available on one Aspect rule that's why we have 3 targets instead of one.

## ESLint custom rule implementation for type-aware linting

- npm_translate_lock with public_hoist_packages in the WORKSPACE file

## Debugging tips

- sandboxing
- print statements in Starlark code
- echo statements in generated bash scripts
- console logs from JS code
