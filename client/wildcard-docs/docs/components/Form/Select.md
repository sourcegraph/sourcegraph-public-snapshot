---
id: "Select"
title: "Select"
sidebar_label: "Select"
sidebar_position: 0
custom_edit_url: null
---

• **Select**: `React.FunctionComponent`<`SelectProps`>

## Overview

A wrapper around the `<select>` element.
Supports both native and custom styling.

Select should be used to provide a user with a list of options within a form.

Please note that this component takes `<option>` elements as children. This is to easily support advanced functionality such as usage of `<optgroup>`.

#### Defined in

[client/wildcard/src/components/Form/Select/Select.tsx:45](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Select/Select.tsx#L45)

## Usage

```tsx
import { Select } from '@sourcegraph/wildcard'

<Select
    label="What is your favorite fruit?"
    message="I am a message"
    value={selected}
    onChange={handleChange}
>
    <option value="">Favorite fruit</option>
    <option value="apples">Apples</option>
    <option value="bananas">Bananas</option>
    <option value="oranges">Oranges</option>
</Select>
```

## Props
___

### isCustomStyle

• `Optional` **isCustomStyle**: `boolean`

Use the Bootstrap custom `<select>` styles.

#### Defined in

[client/wildcard/src/components/Form/Select/Select.tsx:16](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Select/Select.tsx#L16)

---

### selectSize

• `Optional` **selectSize**: ``"sm"`` \| ``"lg"``

Optional size modifier to render a smaller or larger `<select>` variant

#### Defined in

[client/wildcard/src/components/Form/Select/Select.tsx:20](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Select/Select.tsx#L20)

---

### className
• `Optional` **className**: string | undefined

#### Defined in

[client/wildcard/src/components/Form/internal/AccessibleFieldType.ts:4](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/AccessibleFieldType.ts#L4)

---

### isValid

• `Optional` **isValid**: Boolean

Used to control the styling of the field and surrounding elements.
Set this value to `false` to show invalid styling.
Set this value to `true` to show valid styling.

#### Defined in

[client/wildcard/src/components/Form/internal/AccessibleFieldType.ts:10](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/AccessibleFieldType.ts#L10)
     
---

### message

• `Optional` **message**: ReactNode

Optional message to display below the form field.
This should typically be used to display additional information to the user.
It will be styled differently if `isValid` is truthy or falsy.

#### Defined in

[client/wildcard/src/components/Form/internal/AccessibleFieldType.ts:16](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/AccessibleFieldType.ts#L16)

--- 

### label

• **label**: ReactNode

Descriptive text rendered within a `<label>` element

#### Defined in

[client/wildcard/src/components/Form/internal/AccessibleFieldType.ts:23](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/AccessibleFieldType.ts#L23)

---

### id

• **id**: string

A unique ID for the field element. This is required to correctly associate the rendered `<label>` with the field.

#### Defined in

[client/wildcard/src/components/Form/internal/AccessibleFieldType.ts:27](https://github.com/sourcegraph/sourcegraph/blob/d9f5113a0630d58253462f82ff6e62a4f9d7391c/client/wildcard/src/components/Form/internal/AccessibleFieldType.ts#L27)
