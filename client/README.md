# Frontend packages

## List

- **web**: The web-application deployed to http://sourcegraph.com/
- **browser-extensions**: The Sourcegraph browser-extension adds tooltips to code on different code-hosts. [Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en).
- **eslint-plugin-sourcegraph**: Not published package with custom ESLint rules for Sourcegraph. Isn't intended for reuse by other repositories in the Sourcegraph org.
- **extension-api**: The package with types for the [Sourcegraph extension API](https://unpkg.com/sourcegraph/dist/docs/index.html) ([`sourcegraph.d.ts`](https://github.com/sourcegraph/sourcegraph/blob/main/packages/extension-api/src/sourcegraph.d.ts)). Published as `sourcegraph`.
- **extension-api-types**: The Sourcegraph extension API types for client applications. Published as `@sourcegraph/extension-api-types`.
- **ui-kit-legacy-shared**: Contains common TypeScript/React/SCSS client code shared between the browser extension and the web app. Everything in this package is code-host agnostic.
- **ui-kit-legacy-branded**: Contains React components and implements the visual design language we use across our web app and e.g. in the options menu of the browser extension.
- **ui-kit**: Package that encapsulates storybook configuration and contains newly added UI-kit components. Components added to this package should obey to latest rules and conventions added as a part of improving frontend infrastructure.
- **utils**: Package to share utility functions across monorepo packages.

## Further migration plan

1. Content of packages **ui-kit-legacy-shared** and **ui-kit-legacy-branded** should be moved to **ui-kit** and refactored using the latest FE rules and conventions. Having different packages clearly communicates the migration plan. Developers first should look for components in the **ui-kit** package and then fall-back to **legacy** packages if **ui-kit** doesn't have the solution to their problem yet.

2. **ui-kit-legacy-shared** contains utility functions, types, polyfills, etc which is not a part of the UI-kit. These modules should be moved into **utils** package or other new packages which are not yet created: e.g. **api** for GraphQL client and type generators, etc.

3. Packages should use package name (e.g. `@sourcegraph/ui-kit/src/components/Markdown`) for imports instead of the relative paths (e.g. `../../../../ui-kit/src/components/Markdown`) to avoid long relative-paths and make dependency graph between packages clear. (Typescript will warn if packages have circular dependencies). It's easy to refactor such isolated packages, extract functionality into new ones, or even into new repositories.

4. **build** or **config** package should be added later to encapsulate all the configurations reused between packages which will allow removing `jest.config`, `babel.config` from the root of the repo.

5. Probably **foreach-ts-project.sh** can be deprecated in favour of `yarn workspaces run`.
