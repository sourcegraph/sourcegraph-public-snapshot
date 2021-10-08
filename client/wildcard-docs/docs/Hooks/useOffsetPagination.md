---
id: "useOffsetPagination"
title: "Function: useOffsetPagination"
sidebar_label: "useOffsetPagination"
sidebar_position: 0
custom_edit_url: null
---

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

[hooks/useOffsetPagination.ts:107](https://github.com/sourcegraph/sourcegraph/blob/8be9dcbff0/client/wildcard/src/hooks/useOffsetPagination.ts#L107)
