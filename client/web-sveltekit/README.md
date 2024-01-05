# Sourcegraph SvelteKit

This folder contains the experimental [SvelteKit](https://kit.svelte.dev/)
implementation of the Sourcegraph app.

**NOTE:** This is a _very early_ prototype and it will change a lot.

## Developing

```bash
# Install dependencies
pnpm install
# Run dev server
pnpm run dev
```

You can also build the dotcom version by running `pnpm run dev:dotcom`, but it doesn't really differ
in functionality yet.

The dev server can be accessed on http://localhost:5173. API requests and
signin/signout are proxied to an actual Sourcegraph instance,
https://sourcegraph.com by default (can be overwritten via the
`SOURCEGRAPH_API_URL` environment variable.

### Using code from `@sourcegraph/*`

There are some things to consider when using code from other `@sourcegraph`
packages:

- Since we use the [barrel](https://basarat.gitbook.io/typescript/main-1/barrel)
  style of organizing our modules, many (unused) dependencies are imported into
  the app. This isn't ideal and at best will only increase the initial loading
  time. Some modules, especially those that access browser specific features
  during module initialization, can even cause the dev build to fail.
- Reusing code is great, but also potentially exposes someone who modifies the
  reused code to this package and therefore Svelte (if the reused code changes
  in an incompatible way, this package needs to be updated too). To limit the
  exposure, a module of any `@sourcegraph/*` package should only be imported
  once into this package and only into a TypeScript file.
  The current convention is to import any modules from `@sourcegraph/common`
  into `src/lib/common.ts`, etc.

### Tests

There are no tests yet. It would be great to try out Playwright but it looks
like this depends on getting the production build working first (see below).

### Formatting and linting

This package defines its own rules for formatting (which includes support for
Svelte components) and linting. The workspace rules linting and formatting
commands have not been updated yet to keep this experiment contained.

Run

```sh
pnpm run lint
pnpm run format
```

inside this directory.

There is also the `pnpm run check` command which uses `svelte-check` to validate
TypeScript, CSS, etc in Svelte components. This currently produces many errors
because it also validates imported modules from other packages, and we are not
explicitly marking type-only imports with `type` in other parts of the code
base (which is required by this package).

### Data loading with GraphQL

This project makes use of query composition, i.e. components define their own
data dependencies via fragments, which get composed by their callers and are
eventually being used in a query in a loader.
This approach is not used everywhere (yet), but the goal is that any data
loading happens inside or is enabled by a page/layout loader. This will make it
very clear how the data flows.

There are a couple of issues to consider with this approach and sometimes we'll have to make exceptions:

- Caching: If every loader composes its own query it's possible that two
  queries fetch the same data, in which case we miss out on caching. If caching
  the data is more important than data co-location it might be preferable to
  define a reusable query function. Example: File list for currently opened
  folder (sidebar + folder page)
- Shared data from layout loaders: While it's very convenient that pages have access to any data from the ancestor layout loaders, that doesn't work well with data dependency co-location. The layout loaders don't know which sub-layout or sub-page is loaded and what data it needs.
  Possible solutions (to explore):
  - Keep everything as is (query in higher up loader needs to (explicitly) include every field needed by a descendant layout/page.
  - Have pages/layouts define dedicated fragments to be used for higher level loaders. High level loaders would then request the data for all sub-layouts/-pages, whether it's the current page or not (-> overfetching)
  - Sub-layout/-pages don't rely on higher level loaders and instead fetch all their data on their own (-> duplicate data fetching)

## Production build

A production version of this app can be built with

```sh
pnpm run build
```

Currently SvelteKit is configured to create a client-side single page
application.
