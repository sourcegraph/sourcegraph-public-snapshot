---
id: "Grid"
title: "Grid"
sidebar_label: "Grid"
sidebar_position: 0
custom_edit_url: null
---

• **Grid**: `React.FunctionComponent`<[`GridProps`](../interfaces/GridProps)\>

## Overview

A simple Grid component. Can be configured to display a number of columns with different gutter spacing.

#### Defined in

[components/Grid/Grid.tsx:30](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Grid/Grid.tsx#L30)

## Usage

```tsx
import { Grid } from '@sourcegraph/wildcard'

<Grid columnCount={columnCount} spacing={spacing}>
    <div>1</div>
    <div>2</div>
    <div>3</div>
</Grid>

```

## Props

### className

• `Optional` **className**: `string`

#### Defined in

[components/Grid/Grid.tsx:4](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Grid/Grid.tsx#L4)

___

### columnCount

• `Optional` **columnCount**: `number`

The number of grid columns to render.

**`default`** 3

#### Defined in

[components/Grid/Grid.tsx:10](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Grid/Grid.tsx#L10)

___

### spacing

• `Optional` **spacing**: `number`

Rem Spacing between grid columns.

**`default`** 1

#### Defined in

[components/Grid/Grid.tsx:16](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Grid/Grid.tsx#L16)
