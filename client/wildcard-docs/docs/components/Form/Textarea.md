---
id: "TextArea"
title: "TextArea"
sidebar_label: "TextArea"
sidebar_position: 0
custom_edit_url: null
---

• **TextArea**: `ForwardRefExoticComponent`<`TextAreaProps` & `RefAttributes`<`HTMLTextAreaElement`\>\>

## Overview

Displays a textarea with description, error message, visual invalid and valid states.

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:24](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L24)


## Usage

```tsx
import { TextArea } from '@sourcegraph/wildcard'

<TextArea
    value=""
    label="Disabled example"
    disabled={true}
    message="This is helper text as needed."
    placeholder="Please type here..."
/>
```

## Props
___

### label

• `Optional` **label**: string

Title of textarea. Used as label

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:8](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L8)

---

### message

• `Optional` **message**: ReactNode

Description block shown below the textarea

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:10](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L10)

---

### className

• `Optional` **className**: string

Custom class name for root label element.

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:12](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L12)

---

### isError

• `Optional` **isError**: Boolean

Define an error in the textarea

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:14](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L14)

---

### disabled

• `Optional` **disabled**: Boolean

Disable textarea behavior

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:16](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L16)

---

### size
• `Optional` **size**: ``"regular"`` \| ``"small"``

Determines the size of the textarea

#### Defined in

[client/wildcard/src/components/Form/TextArea/TextArea.tsx:18](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L18)
