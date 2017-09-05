
# Sourcegraph WebApp

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
 - Never hardcode colors, use SCSS variables
 - Try to _minimize_ the usage of advanced SCSS features. They can lead to bugs and complicate styles.
   - Encouraged features are nesting and imports (which is the intersection of Less', SCSS' and PostCSS' feature set)
 - Think about mobile at least so much that no feature breaks when the browser window is resized
 - Don't couple the styles to the DOM structure. It should be possible to change the structure without changing the styles and vice versa.
 - Prefer flexbox over absolute positioning
 - Avoid styling the children of your components. This couples your component to the implementation of the child
 - Order your rules so that layout rules (that describe how the component is layed out to its parents) come first, then rules that describe the layout of its children, and finally visual details.
