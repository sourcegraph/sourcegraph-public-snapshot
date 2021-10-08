---
id: "Tabs"
title: "Variable: Tabs"
sidebar_label: "Tabs"
sidebar_position: 0
custom_edit_url: null
---

• **Tabs**: `React.FunctionComponent`<[`TabsProps`](../interfaces/TabsProps)\>

## Overview

reach UI tabs component with steroids, this tabs handles how the data should be loaded
in terms of a11y tabs are following all the WAI-ARIA Tabs Design Pattern.

See: https://reach.tech/tabs/

#### Defined in

[client/wildcard/src/components/Tabs/Tabs.tsx:48](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Tabs/Tabs.tsx#L48)

## Usage

:::caution

Tabs component is a compound of components that should be used together.
Don't use these components in isolation.

:::

```tsx
<Tabs lazy={lazy} behavior={behavior} size={size} {...props}>
    <TabList actions={actions ? <div>custom component rendered</div> : null}>
        <Tab>Tab 1</Tab>
        <Tab>Tab 2</Tab>
    </TabList>
    <TabPanels>
        <TabPanel>Panel 1</TabPanel>
        <TabPanel>Panel 2</TabPanel>
    </TabPanels>
</Tabs>
```

## Hierarchy

- `ReachTabsProps`

- `TabsState`

  ↳ **`TabsProps`**

## Properties

### behavior

• `Optional` **behavior**: ``"memoize"`` \| ``"forceRender"``

This prop is lazy dependant, only should be used when lazy is true
memoize: Once a selected tabPanel is rendered this will keep mounted
forceRender: Each time a tab is selected the associated tabPanel is mounted
and the rest is unmounted

#### Inherited from

TabsState.behavior

#### Defined in

[client/wildcard/src/components/Tabs/useTabs.ts:18](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Tabs/useTabs.ts#L18)

___

### children

• **children**: `ReactNode` \| (`props`: `TabsContextValue`) => `ReactNode`

Tabs expects `<TabList>` and `<TabPanels>` as children. The order doesn't
matter, you can have tabs on the top or the bottom. In fact, you could have
tabs on both the bottom and the top at the same time. You can have random
elements inside as well.

You can also pass a render function to access data relevant to nested
components.

**`see`** Docs https://reach.tech/tabs#tabs-children

#### Inherited from

ReachTabsProps.children

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:50

___

### className

• `Optional` **className**: `string`

#### Defined in

[client/wildcard/src/components/Tabs/Tabs.tsx:23](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Tabs/Tabs.tsx#L23)

___

### defaultIndex

• `Optional` **defaultIndex**: `number`

Starts the tabs at a specific index.

**`see`** Docs https://reach.tech/tabs#tabs-defaultindex

#### Inherited from

ReachTabsProps.defaultIndex

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:77

___

### index

• `Optional` **index**: `number`

Like form inputs, a tab's state can be controlled by the owner. Make sure
to include an `onChange` as well, or else the tabs will not be interactive.

**`see`** Docs https://reach.tech/tabs#tabs-index

#### Inherited from

ReachTabsProps.index

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:57

___

### keyboardActivation

• `Optional` **keyboardActivation**: `TabsKeyboardActivation`

Describes the activation mode when navigating a tablist with a keyboard.
When set to `"auto"`, a tab panel is activated automatically when a tab is
highlighted using arrow keys. When set to `"manual"`, the user must
activate the tab panel with either the `Spacebar` or `Enter` keys. Defaults
to `"auto"`.

**`see`** Docs https://reach.tech/tabs#tabs-keyboardactivation

#### Inherited from

ReachTabsProps.keyboardActivation

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:67

___

### lazy

• `Optional` **lazy**: `boolean`

#### Inherited from

TabsState.lazy

#### Defined in

[client/wildcard/src/components/Tabs/useTabs.ts:11](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Tabs/useTabs.ts#L11)

___

### orientation

• `Optional` **orientation**: `TabsOrientation`

Allows you to switch the orientation of the tabs relative to their tab
panels. This value can either be `"horizontal"`
(`TabsOrientation.Horizontal`) or `"vertical"`
(`TabsOrientation.Vertical`). Defaults to `"horizontal"`.

**`see`** Docs https://reach.tech/tabs#tabs-orientation

**`see`** MDN  https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Logical_Properties

#### Inherited from

ReachTabsProps.orientation

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:87

___

### readOnly

• `Optional` **readOnly**: `boolean`

**`see`** Docs https://reach.tech/tabs#tabs-readonly

#### Inherited from

ReachTabsProps.readOnly

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:71

___

### size

• **size**: ``"small"`` \| ``"medium"`` \| ``"large"``

#### Inherited from

TabsState.size

#### Defined in

[client/wildcard/src/components/Tabs/useTabs.ts:8](https://github.com/sourcegraph/sourcegraph/blob/49e75f130e/client/wildcard/src/components/Tabs/useTabs.ts#L8)

## Methods

### onChange

▸ `Optional` **onChange**(`index`): `void`

Calls back with the tab index whenever the user changes tabs, allowing your
app to synchronize with it.

**`see`** Docs https://reach.tech/tabs#tabs-onchange

#### Parameters

| Name | Type |
| :------ | :------ |
| `index` | `number` |

#### Returns

`void`

#### Inherited from

ReachTabsProps.onChange

#### Defined in

node_modules/@reach/tabs/dist/index.d.ts:94
