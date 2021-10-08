---
id: "PageSelector"
title: "PageSelector"
sidebar_label: "PageSelector"
sidebar_position: 0
custom_edit_url: null
---

• **PageSelector**: `React.FunctionComponent`<[`PageSelectorProps`](../interfaces/PageSelectorProps)\>

## Overview

PageSelector should be used to render offset-pagination controls.
It is a controlled-component, the `currentPage` should be controlled by the consumer.

## Usage
```tsx
import { PageSelector } from '@sourcegraph/wildcard'
const [page, setPage] = useState(1)

<PageSelector currentPage={page} onPageChange={setPage} totalPages={10} />
```

#### Defined in

[components/PageSelector/PageSelector.tsx:67](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/PageSelector/PageSelector.tsx#L67)

## Props

### className

• `Optional` **className**: `string`

#### Defined in

[components/PageSelector/PageSelector.tsx:38](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/PageSelector/PageSelector.tsx#L38)

___

### currentPage

• **currentPage**: `number`

Current active page

#### Defined in

[components/PageSelector/PageSelector.tsx:33](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/PageSelector/PageSelector.tsx#L33)

___

### totalPages

• **totalPages**: `number`

Maximum pages to use

#### Defined in

[components/PageSelector/PageSelector.tsx:37](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/PageSelector/PageSelector.tsx#L37)

## Methods

### onPageChange

▸ **onPageChange**(`page`): `void`

Fired on page change

#### Parameters

| Name | Type |
| :------ | :------ |
| `page` | `number` |

#### Returns

`void`

#### Defined in

[components/PageSelector/PageSelector.tsx:35](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/PageSelector/PageSelector.tsx#L35)

