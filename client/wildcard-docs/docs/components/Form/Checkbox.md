---
id: "Checkbox"
title: "Variable: Checkbox"
sidebar_label: "Checkbox"
sidebar_position: 0
custom_edit_url: null
---

• **Checkbox**: `React.FunctionComponent`<[`CheckboxProps`](../types/CheckboxProps)\>

## Overview

Renders a single checkbox.

Checkboxes should be used when a user can select any number of choices from a list of options.
They can often be used stand-alone, for a single option that a user can turn on or off.

Grouped checkboxes should be visually presented together.

Useful article comparing checkboxes to radio buttons: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/

#### Defined in

[client/wildcard/src/components/Form/Checkbox/Checkbox.tsx:17](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Checkbox/Checkbox.tsx#L17)


## Usage
```tsx
import { Checkbox } from '@sourcegraph/wildcard'

<Checkbox
    name={name}
    id={name}
    value="first"
    checked={isChecked}
    onChange={handleChange}
    label="Check me!"
    message="Hello world!"
    {...props}
/>
```

## Props

---

### type
• `Optional` **type**: ``"radio"`` \| ``"checkbox"`` \| undefined

The `<input>` type. Use one of the currently supported types.

#### Defined in

[client/wildcard/src/components/Form/internal/BaseControlInput.tsx:16](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/BaseControlInput.tsx#L16)
