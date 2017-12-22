# Sourcegraph WebApp

## Components

* Try to do one component per file. This makes it easy to encapsulate corresponding styles.
* Don't pass arrow functions as React bindings unless unavoidable

## Styles

* Styles are written in SCSS
* Every component .tsx file should have a corresponding stylesheet named like the .tsx file
  * The stylesheet must contain a top-level selector to scope it to a class that is the kebab-case version of the component name.
    The component must apply that class to its top-level element.
* Use [BEM](http://getbem.com/). "Block" here is the component name, element a non-component child of the component.
* Only use descendent/child selectors where unavoidable. Prefer BEM-style class names that are nested in SCSS through the `&` operator
* Create utility classes for styles that should be shared horizontally between components
* Always use `rem` units (when converting designs, `1rem` = `16px`). This allows us to scale the whole UI by modifying the root font size.
* Avoid hardcoding colors, use SCSS variables if they are available / the color makese sense to share.
* Try to _minimize_ the usage of advanced SCSS features. They can lead to bugs and complicate styles.
  * Encouraged features are nesting and imports (which is the intersection of Less', SCSS' and PostCSS' feature set)
* Think about mobile at least so much that no feature breaks when the browser window is resized
* Don't couple the styles to the DOM structure. It should be possible to change the structure without changing the styles and vice versa.
* Prefer flexbox over absolute positioning
* Avoid styling the children of your components. This couples your component to the implementation of the child
* Order your rules so that layout rules (that describe how the component is layed out to its parents) come first, then rules that describe the layout of its children, and finally visual details.

## Formatting

We use [Prettier](https://github.com/prettier/prettier) so you never have to worry about how to format your code.
`npm run prettier` will check & autoformat all code. It is also run as part of `npm run lint`.

## Tests

* We write unit tests and e2e tests.
* Unit tests are for things that can be tested in isolation; you provide inputs and make assertion on the outputs and/or side effects.
* Run unit tests via `npm run test`.
* E2E tests are for the whole app: JS, CSS, and backend. These tests require hitting a backend like https://sourcegraph.com or https://sourcegraph.sgdev.org (default http://localhost:3080).
* Run E2E tests via `npm run test-e2e`.
* E2E tests send messages to a chrome debugger port (9222), telling chrome to do things like "go to this URL" and "click on this selector" and "execute this JavaScript in the page".
* `npm run test-e2e` will automatically start a headless chrome process; to prevent that, set the environment variable `SKIP_LAUNCH_CHROME=t`
* E2E caveats:
  * don't overdo them; they are the tip of the testing pyramid and the mass of tests should be at lower layers
  * avoid coupling your tests too tightly to the implementation (e.g. with strict selectors that are coupled to the DOM structure), or we will have many test failures
    * for example, mark your E2E element targets with a class like `.e2e-j2d-button`
  * if you change a lot about the UI you should run e2e tetss before committing b/c you will likely break something.
