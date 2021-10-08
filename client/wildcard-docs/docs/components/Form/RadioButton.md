---
id: "RadioButton"
title: "RadioButton"
sidebar_label: "RadioButton"
sidebar_position: 0
custom_edit_url: null
---

• **RadioButton**: `React.FunctionComponent`<`RadioButtonProps`\>

## Overview

Renders a single radio button.

Radio buttons should be used when a user must make a single choice from a list of two or more mutually exclusive options.

Grouped radio buttons should be visually presented together.

Useful article comparing radio buttons to checkboxes: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/

#### Defined in

[client/wildcard/src/components/Form/RadioButton/RadioButton.tsx:22](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/RadioButton/RadioButton.tsx#L22)


## Usage
```tsx
import { RadioButton } from '@sourcegraph/wildcard'

<RadioButton
    id={`${name}-1`}
    name={name}
    value="first"
    checked={selected === 'first'}
    onChange={handleChange}
    label="First"
    message="Hello world!"
    {...props}
/>
<RadioButton
    id={`${name}-2`}
    name={name}
    value="second"
    checked={selected === 'second'}
    onChange={handleChange}
    label="Second"
    message="Hello world!"
    {...props}
/>
<RadioButton
    id={`${name}-3`}
    name={name}
    value="third"
    checked={selected === 'third'}
    onChange={handleChange}
    label="Third"
    message="Hello world!"
    {...props}
/>
```

## Props
___

### name

• **name**: `string`

The name of the radio group. Used to group radio controls together to ensure mutual exclusivity. 
If you do not need this prop, consider if a checkbox is better suited for your use case.

#### Defined in

[client/wildcard/src/components/Form/RadioButton/RadioButton.tsx:10](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/RadioButton/RadioButton.tsx#L10)

___

### type
• `Optional` **type**: ``"radio"`` \| ``"checkbox"`` \| undefined

The `<input>` type. Use one of the currently supported types.

#### Defined in

[client/wildcard/src/components/Form/internal/BaseControlInput.tsx:16](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/BaseControlInput.tsx#L16)
