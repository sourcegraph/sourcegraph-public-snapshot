---
id: "Input"
title: "Input"
sidebar_label: "Input"
sidebar_position: 0
custom_edit_url: null
---


• **Input**: `ForwardReferenceComponent`<``"input"``, `InputProps`\>

## Overview

Displays the input with description, error message, visual invalid and valid states.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:31](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L31)

## Usage
```tsx
import { Input } from '@sourcegraph/wildcard'

<Input
    value={selected}
    label="Input small"
    onChange={handleChange}
    message="random message"
    status="valid"
    disabled={false}
    placeholder="testing this one"
    variant="small"
/>
```

## Props

### label

• `Optional` **label**: `string`

text label of input.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:11](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L11)

---

### message
• `Optional` **message**: `ReactNode`

Description block shown below the input.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:13](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L13)

---

### className
• `Optional` **className**: `string`

Custom class name for input element.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:15](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L15)

---

### inputClassName
• `Optional` **inputClassName**: `string`

Custom class name for input element.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:17](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L17)

---

### inputSymbol
• `Optional` **inputSymbol**: `ReactNode`

Input icon (symbol) which render right after the input element.

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:19](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L19)

---

### status
• `Optional` **status**: ``"error"`` \| ``"loading"`` \| ``"valid"``

Exclusive status

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:21](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L21)

---

### disabled
• `Optional` **disabled**: `Boolean`

Disable input behavior

#### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:23](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L23)

---

### variant

• `Optional` **variant**: ``"regular"`` \| ``"small"``

 Determines the size of the input

 #### Defined in

[client/wildcard/src/components/Form/Input/Input.tsx:25](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/Input/Input.tsx#L25)
