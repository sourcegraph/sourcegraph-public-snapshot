# Developing the Sourcegraph web app

Guide to contribute to the Sourcegraph web app. Please also see our general [TypeScript documentation](../languages/typescript.md).

## Local development

See [common commands for local development](../../setup/quickstart.md).
Commands specifically useful for the web team can be found in the root [package.json](https://github.com/sourcegraph/sourcegraph/blob/main/package.json).
Also, check out the web app [README](https://github.com/sourcegraph/sourcegraph/blob/main/client/web/README.md).

### Prerequisites

The `sg` CLI tool is required for key local development commands. Check out [the `sg` documentation](../sg/index.md).

To install it, [see the instructions](../../setup/quickstart.md).

### Commands

1. Start the web server and point it to any deployed API instance. See more info in the web app [README](https://github.com/sourcegraph/sourcegraph/blob/main/client/web/README.md).

    ```sh
    sg start web-standalone
    ```

    For the open-source version:

    ```sh
    sg start oss-web-standalone
    ```

    To use a public API that doesn't require authentication for most of the functionality:

    ```sh
    SOURCEGRAPH_API_URL=https://sourcegraph.com sg start web-standalone
    ```

    For open-source version:

    ```sh
    SOURCEGRAPH_API_URL=https://sourcegraph.com sg start oss-web-standalone
    ```

2. Start all backend services with the frontend server.

    ```sh
    sg start # which defaults to `sg start enterprise`
    ```

    For the open-source version:

    ```sh
    sg start oss
    ```

3. Regenerate GraphQL schema, Typescript types for GraphQL operations and CSS Modules.

    ```sh
    yarn generate
    ```

    To regenerate on file change:

    ```sh
    yarn watch-generate
    ```

### Storybook

Storybook is used to work on the components in isolation. The latest version is deployed at http://storybook.sgdev.org/.

To use it locally, use `yarn storybook` command to start the Storybook development server. This will load stories from all the workspaces that we have in the monorepo.

To boost the build/recompilation performance of the Storybook, it's possible to load only a subset of stories needed for the current feature implementation. This is done via the environment variable `STORIES_GLOB`:

```sh
STORIES_GLOB='client/web/src/**/*.story.tsx' yarn workspace @sourcegraph/storybook run start
```

It's common for a developer to work only in one client workspace, e.g., `web` or `browser`.
The root `package.json` has commands to launch Storybook only for each individual workspace, which greatly increases the build performance.

```sh
yarn storybook:branded
yarn storybook:browser
yarn storybook:shared
yarn storybook:web
yarn storybook:wildcard
```

## Naming files

If the file only contains one main export (e.g. a component class + some interfaces), name the file like the main export.
This name is PascalCase if the main export is a class.
This makes it easy to find it with file search.
If the file has no main export (e.g. a file with utility functions), give the file a name that groups the exports semantically.
This name is usually short and lowercase or kebab-case, e.g. `util/errors.ts` contains error utilities.
Avoid adding utilities into a `util.ts` file, it is doomed to become a mess over time.

### `.ts` vs `.tsx`

You must use the `tsx` file extension if the file contains TSX (React) syntax.
You should use the normal `ts` extension if it does not.
The `tsx` extension makes certain generic syntax impossible and also enables emmet suggestions in editors, which are annoying in normal TypeScript code.

### `index.*` files

Index files should never contain declarations on their own.
Their purpose is to reexport symbols from a number of other files to make imports easier and define the the public API.

## Components

- Try to do one component per file. This makes it easy to encapsulate corresponding styles.
- Don't pass arrow functions as React bindings unless unavoidable

## Code splitting

[Code splitting](https://reactjs.org/docs/code-splitting.html) refers to the practice of bundling the frontend code into multiple JavaScript files, each with a subset of the frontend code, so that the client only needs to download and parse/run the code necessary for the page they are viewing. We use the [react-router code splitting technique](https://reactjs.org/docs/code-splitting.html#route-based-code-splitting) to accomplish this.

When adding a new route (and new components), you can opt into code splitting for it by referring to a lazy-loading reference to the component (instead of a static import binding of the component). To create the lazy-loading reference to the component, use `React.lazy`, as in:

``` typescript
const MyComponent = React.lazy(async () => ({
    default: (await import('./path/to/MyComponent)).MyComponent,
}))
```

If you don't do this (and just use a normal `import`), it will still work, but it will increase the initial bundle size for all users.

(It is necessary to return the component as the `default` property of an object because `React.lazy` is hard-coded to look there for it. We could avoid this extra work by using default exports, but we chose not to use default exports ([reasons](https://blog.neufund.org/why-we-have-banned-default-exports-and-you-should-do-the-same-d51fdc2cf2ad)).)

## Formatting

We use [Prettier](https://github.com/prettier/prettier) so you never have to worry about how to format your code.
`yarn run format` will check & autoformat all code.

## Tests

We write unit tests and e2e tests.

### Unit tests

Unit tests are for things that can be tested in isolation; you provide inputs and make assertion on the outputs and/or side effects.

React component snapshot tests are a special kind of unit test that we use to test React components. See "[React component snapshot tests](../../how-to/testing.md#react-component-snapshot-tests)" for more information.

You can run unit tests via `yarn test` (to run all) or `yarn test --watch` (to run only tests changed since the last commit). See "[Testing](../../how-to/testing.md)" for more information.

### E2E tests

See [testing.md](../../how-to/testing.md).
