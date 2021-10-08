---
id: "useControlledState"
title: "Function: useControlledState"
sidebar_label: "useControlledState"
sidebar_position: 0
custom_edit_url: null
---

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

[hooks/useControlledState.ts:12](https://github.com/sourcegraph/sourcegraph/blob/8be9dcbff0/client/wildcard/src/hooks/useControlledState.ts#L12)
