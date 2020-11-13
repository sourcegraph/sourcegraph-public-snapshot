# Context key expressions 

**Context key expressions** are a way to express values dynamically in the
fields of a Sourcegraph extension's manifest.

With string interpolation, you can insert these expressions directly into string
fields.

Context keys are like variables that you can use inside of these interpolated
expressions. Context keys give you access to values that are available to your
extension dynamically, such as `resource` for the currently viewed resource, or
`config` for the configuration settings.


## Template interpolation

Context key expressions can be interpolated inside of strings. In manifest
fields that support interpolated expressions, you can interpolate an expression
by surrounding it with `${` and `}` tags.

This syntax for interpolation is based on JavaScript's template interpolation
syntax.


## Supported fields in the manifest

These fields in the manifest support context key expressions.


### In contributed actions

String fields that accept interpolated expressions, in [contributed
actions:](contributions.md#actions)

- `title`
- `category`
- `description`
- `iconURL`
- `actionItem.label`
- `actionItem.description`
- `actionItem.iconURL`
- `actionItem.iconDescription`
- `commandArguments`: each string item in the array accepts interpolated
  expressions.

Fields that expect a context key expression:

- `actionItem.pressed`: renders the action in a pressed state when this
  expression evaluates to true.


### In menu contributions

Fields that expect a context key expression, in [menu
contributions](contributions.md#menus)

- `when`: enables the menu contribution when this expression evaluates to true.


## Available context keys

Context keys are the variables that you can use inside of expressions.

- `config`: a namespace containing all the settings that are available. For
  example, `config.sourcegraphBaseUrl` contains the `sourcegraphBaseUrl` value
  from Sourcegraph settings.
- `resource`: the current resource being viewed, such as a file.
  - `resource.uri`
  - `resource.basename`
  - `resource.dirname`
  - `resource.extname`
  - `resource.language`
  - `resource.type`
- `component`: true if a component is open, such as a panel, directory view, or
  file view.
- `component.selections`: an object representing the current selections.
- `panel`: the panel component, if a panel is open.
  - `panel.activeView.id`
  - `panel.activeView.hasLocations`

## Available operators

The available operators mimic the behavior of the same operators in JavaScript:

- Boolean: `&&` `||` (with the same truthy/falsy semantics as JavaScript)
- Comparison: `==` `!=` `===` `!==` `<` `>` `<=` `>=` (with the strict/loose
  equality rules of JavaScript)
- Arithmetic: `+` `-` `*` `/` `^` `%`
- Unary: `!` `+` `-`

## Available functions

### `get(object, key)`

Returns the value of a property named `key` on `object` or return undefined if
either the object or the property doesn't exist.

### `json(object)`

Returns the object converted to a JSON string using `JSON.stringify`.

## Limitations

The expression syntax is simple and isn't intended to be a full programming
language, so there are some limitations.

- **Lack of operator precedence.** Because of the simplicity of the parser,
  operators do not have any precedence and are simply evaluated left-to-right.
  Use parentheses to specify precedence.
- **Lack of a ternary operator.** Instead of a ternary operator, you can use
  combinations of `&&` and `||` operators to achieve a similar result.

