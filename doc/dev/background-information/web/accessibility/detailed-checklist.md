# Detailed Checklist

Before going through this checklist, you should ensure you have the browser extension [ARC Toolkit](https://chrome.google.com/webstore/detail/arc-toolkit/chdkkkccnlfncngelccgbgfmjebmkmce) installed - it will be useful to help debug accessibility issues.

Note: This checklist is written against the [WCAG 2.1](https://www.w3.org/TR/WCAG21/) AA specification. It should be revised to meet other WCAG versions or grades.

## Headings

We should ensure the following statements are true:
- All Headings use the correct semantic `<hX>` tag (e.g. `<h1>`).
- There is only be one `<h1>` element per page.
- All headings are written in the correct logical sequence.
  - The order should descend, a `<h4>` should not appear on the page before the first `<h3>` element
  - Heading levels should not be skipped. Don't jump from `<h1>` to `<h3>` and skip `<h2>`. See the *tips* section below for how you can fix styling issues caused by this.
- All heading text correctly describes the page/section it is in.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Headings" test group. It will visually identify each heading on the page and raise any warnings.
- Navigate through the journey with a screen reader. Use only the "Navigate to next heading" shortcut.

**Tips:**

- If we have to use a `<h2>` need the styles of a `<h3>`, we can easily select the correct styles using the [Wildcard Heading components](https://storybook.sgdev.org/?path=/story/wildcard-typography-all--simple).
  - Example: `<H3 as={H2}>Hello</H3>`
- If you are still unsure about something, consult the [W3 guide on headings](https://www.w3.org/WAI/tutorials/page-structure/headings/).

## Images / Graphical content

### All `img` OR `svg` elements have an appropiate descriptive attribute

`<img>`:
- If it adds value to the user journey, the `alt` should describe the image.
- If it does not add value (it is purely decorative), the `alt` attribute should explicitly be set as "".
  - Example: `<img alt=""/>`
- If it contains text, the alt description should include the image's text.

`<svg>`:
- If it adds value to the user journey, it should include a `<title>` element within the SVG. This element should be referenced through `aria-labelledby` on the `svg` element.
  - Example:
    ```
      <svg aria-labelledby="svgtitle1">
        <title id="svgtitle1">Settings</title> [other svg code]
      </svg>
    ```

### Videos
- TODO

**How to test:**

- Run an audit in ARC Toolkit. Select the "Images" test group. It will visually identify each image on the page and raise any warnings.
- Navigate through the journey with a screen reader. Watch out for any "unlabelled image" (or similar) readouts.

**Tips:**

- If you are still unsure about something, consult the [W3 guide on images](https://www.w3.org/WAI/tutorials/images/).
- Note these same rules apply to 

## Lists

- Lists of items that are related to each other, should be use the correct list elements.
  - Documentation on each list element: [`<ol>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/ol), [`<ul>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/ul), [`<li>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/li), [`<dl>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/dl).

**How to test:**

- Run an audit in ARC Toolkit. Select the "Lists" test group. It will visually identify each list on the page and raise any warnings. Check that everything that *should* be a list *is* a list.

## User actions

### Navigation actions should use the `<a>` element and be recognizable as link.

- Any action that navigates should use the `<a>` element.
- All `<a>` elements should have a valid `href` attribute that describes where the user will be navigated to.
- All navigation actions should be recognizable as links.
- All navigation actions should have correct focus styles.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Links" test group. It will visually identify each link on the page and raise any warnings. Check that everything that *should* be a link *is* a link.

**Tips**

- TODO. Use the navlink or routerlink Wildcard components

### Other actions should use the `<button>` element and be recognizable as buttons.

- Any action that does not navigate should use the `<button>` element.
- All `<button>` elements should have a valid `type` attribute that describes the action.
- All button elements should be recognizable as buttons.
- All button elements should have correct focus styles.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Buttons" test group. It will visually identify each button on the page and raise any warnings. Check that everything that *should* be a button *is* a button.

**Tips**

- TODO. Use the Button wildcard components

## Tables

- All tabular data should use the [`<table>`](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/table) element and associated child elements such as `th` and `td`.
  - Do we need to display data in rows and columns? Use `<table>`.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Tables" test group. It will visually identify each table on the page and raise any warnings.

**Tips**

- If you are still unsure about something, consult the [W3 guide on images](https://www.w3.org/WAI/tutorials/tables/).


## Forms

- All form inputs should have a corresponding `<label>` element or `aria-label` attribute.
- We should use `<fieldset>` and `<legend>` to group related form elements.
  - Example: A group of radio buttons should be grouped within a `fieldset`
- Any form input errors are displayed in a list above the form after submission. TODO: Check
- Any errors should be correctly associated with the relevant input. Use `aria-describedby` to link. TODO: Check
- It is possible to identify error, warning and success states of the form through text. We cannot assume that a user can identify these through color alone.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Forms" test group. It will visually identify each form and input on the page, show respective labels and raise any warnings.
- Complete the form using a screen reader. Be sure to test all states, including any errors.

**Tips**:
- Use the `<form>`> element? TODO: Check

## Color contrast

- Borders and icons should have a contrast ratio of at least 3:1 against their backgrounds
- Body text should have contrast ratio of at least 4.5:1 against its background.
- Large text should have a contrast ratio of at least 3:1 against its background.

**How to test:**

- Run an audit in ARC Toolkit. Select the "Color Contrast" test group. It will report any contrast violations on the current page.
- Be sure to run this audit in **both** dark and light themes.

## Page Reflow / Sizing + Text spacing


## Mobile devices

- TODO:

## Check page reflow TODO
