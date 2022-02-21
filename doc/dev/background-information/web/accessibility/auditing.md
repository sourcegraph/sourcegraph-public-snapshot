# How to conduct an accessibility audit

Before starting an audit, you should ensure you can confidently [control accessibility tools](tooling.md).

## Preparation

Before starting your audit, you should ensure the following statements are true:

- The page or journey you are about to audit feels appropriately scoped.
  - Example: Instead of auditing the entirety of code insights, the audit should be focused on a specific part, such as creating a new code insight.
- A relevant GitHub issue has been created for this audit, and it is attached to the [WCAG 2.1 auditing tracking issue](https://github.com/sourcegraph/sourcegraph/issues/31475).


Needs adding:
- Tooltips. Need to decide if global or local



## Content

### Headings are used correctly and in the correct order

- Headings must use the correct semantic `<hX>` tag (e.g. `<h1>`).
- There should only be one `<h1>` element per page.
- Headings must be written in the correct logical sequence.
  - The order should descend, a `<h4>` should not appear on the page before the first `<h3>` element
  - Heading levels should not be skipped. Don't jump from `<h1>` to `<h3>` and skip `<h2>`. See the *tips* section below for how you can fix styling issues caused by this.
- The heading text correctly describes the page/section it is in.

**Tips**

- If we have to use a `<h2>` need the styles of a `<h3>`, we can easily select the correct styles using the [Wildcard Heading components](https://storybook.sgdev.org/?path=/story/wildcard-typography-all--simple).
  - Example: `<H3 as={H2}>Hello</H3>`
- We should be able to use some tooling to help find this. TODO UPDATE


## Keyboard navigation

### The entire user journey is keyboard navigable

- It is possible for a user to correctly complete the user journey using **only** the keyboard (no mouse).

### The keyboard focus order follows a logical sequence that is predictable

- It should be easy to guess where next

### All interactive elements have a visible focus style

- blah blah

### Invisible elements are never focused

- blajh blah


## Images

### All `img` elements have an appropiate alt attribute

- If an image adds value to the user journey, the `alt` should describe the image.
- If an image does not add value (it is purely decorative), the `alt` attribute should explicitly be set as "".
  - Example: `<img alt=""/>`
- For images containing text, the alt description should include the image's text.

TODO: Check what we need to do for `svg` elements

### Ensure that complex images such as charts and graphs have an appropriate text alternative

TODO: Understand this more. Useful link: https://www.a11yproject.com/checklist/#provide-a-text-alternative-for-complex-images-such-as-charts-graphs-and-maps

## Lists

TODO Update:

## User actions

### Navigation actions should use the `<a>` element and be recognizable as link.

- Any action that navigates should use the `<a>` element.
- All `<a>` elements should have a valid `href` attribute that describes where the user will be navigated to.
- All navigation actions should be recognizable as links.
- All navigation actions should have correct focus styles.

**Tips**

- TODO. Use the navlink or routerlink Wildcard components

### Other actions should use the `<button>` element and be recognizable as buttons.

- Any action that does not navigate should use the `<button>` element.
- All `<button>` elements should have a valid `type` attribute that describes the action.
- All button elements should be recognizable as buttons.
- All button elements should have correct focus styles.

**Tips**

- TODO. Use the Button wildcard components


## Tables

