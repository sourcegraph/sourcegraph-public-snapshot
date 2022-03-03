import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { ReactElement } from 'react'
import Select, {
    components,
    Props as SelectProps,
    StylesConfig,
    ClearIndicatorProps,
    DropdownIndicatorProps,
    MultiValueGenericProps,
    MultiValueRemoveProps,
    GroupBase,
} from 'react-select'

import { Badge } from '../..'
import { AccessibleFieldProps } from '../internal/AccessibleFieldType'
import { FormFieldLabel } from '../internal/FormFieldLabel'
import { FormFieldMessage } from '../internal/FormFieldMessage'
import { getValidStyle } from '../internal/utils'
import selectStyles from '../Select/Select.module.scss'

import styles from './MultiSelect.module.scss'
import { STYLES } from './styles'
import { THEME } from './theme'

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

export type MultiSelectProps<Option = unknown> = AccessibleFieldProps<
    SelectProps<Option, true> & {
        /**
         * Optional label position. Default is 'inline'
         */
        labelVariant?: 'inline' | 'block'
    }
>

/**
 * A wrapper around `react-select`'s `Select` component for producing multiselect dropdown
 * components.
 *
 * `MultiSelect` should be used to provide a user with a list of options from which they
 * can select none to many, within a form.
 *
 * @param options An array of the `Option`s to be listed from the dropdown.
 */
export const MultiSelect = <OptionValue extends unknown = unknown>({
    className,
    labelVariant,
    message,
    options,
    ...props
}: MultiSelectProps<MultiSelectOption<OptionValue>>): ReactElement => (
    <div className={classNames('form-group', className)}>
        {'label' in props && (
            <FormFieldLabel
                htmlFor={props.id}
                className={labelVariant === 'block' ? selectStyles.labelBlock : undefined}
            >
                {props.label}
            </FormFieldLabel>
        )}
        <Select<MultiSelectOption<OptionValue>, true, GroupBase<MultiSelectOption<OptionValue>>>
            isMulti={true}
            className={classNames(styles.multiSelect, getValidStyle(props.isValid))}
            options={options}
            theme={THEME}
            styles={STYLES as StylesConfig<MultiSelectOption<OptionValue>>}
            hideSelectedOptions={false}
            components={{
                ClearIndicator,
                DropdownIndicator,
                MultiValueContainer,
                MultiValueLabel,
                MultiValueRemove,
            }}
            {...props}
        />
        {message && <FormFieldMessage isValid={props.isValid}>{message}</FormFieldMessage>}
    </div>
)

// Overwrite the clear indicator with `CloseIcon`
const ClearIndicator = <OptionValue extends unknown = unknown>(
    props: ClearIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.ClearIndicator {...props}>
        <CloseIcon className={styles.clearIcon} />
    </components.ClearIndicator>
)

// Overwrite the dropdown indicator with `ChevronDownIcon`
const DropdownIndicator = <OptionValue extends unknown = unknown>(
    props: DropdownIndicatorProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.DropdownIndicator {...props}>
        <ChevronDownIcon className={styles.dropdownIcon} />
    </components.DropdownIndicator>
)

// Overwrite the multi value container with Wildcard `Badge`
const MultiValueContainer = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => (
    <Badge variant="secondary" className={styles.multiValueContainer} {...props} />
)

// Remove extra wrappers around multi value label
const MultiValueLabel = <OptionValue extends unknown = unknown>({
    innerProps: _innerProps,
    selectProps: _selectProps,
    ...props
}: MultiValueGenericProps<MultiSelectOption<OptionValue>, true>): ReactElement => <span {...props} />

// Overwrite the multi value remove indicator with `CloseIcon`
const MultiValueRemove = <OptionValue extends unknown = unknown>(
    props: MultiValueRemoveProps<MultiSelectOption<OptionValue>, true>
): ReactElement => (
    <components.MultiValueRemove {...props}>
        <CloseIcon className={styles.removeIcon} />
    </components.MultiValueRemove>
)
