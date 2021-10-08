---
id: "Button"
title: "Button"
sidebar_label: "Button"
sidebar_position: 0
custom_edit_url: null
---

• **Button**: `React.FunctionComponent`<[`ButtonProps`](../interfaces/ButtonProps)\>

## Overview

Simple button.

Style can be configured using different button `variant`s.

Buttons should be used to allow users to trigger specific actions on the page.
Always be mindful of how intent is signalled to the user when using buttons. We should consider the correct button `variant` for each action.

Some examples:
- The main action a user should take on the page should usually be styled with the `primary` variant.
- Other additional actions on the page should usually be styled with the `secondary` variant.
- A destructive 'delete' action should be styled with the `danger` variant.

Tips:
- Avoid using button styling for links where possible. Buttons should typically trigger an action, links should navigate to places.

#### Defined in

[client/wildcard/src/components/Button/Button.tsx:42](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Button/Button.tsx#L42)


## Usage
```tsx
import { Button } from '@sourcegraph/wildcard'

<Button variant="primary" size="sm" outline="true">Accept</Button>
```

## Props
___

### as

• `Optional` **as**: `ElementType`<`any`\>

Used to change the element that is rendered.
Useful if needing to style a link as a button, or in certain cases where a different element is required.
Always be mindful of potentially accessibility pitfalls when using this!
Note: This component assumes `HTMLButtonElement` types, providing a different component here will change the potential types that can be passed to this component.

#### Defined in

[client/wildcard/src/components/Button/Button.tsx:20](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Button/Button.tsx#L20)

___

### outline

• `Optional` **outline**: `boolean`

Modifies the button style to have a transparent/light background and a more pronounced outline.

#### Defined in

[client/wildcard/src/components/Button/Button.tsx:15](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Button/Button.tsx#L15)

___

### size

• `Optional` **size**: ``"sm"`` \| ``"lg"``

Allows modifying the size of the button. Supports larger or smaller variants.

#### Defined in

[client/wildcard/src/components/Button/Button.tsx:13](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Button/Button.tsx#L13)

___

### variant

• `Optional` **variant**: ``"primary"`` \| ``"secondary"`` \| ``"success"`` \| ``"danger"`` \| ``"warning"`` \| ``"info"`` \| ``"merged"`` \| ``"link"``

The variant style of the button. Defaults to `primary`

#### Defined in

[client/wildcard/src/components/Button/Button.tsx:11](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Button/Button.tsx#L11)

___
