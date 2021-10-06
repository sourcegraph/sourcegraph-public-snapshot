---
id: "modules"
title: "@sourcegraph/wildcard"
sidebar_label: "Exports"
sidebar_position: 0.5
custom_edit_url: null
---

## Variables

### Button

• **Button**: `React.FunctionComponent`<`ButtonProps`\>

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

[components/Button/Button.tsx:47](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Button/Button.tsx#L47)

___

### Checkbox

• **Checkbox**: `React.FunctionComponent`<`CheckboxProps`\>

Renders a single checkbox.

Checkboxes should be used when a user can select any number of choices from a list of options.
They can often be used stand-alone, for a single option that a user can turn on or off.

Grouped checkboxes should be visually presented together.

Useful article comparing checkboxes to radio buttons: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/

#### Defined in

[components/Form/Checkbox/Checkbox.tsx:17](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Form/Checkbox/Checkbox.tsx#L17)

___

### Container

• **Container**: `React.FunctionComponent`<`Props`\>

A container wrapper. Used for grouping content together.

#### Defined in

[components/Container/Container.tsx:11](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Container/Container.tsx#L11)

___

### Grid

• **Grid**: `React.FunctionComponent`<`GridProps`\>

A simple Grid component. Can be configured to display a number of columns with different gutter spacing.

#### Defined in

[components/Grid/Grid.tsx:30](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Grid/Grid.tsx#L30)

___

### Input

• **Input**: `ForwardReferenceComponent`<``"input"``, `InputProps`\>

Displays the input with description, error message, visual invalid and valid states.

#### Defined in

[components/Form/Input/Input.tsx:31](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Form/Input/Input.tsx#L31)

___

### LoadingSpinner

• **LoadingSpinner**: `React.FunctionComponent`<`LoadingSpinnerProps`\>

A simple wrapper around the generic Sourcegraph React loading spinner

Supports additional custom styling relevant to this codebase.

#### Defined in

[components/LoadingSpinner/LoadingSpinner.tsx:18](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/LoadingSpinner/LoadingSpinner.tsx#L18)

___

### NavAction

• **NavAction**: `React.FunctionComponent`<`NavActionsProps`\>

#### Defined in

[components/NavBar/NavBar.tsx:91](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L91)

___

### NavActions

• **NavActions**: `React.FunctionComponent`<`NavActionsProps`\>

#### Defined in

[components/NavBar/NavBar.tsx:87](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L87)

___

### NavItem

• **NavItem**: `React.FunctionComponent`<`NavItemProps`\>

#### Defined in

[components/NavBar/NavBar.tsx:99](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L99)

___

### NavLink

• **NavLink**: `React.FunctionComponent`<`NavLinkProps`\>

#### Defined in

[components/NavBar/NavBar.tsx:113](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L113)

___

### PageHeader

• **PageHeader**: `React.FunctionComponent`<`Props`\>

#### Defined in

[components/PageHeader/PageHeader.tsx:40](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/PageHeader/PageHeader.tsx#L40)

___

### PageSelector

• **PageSelector**: `React.FunctionComponent`<`PageSelectorProps`\>

PageSelector should be used to render offset-pagination controls.
It is a controlled-component, the `currentPage` should be controlled by the consumer.

#### Defined in

[components/PageSelector/PageSelector.tsx:67](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/PageSelector/PageSelector.tsx#L67)

___

### RadioButton

• **RadioButton**: `React.FunctionComponent`<`RadioButtonProps`\>

Renders a single radio button.

Radio buttons should be used when a user must make a single choice from a list of two or more mutually exclusive options.

Grouped radio buttons should be visually presented together.

Useful article comparing radio buttons to checkboxes: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/

#### Defined in

[components/Form/RadioButton/RadioButton.tsx:22](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Form/RadioButton/RadioButton.tsx#L22)

___

### Select

• **Select**: `React.FunctionComponent`<`SelectProps`\>

A wrapper around the <select/> element.
Supports both native and custom styling.

Select should be used to provide a user with a list of options within a form.

Please note that this component takes <option/> elements as children. This is to easily support advanced functionality such as usage of <optgroup/>.

#### Defined in

[components/Form/Select/Select.tsx:45](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Form/Select/Select.tsx#L45)

___

### Tab

• **Tab**: `React.FunctionComponent`<`TabProps`\>

#### Defined in

[components/Tabs/Tabs.tsx:71](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Tabs/Tabs.tsx#L71)

___

### TabList

• **TabList**: `React.FunctionComponent`<`TabListProps`\>

#### Defined in

[components/Tabs/Tabs.tsx:61](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Tabs/Tabs.tsx#L61)

___

### TabPanel

• **TabPanel**: `React.FunctionComponent`

#### Defined in

[components/Tabs/Tabs.tsx:89](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Tabs/Tabs.tsx#L89)

___

### TabPanels

• **TabPanels**: `React.FunctionComponent`<`TabPanelsProps`\>

#### Defined in

[components/Tabs/Tabs.tsx:84](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Tabs/Tabs.tsx#L84)

___

### Tabs

• **Tabs**: `React.FunctionComponent`<`TabsProps`\>

reach UI tabs component with steroids, this tabs handles how the data should be loaded
in terms of a11y tabs are following all the WAI-ARIA Tabs Design Pattern.

See: https://reach.tech/tabs/

#### Defined in

[components/Tabs/Tabs.tsx:48](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Tabs/Tabs.tsx#L48)

___

### TextArea

• **TextArea**: `ForwardRefExoticComponent`<`TextAreaProps` & `RefAttributes`<`HTMLTextAreaElement`\>\>

Displays a textarea with description, error message, visual invalid and valid states.

#### Defined in

[components/Form/TextArea/TextArea.tsx:24](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/Form/TextArea/TextArea.tsx#L24)

## Functions

### NavBar

▸ `Const` **NavBar**(`__namedParameters`): `Element`

#### Parameters

| Name | Type |
| :------ | :------ |
| `__namedParameters` | `NavBarProps` |

#### Returns

`Element`

#### Defined in

[components/NavBar/NavBar.tsx:55](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L55)

___

### NavGroup

▸ `Const` **NavGroup**(`__namedParameters`): `Element`

#### Parameters

| Name | Type |
| :------ | :------ |
| `__namedParameters` | `NavGroupProps` |

#### Returns

`Element`

#### Defined in

[components/NavBar/NavBar.tsx:67](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/components/NavBar/NavBar.tsx#L67)

___

### useAutoFocus

▸ `Const` **useAutoFocus**(`__namedParameters`): `void`

Hook to ensure that an element is focused correctly.
Relying on the `autoFocus` attribute is not reliable within React.
https://reactjs.org/docs/accessibility.html#programmatically-managing-focus

#### Parameters

| Name | Type |
| :------ | :------ |
| `__namedParameters` | `UseAutoFocusParameters` |

#### Returns

`void`

#### Defined in

[hooks/useAutoFocus.ts:13](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/hooks/useAutoFocus.ts#L13)

___

### useControlledState

▸ **useControlledState**<`T`\>(`__namedParameters`): [`T`, (`item`: `T`) => `void`]

A hook to allow other components & hooks to easily support both controlled and uncontrolled variations of state.
`useControlledState` acts like `useState` except it assumes it can defer state management to the caller if an `onChange` parameter is passed.

#### Type parameters

| Name |
| :------ |
| `T` |

#### Parameters

| Name | Type |
| :------ | :------ |
| `__namedParameters` | `UseContolledParameters`<`T`\> |

#### Returns

[`T`, (`item`: `T`) => `void`]

#### Defined in

[hooks/useControlledState.ts:12](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/hooks/useControlledState.ts#L12)

___

### useDebounce

▸ `Const` **useDebounce**<`T`\>(`value`, `delay`): `T`

This function will trail debounce a changing value

#### Type parameters

| Name |
| :------ |
| `T` |

#### Parameters

| Name | Type | Description |
| :------ | :------ | :------ |
| `value` | `T` | The value expected to change |
| `delay` | `number` | Delay before updating the value |

#### Returns

`T`

The updated value

#### Defined in

[hooks/useDebounce.ts:10](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/hooks/useDebounce.ts#L10)

___

### useOffsetPagination

▸ **useOffsetPagination**(`__namedParameters`): `PaginationItem`[]

useOffsetPagination is a hook to easily manage offset-pagination logic.
This hook is capable of controlling its own state, however it is possible to override this
by listening to the component state with `onChange` and updating `page` manually

#### Parameters

| Name | Type |
| :------ | :------ |
| `__namedParameters` | `UseOffsetPaginationParameters` |

#### Returns

`PaginationItem`[]

#### Defined in

[hooks/useOffsetPagination.ts:107](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/hooks/useOffsetPagination.ts#L107)

___

### useSearchParameters

▸ `Const` **useSearchParameters**(): `URLSearchParams`

Return a new search parameters object based on the current URL.

#### Returns

`URLSearchParams`

#### Defined in

[hooks/useSearchParameters.ts:6](https://github.com/sourcegraph/sourcegraph/blob/86b8dee8a2/client/wildcard/src/hooks/useSearchParameters.ts#L6)
