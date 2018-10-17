# Developing the Sourcegraph web app

Guide to contribute to the Sourcegraph webapp.

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

Index files should not never contain declarations on their own.
Their purpose is to reexport symbols from a number of other files to make imports easier and define the the public API.

## Components

- Try to do one component per file. This makes it easy to encapsulate corresponding styles.
- Don't pass arrow functions as React bindings unless unavoidable

## Styles

- Styles are written in SCSS
- Every component .tsx file should have a corresponding stylesheet named like the .tsx file
  - The stylesheet must contain a top-level selector to scope it to a class that is the kebab-case version of the component name.
    The component must apply that class to its top-level element.
- Use [BEM](http://getbem.com/). "Block" here is the component name, element a non-component child of the component.
- Only use descendent/child selectors where unavoidable. Prefer BEM-style class names that are nested in SCSS through the `&` operator
- Create utility classes for styles that should be shared horizontally between components
- Always use `rem` units (when converting designs, `1rem` = `16px`). This allows us to scale the whole UI by modifying the root font size.
- Avoid hardcoding colors, use SCSS variables if they are available / the color makese sense to share.
- Try to _minimize_ the usage of advanced SCSS features. They can lead to bugs and complicate styles.
  - Encouraged features are nesting and imports (which is the intersection of Less', SCSS' and PostCSS' feature set)
- Think about mobile at least so much that no feature breaks when the browser window is resized
- Don't couple the styles to the DOM structure. It should be possible to change the structure without changing the styles and vice versa.
- Prefer flexbox over absolute positioning
- Avoid styling the children of your components. This couples your component to the implementation of the child
- Order your rules so that layout rules (that describe how the component is laid out to its parents) come first, then rules that describe the layout of its children, and finally visual details.

### Theming

Theming is done through toggling top-level CSS classes `theme-light` and `theme-dark`.
Any style can be made different on either theme by scoping it to one of those two classes.
Where possible, we use CSS variables, but unfortunately they don't work with compile-time color manipulation (`darken()` etc)
and runtime color manipulation is not yet implemented in CSS (coming in CSS Color Level 4).

## Formatting

We use [Prettier](https://github.com/prettier/prettier) so you never have to worry about how to format your code.
`yarn run prettier` will check & autoformat all code.

## Tests

- We write unit tests and e2e tests.
- Unit tests are for things that can be tested in isolation; you provide inputs and make assertion on the outputs and/or side effects.
- Run unit tests via `yarn run test`.
- E2E tests are for the whole app: JS, CSS, and backend. These tests require hitting a backend like https://sourcegraph.com or https://sourcegraph.sgdev.org (default `SOURCEGRAPH_BASE_URL=http://localhost:3080`).
- Run E2E tests via `yarn run test-e2e`.
- E2E tests send messages to a chrome debugger port (9222), telling chrome to do things like "go to this URL" and "click on this selector" and "execute this JavaScript in the page".
- `yarn run test-e2e` will automatically start a headless chrome process; to prevent that, set the environment variable `SKIP_LAUNCH_CHROME=t`
- E2E caveats:
  - don't overdo them; they are the tip of the testing pyramid and the mass of tests should be at lower layers
  - avoid coupling your tests too tightly to the implementation (e.g. with strict selectors that are coupled to the DOM structure), or we will have many test failures
    - for example, mark your E2E element targets with a class like `.e2e-j2d-button`
  - if you change a lot about the UI you should run e2e tetss before committing b/c you will likely break something.
