# Sourcegraph SvelteKit

This folder contains the experimental [SvelteKit](https://kit.svelte.dev/)
implementation of the Sourcegraph app.

**NOTE:** This is a _very early_ prototype and it will change a lot.

## Developing

```bash
# Install dependencies
pnpm install
# Generate GraphQL types
pnpm run -w generate
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
  the app. This isn't really available, and at best will only increase the
  initial loading time. But some modules, especially those that access browser
  specific features during module initialization, can even cause the dev build
  to fail.
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

## Production build

A production version of this app can be built with

```sh
pnpm run build
```

Currently SvelteKit is configured to create a client-side single page
application.

