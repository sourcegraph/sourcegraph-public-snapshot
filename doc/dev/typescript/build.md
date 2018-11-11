# TypeScript build documentation

This document describes the TypeScript projects in this repository and how they are built.

## Build products

We use TypeScript for two products:

- [`web`](../../../web): The main Sourcegraph web application
  - 2 different entrypoints: [OSS `main.tsx`](../../../web/src/main.tsx) and [Enterprise `main.tsx`](../../../web/src/enterprise/main.tsx)
- [`client/browser`](../../../client/brower): The Sourcegraph browser extension

These both use shared TypeScript code in [`../shared`](../../../shared). Each product has its own separate Webpack configuration.

## Build process and configuration

### Goals

- It should be simple for anyone to make changes to the web app or browser extension.
  - The TypeScript build configurations should work well with Webpack, `tsc`, storybooks, and VS Code (and other editors that use `tsserver`).
  - Go-to-definition, find-references, auto-import-completion, and other editor features should work across all shared code (with no jumps to generated `.d.ts` files).
  - An edit to a shared TypeScript file should be directly reflected in both products' build processes in all of those tools.
- It should feel like a single, consistent user experience to use the web app and browser extension.
  - Corollary: These should be developed together most of the time. They should feel like the same codebases, and new features that are relevant to both should be made by the same person and in the same commit/PR. (The browser extension needs more backcompat than the web app, because the browser extension must support communicating with older Sourcegraph instances.)
- Make the edit-reload-debug cycle for errors as quick as possible.

### Background

We have tried two things that ended up not satisfying our needs:

- One repository per package: The overhead of sharing code was too high. It required publishing intermediate packages (that were not used by any other consumers).
- [Yarn workspaces](https://yarnpkg.com/lang/en/docs/workspaces/): The overhead of sharing code was still too high. Also, we encountered bugs (like [#4964](https://github.com/yarnpkg/yarn/issues/4964)) that made us feel it was not ready for production use.

### Design

Based on our experience, we decided to:

- Use only the most standard tools: `tsc` and `yarn`. (Bonus points for not using `yarn`-specific features, to preserve optionality to switch back to `npm`.)
- Do not build shared code to an intermediate output directory. Instead, import shared `.ts` and `.tsx` files directory from product code.
- Use manual solutions to fill in the gaps:
  - For shared dependencies whose versions should be consistent across both products (such as `react`, `@types/react`, `rxjs`, etc.), use symlinks by specifying `dependencies` and `devDependencies` entries like `"rxjs": "link:../node_modules/rxjs"` in [`web/package.json`](../../../web/package.json), [`client/browser/package.json`](../../../client/browser/package.json), [`shared/package.json`](../../../shared/package.json). We accept the risk of this introducing an unintended, unhoisted version of a linked package from another dependency that also depends on it (which is why it sometimes results in a warning that the `yarn.lock` version is incorrect).
  - To kick off `yarn install` for all products and shared code, add a `prepublish` script to the [root `package.json`](../../../package.json) that runs this for each other build directory.

The first 2 items are big simplicity wins. The 3rd item unfortunately offsets the simplicity gains. There is no silver bullet.

### Howtos

#### Add a dependency

In most cases, just run `yarn add PACKAGE` or `yarn add -D PACKAGE` in the directory containing the relevant `package.json` as you would normally.

If the dependency is used by multiple `package.json`s and you know it's important for the same version of the dependency to be used by all products and shared code (e.g., for `@types/react` or `rxjs`, which need fine-grained type compatibility), then follow these steps to add the dependency:

1. In the **root directory**, run `yarn add PACKAGE` or `yarn add -D PACKAGE` (where `PACKAGE` is the dependency you want to add).
1. In each of the directories containing the `package.json`s of the products or shared code that need this dependency (e.g., [`web`](../../../web) and [`client/browser`](../../../client/browser)):
   1. Add a `dependencies` or `devDependencies` entry linking to the root's `node_modules/PACKAGE` directory:
      - `"PACKAGE": "link:../node_modules/PACKAGE"`
      - `"PACKAGE": "link:../../node_modules/PACKAGE"` (from [`client/browser`](../../../client/browser) because it is 2 directories deep)
   1. Run `yarn` in the directory containing the `package.json` that you just modified (to create the symlink).

The directory structure now looks like (supposing that you both [`web`](../../../web) and [`client/browser`](../../../client/browser) need this dependency):

```
/package.json:
    file containing dependencies/devDependencies "PACKAGE": "^N.N.N"
/node_modules/PACKAGE
    contains the dependency's files
/client/browser/package.json
   file containing dependencies/devDependencies "PACKAGE": "link:../../node_modules/PACKAGE"
/client/browser/node_modules/PACKAGE
   symlink to /node_modules/PACKAGE
/web/package.json
   file containing dependencies/devDependencies "PACKAGE": "link:../node_modules/PACKAGE"
/web/node_modules/PACKAGE
   symlink to /node_modules/PACKAGE
```

(This basically entails manually doing what [Yarn workspaces](https://yarnpkg.com/lang/en/docs/workspaces/) attempts to do automatically. It is better than using Yarn workspaces because it lets us avoid the other bugs we encountered and doesn't require us to change the directory structure of our repository.)

#### Upgrade a dependency

In the directory containing the `package.json` that has the dependency's version number, run `yarn upgrade -L PACKAGE` (as you would normally).

### FAQ

#### Yarn warning: Lockfile has incorrect entry

When you run `yarn`, you may see warning messages like:

```
warning Lockfile has incorrect entry for "@babel/core@^7.1.2". Ignoring it.
warning Lockfile has incorrect entry for "marked@^0.5.0". Ignoring it.
warning Lockfile has incorrect entry for "react-router@^4.3.1". Ignoring it.
```

This can occur for any dependency whose entry in `dependencies` or `devDependencies` uses `link:` (e.g., `"@babel/core": "link:../node_modules/@babel/core"`). It is usually harmless and is a consequence of the [design](#design) of our TypeScript build configuration.

If you believe that this is causing a problem, then you can fix it as follows (using `@babel/core` as an example):

1. Run `yarn why @babel/core` to find which other package is asking for the other version of the dependency. The output in this case mentions:
  - `Found "@babel/core@7.1.5"`, which means that the [root `package.json`](../../../package.json) is using `@babel/core@7.1.5`.
  - `Hoisted from stylelint#postcss-jsx#@babel#core`, which means the `postcss-jsx` indirect dependency is requesting `@babel/core@^7.1.2`.
1. In the root directory, run `yarn add -D postcss-jsx` (if this dependency is not for development, omit the `-D` as you normally would).
1. In the `package.json` of the directory that these warnings came from, add a `dependencies` or `devDependencies` entry linking to the root's `node_modules/postcss-jsx` directory: `"postcss-jsx": "link:../node_modules/postcss-jsx"`.
1. Rerun `yarn` in the original directory (to create the symlink).
1. If the warning is still printed, follow steps 2-4 for `stylelint` as well.

To completely eliminate this problem, we could add **all** dependencies to the [root `package.json`](../../../package.json). If this problem becomes costly, we will do that.

##### Why this occurs (technical details)

It occurs when another dependency in the same `package.json` also depends on the same package but at a different version. Take `@babel/core` as an example. (The specific version numbers mentioned here will change in the future, but the explanation still applies.) Our [`shared/package.json`](../../../shared/package.json) depends on `"@babel/core": "link:../node_modules/@babel/core"`. To determine what version is actually used, consult the root [`package.json`](../../../package.json), which has `"@babel/core": "^7.1.5"`. However, when computing the `yarn.lock` lockfile, Yarn only consults the versions specified by non-linked dependencies, and the indirect dependency `postcss-jsx` depends on `@babel/core@^7.1.2`. That's the version that gets written to the lockfile.

#### Types of A and B are incompatible

You may see TypeScript compiler errors like:

```
Types of parameters 'subscriber' and 'subscriber' are incompatible.
      Type 'import("$SOURCEGRAPH/node_modules/rxjs/internal/Subscriber").Subscriber<any>' is not assignable to type 'import("$SOURCEGRAPH/node_modules/@sourcegraph/codeintellify/node_modules/rxjs/internal/Subscriber").Subscriber<any>'.
```

This occurs because the `@sourcegraph/codeintellify` dependency exports API definitions that reference types at a *different* version of `rxjs` from that used by the consumer of `@sourcegraph/codeintellify`.

To see the specific versions of `rxjs` in use, run `yarn why rxjs` from the directory containing the `package.json` of the product or shared code where this error comes from. To fix the issue, upgrade `@sourcegraph/codeintellify` to a version that depends on an up-to-date `rxjs` release.
