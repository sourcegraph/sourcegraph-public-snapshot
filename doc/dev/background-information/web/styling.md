# Styling UI

Sourcegraph has many UI components. A unique constraint for these is that they need to run within different _environments_.
Wherever a component is running, it should look native to its environment and consistent with the design of that environment.
For example, our hover overlay needs to work and _behave_ the same in the Sourcegraph webapp and in the browser extension, where it is injected into a variety of code hosts, but _look_ native to the environment.
Components need to be able to adapt styles from CSS stylesheets already loaded on the page (no matter how those were architected).

## Goals

1. Components decoupled from styling, that look **consistent** with the host environment.
2. Support light and dark themes.
3. Tooling support:
   - Autocompletion for styles when writing components
   - Autocompletion when writing styles
   - Linting for styles
   - Browser dev tools to easily inspect and iterate on styles
   - Autoprefixer
4. Support for advanced CSS features, like state selectors, pseudo elements, flexbox, grid, media queries, CSS variables, ...

## Environments

### Host-agnostic UI

Components that need to run in different environments (any UI shared between our browser extension and the webapp) adopt styles from their environments through configurable CSS class names (as opposed to trying to replicate the styling with copied CSS).
A component may accept multiple class names for different elements of the component.
An example of this is `<HoverOverlay/>`: see how [the different props it accepts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4047521a92904054e782d341001d08d61945c86f/-/blob/shared/src/hover/HoverOverlay.tsx#L27-39) for its child components' class names are passed in the [webapp](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4047521a92904054e782d341001d08d61945c86f/-/blob/web/src/components/shared.tsx#L35-44) and in [code host integrations](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4047521a92904054e782d341001d08d61945c86f/-/blob/browser/src/shared/code-hosts/shared/codeHost.tsx#L443:1).

### Host-specific UI

In the environments we control ourselves (such as our webapp, the options page of the browser extension, or our marketing website), we use a customized version of [Bootstrap](https://getbootstrap.com/) as a CSS framework.
Any code inside our webapp can and should make use of the CSS classes that Bootstrap provides as building blocks (and should generally do so instead of writing custom styles).
This includes classes like [cards](https://getbootstrap.com/docs/4.5/components/card/), [buttons](https://getbootstrap.com/docs/4.5/components/buttons/) or [input groups](https://getbootstrap.com/docs/4.5/components/input-group/), but also utility classes for [layout](https://getbootstrap.com/docs/4.5/utilities/flex/) and [spacing](https://getbootstrap.com/docs/4.5/utilities/spacing/).
Please refer to the excellent [Bootstrap documentation](https://getbootstrap.com/docs/4.5/) for everything that is available for use.
To see what our customizations look like visually (in both light and dark theme), you can find a [showcase in our Storybook](https://5f0f381c0e50750022dc6bf7-oozlspcdwk.chromatic.com/?path=/story/web-global-styles).

Components only used in a specific host environment do not need to support customization through class names.
They can however utilize environment-agnostic components by passing our Bootstrap classes as custom `className` values.

## Our approach to styling

### Structuring style sheets

A component may need styles that are common to all environments, like internal layout.
We write those styles in SCSS stylesheets that are imported into the host environment.
In some cases these can be overridden by passing another class name for that element.

To avoid naming conflicts we structure these files using the [BEM convention](http://getbem.com/naming/) (Block - Element - Modifier).
The _block_ name is always the React component name, _elements_ and _modifiers_ are used as specified in BEM.
A _block_ must not be referenced in any other React component than the one with the matching name.

Example:

```scss
.some-component {
    // .. styles ...

    &__element {
        // ... styles ...

        &--modifier {
            // ... styles ...
        }
    }
}
```

- **Block**: A React component name in kebab-case. This class is always assigned to the root DOM element of the component.
- **Element**: A sub-element of the component. This should be a name that describes the semantic of this element within the component.
- **Modifier**: A modifier of the _element_, e.g. `--loading` or `--closed`. This is only rarely needed.

Please note that there is no hierarchy in _elements_, as that would couple the styling to the DOM structure. Element names should be unambiguous within their component/_block_, or be split into a separate component/_block_.

We colocate stylesheets next to the component using the same file name.
This approach ensures styles are easy to find and makes it easy to tell which styles apply to which elements (by putting them side-by-side in an editor, or through a simple text search).

Using classes, as opposed to child and descendant selectors, decouples styles from the DOM structure of the component, ensures encapsulation and avoids CSS specificity issues.
Descendant selectors can still be useful in rare cases though, like styling in the browser extension, or styling markdown content.

### Typography

Avoid ever overriding font family, text sizes or text colors.
These are set globally by the host environment for semantic HTML elements, e.g. `<h1>`, `<a>`, `<code>` or `<small>`.

### Colors and theming

The brand color palette is [OpenColor](https://yeun.github.io/open-color/).
In addition to these, we define a blueish grayscale palette for backgrounds, text and borders.
These colors are all available as SCSS and CSS variables.
However, directly referencing these may not work well in both light and dark themes, and may not match code host themes (if the component is shared).
The best approach is to not reference colors at all and use building blocks that have borders, text colors etc defined.
This saves code and makes it easy to maintain design consistency even if we want to change colors in the future.
When that is not possible (for example UI contributed by extensions), prefer to reference CSS variables with semantic colors like `var(--danger)`, `var(--success)`, `var(--border-color)`, `var(--body-bg)` etc.
The values of these variables are changed globally when the theme changes.
Be aware that this means our stylesheets for each host environment need to define these variables too.

Defining different styles in the webapp depending on the `theme-dark` and `theme-light` classes predates CSS variables and is discouraged.

### Spacing

We use `rem` units in all component styling and strive to use `0.25rem` steps.
This ensures our spacing generally aligns with an [8pt grid](https://medium.com/swlh/the-comprehensive-8pt-grid-guide-aa16ff402179), but also gracefully scales in environments that have a different base `rem` size.
In our webapp, it is recommended to make use of [Bootstrap's margin and padding utilities](https://getbootstrap.com/docs/4.5/utilities/spacing/), which are configured to align with the 8pt grid.

### Layout

We use modern CSS for our layouting needs. You can find a [small playground in our Storybook](https://5f0f381c0e50750022dc6bf7-oozlspcdwk.chromatic.com/?path=/story/web-global-styles--layout). The dev tools of modern browsers provide a lot of useful tooling to work with CSS layouts.

Layouts should always be _responsive_ to make sure Sourcegraph is usable with different screen resolutions and window sizes, e.g. when resizing the browser window and using Sourcegraph side-by-side with an editor.

[CSS Flexbox](https://css-tricks.com/snippets/css/a-guide-to-flexbox/) is used for **one-dimensional** layouts (single rows or columns, with optional wrapping). In the webapp, you can use utility classes for simple flexbox layouts and responsive layouts. This is the most common layout method.

For complex **two-dimensional** layouts, [CSS Grid](https://learncssgrid.com/) can be used.
