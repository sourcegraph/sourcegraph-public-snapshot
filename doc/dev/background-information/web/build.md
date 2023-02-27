# TypeScript build documentation

This document describes the TypeScript projects in this repository and how they are built.

## Build products

We use TypeScript for two products:

- [`web`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/web): The main Sourcegraph web application
  - 2 different entrypoints: [OSS `main.tsx`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/web/src/main.tsx) and [Enterprise `main.tsx`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/web/src/enterprise/main.tsx)
- [`browser`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/client/brower): The Sourcegraph browser extension

These both use shared TypeScript code in [`../shared`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/shared). Each product has its own separate Webpack configuration.

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

### Debugging production build

The web application exposes a global `window.buildInfo` object containing `version` and `commitSHA` used to build the application bundle. If you're unsure what version of the client bundle you're debugging, use these variables to validate your assumptions.

### Design

Based on our experience, we decided to:

- Use only the most standard tools: `tsc` and `pnpm`. (Bonus points for not using `pnpm`-specific features, to preserve optionality to switch back to `npm`.)
- Do not build shared code to an intermediate output directory. Instead, import shared `.ts` and `.tsx` files directory from product code.
- Use a single root `package.json` that specifies all dependencies needed by any product or shared code.

The one "hack" is that each subproject's `node_modules/.bin` is symlinked to the root `node_modules/.bin` so that `package.json` scripts can refer to programs installed by dependencies. (Subprojects' `node_modules` directories are otherwise empty.)

### HowTos

#### Add a dependency

Run `pnpm add PACKAGE` or `pnpm add -D PACKAGE` in the root directory.

#### Upgrade a dependency

Run `pnpm up --latest PACKAGE`.

### esbuild

[esbuild](https://esbuild.github.io/) is an alternative to Webpack for building `client/web` (the web app). Its usage on our codebase is **EXPERIMENTAL** and optional.

To use esbuild instead of Webpack, set the env var `DEV_WEB_BUILDER=esbuild` (when using `sg start`).

Comparison vs. Webpack:

- esbuild: faster initial build (esbuild ~3s vs. Webpack ~53s)
- esbuild: faster recompilation (esbuild ~900ms vs. Webpack ~5000ms)
- esbuild: smaller total asset size in dev (Chrome devtools network resources size for `/search`: esbuild ~24.1MB vs. Webpack ~64.1MB)
- esbuild: faster DOMContentLoaded (on `/search`: esbuild ~1.2s vs. Webpack ~2.3s)
- Webpack: [fast refresh](https://www.npmjs.com/package/react-refresh) (not supported/implemented yet in esbuild, so you need to manually reload the page after each change)

Notes:

- It's probably possible to configure Webpack to be faster and produce smaller dev bundles, so consider these comparisons as reflecting the current state, not the hypothetical ideal state after more optimization.
- Webpack is still used for all other web builds (including storybooks and the browser extension).
- esbuild is not configured to make a production build. Just use it for local dev for now.

Questions or problems with esbuild? Ask in [#frontend-platform](https://app.slack.com/client/T02FSM7DL/C01LTKUHRL3) and mention `@sqs` (who is responsible for this esbuild experiment).
