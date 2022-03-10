import { GroupBase } from 'react-select'

import { AccessibleFieldProps } from '../internal/AccessibleFieldType'

/**
 * Generic type for an option to be listed from the `MultiSelect` dropdown.
 *
 * @param OptionValue The type of the value of the option, i.e. a union set of all
 * possible values.
 */
export interface MultiSelectOption<OptionValue = unknown> {
    value: OptionValue
    label: string
}

/**
 * Generic type for the state a consumer of `MultiSelect` should expect to manage.
 *
 * @param OptionValue The type of the value of the option, i.e. a union set of all
 * possible values.
 */
export type MultiSelectState<OptionValue = unknown> = readonly MultiSelectOption<OptionValue>[]

// We use module augmentation to make TS aware of custom props available from `Select`
// custom components, styles, theme, etc.
// See: https://react-select.com/typescript#custom-select-props
declare module 'react-select/dist/declarations/src/Select' {
    export interface Props<Option, IsMulti extends boolean, Group extends GroupBase<Option>> {
        isValid?: AccessibleFieldProps<{}>['isValid']
    }
}
