# Accessibility

People of all abilities should have first-class access to code and coding. In order to achieve this it is vital that we build our application to be accessible to all. All client code should strive to [meet our accessibility standards](https://about.sourcegraph.com/handbook/product/design/design-and-interaction-guidelines#accessibility-standards).

## What this means for development

It shouldn't be difficult for us to build accessible frontends providing we follow these key points:

- Challenge designs. Accessibility outweighs aesthetics. If you don't believe a design is accessible then raise this.
- Write semantic HTML. This is the foundation of good accessibility, assistive technology is built to understand these semantics.
- Use labels. Don't assume the user will be able to infer the meaning of an image or input based on its visual design.
- Use ARIA where required. Sometimes semantics aren't enough, ARIA attributes can be used to ensure content is still accessible. [See this guide for using ARIA](https://www.w3.org/TR/aria-in-html/).
- Manually test for accessibility. Ensure your frontend code can be navigated easily using a keyboard. [Learn to effectively use a screen reader](https://www.tpgi.com/basic-screen-reader-commands-for-accessibility-testing/) and check for issues. Read about the importance of [manual accessibility testing](https://www.smashingmagazine.com/2018/09/importance-manual-accessibility-testing/).
- Make good use of the Wildcard component library. These components are designed and build specifically with accessibility in mind. Often it may be better utilise these components instead of reimplementing a similar design.

### Tooling

There is a lot of useful tooling to help us catch and fix acccessibility issues in our code.

- [eslint-plugin-jsx-a11y](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y). Statically analyze our JSX for potential accessibility violations.
- [@storybook/addon-a11y](https://storybook.js.org/addons/@storybook/addon-a11y). This addon uses [axe-core](https://github.com/dequelabs/axe-core) to audit rendered components and raise accessibility issues. It also has some useful features to simulate vision impairments such as blurred vision and color blindness.
- [Lighthouse](https://developers.google.com/web/tools/lighthouse) allows you to run accessibility audits anywhere.
- Browser Accessibility DevTools in [Chrome](https://developers.google.com/web/tools/chrome-devtools/accessibility/reference) and [Firefox](https://developer.mozilla.org/en-US/docs/Tools/Accessibility_inspector).

## Further Reading

- [Accessibility - React](https://reactjs.org/docs/accessibility.html)
- [Accessibility - MDN](https://developer.mozilla.org/en-US/docs/Learn/Accessibility)
- [A complete guide to accessible front-end components](https://www.smashingmagazine.com/2021/03/complete-guide-accessible-front-end-components/)
