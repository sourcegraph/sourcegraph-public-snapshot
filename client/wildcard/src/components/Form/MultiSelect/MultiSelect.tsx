import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { ReactElement } from 'react'
import Select, {
    components,
    Props as SelectProps,
    StylesConfig,
    ThemeConfig,
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

const THEME: ThemeConfig = theme => ({
    ...theme,
    borderRadius: 3,
    // Each identifiable instance of `Select` components using a theme color has been
    // overwritten by `STYLES` in order to switch to light mode/dark mode. These colors
    // defined here only serve as a fallback in case of unexpected or missed color usage.
    colors: {
        primary: 'var(--primary)',
        // Never used.
        primary75: 'var(--primary)',
        // Used for `option` background color, which is overwritten in `STYLES`.
        primary50: 'var(--primary-2)',
        // Used for `option` background color, which is overwritten in `STYLES`.
        primary25: 'var(--primary-2)',
        // Used for `multiValueRemove` color, which is replaced with a custom component.
        danger: 'var(--danger)',
        // Used for `multiValueRemove` background color, which is overwritten in `STYLES`.
        dangerLight: 'var(--danger)',
        // Used for `menu` and `control` background color as well as `option` color, all
        // of which are overwritten in `STYLES`.
        neutral0: 'var(--react-select-neutral0)',
        // Used for `control` background color and `placeholder` color, both of which are
        // overwritten in `STYLES`.
        neutral5: 'var(--react-select-neutral5)',
        // Used for `indicatorSeparator` background color and `control` border color, both
        // of which are overwritten in `STYLES`, and `multiValue` background color, which
        // is replaced with a custom component.
        neutral10: 'var(--react-select-neutral10)',
        // Used for `indicatorSeparator` background color, `control` border color, and
        // `option` color, all of which are overwritten in `STYLES`.
        neutral20: 'var(--react-select-neutral20)',
        // Used for `control` border color, which is overwritten in `STYLES`.
        neutral30: 'var(--react-select-neutral30)',
        // Used by components that aren't used for `isMulti=true` or have been replaced
        // with custom ones.
        neutral40: 'var(--react-select-neutral40)',
        // Used for `placeholder` color, which is overwritten in `STYLES`.
        neutral50: 'var(--react-select-neutral50)',
        // Used by components that have been replaced with custom ones.
        neutral60: 'var(--react-select-neutral60)',
        // Never used.
        neutral70: 'var(--react-select-neutral70)',
        // Used for `input` and `multiValue` color, which are both overwritten in
        // `STYLES`, as well as components that aren't used for `isMulti=true` or have
        // been replaced with custom ones.
        neutral80: 'var(--react-select-neutral80)',
        // Never used.
        neutral90: 'var(--react-select-neutral90)',
    },
})

const STYLES: StylesConfig = {
    clearIndicator: provided => ({
        ...provided,
        padding: '0 0.125rem',
    }),
    control: (provided, state) => ({
        ...provided,
        // Styles here replicate the styles of `wildcard/Select`
        backgroundColor: state.isDisabled ? 'var(--input-disabled-bg)' : 'var(--input-bg)',
        borderColor: state.selectProps.isValid
            ? 'var(--success)'
            : state.selectProps.isValid === false
            ? 'var(--danger)'
            : state.isFocused
            ? state.theme.colors.primary
            : 'var(--input-border-color)',
        boxShadow: state.isFocused
            ? // These are stolen from `wildcard/Input` and `wildcard/Select`, which seem to come from Bootstrap
              state.selectProps.isValid
                ? 'var(--input-focus-box-shadow-valid)'
                : state.selectProps.isValid === false
                ? 'var(--input-focus-box-shadow-invalid)'
                : 'var(--input-focus-box-shadow)'
            : undefined,
        '&:hover': {
            borderColor: undefined,
        },
    }),
    dropdownIndicator: provided => ({
        ...provided,
        padding: '0 0.125rem',
    }),
    indicatorSeparator: (provided, state) => ({
        ...provided,
        backgroundColor: state.hasValue ? 'var(--input-border-color)' : 'transparent',
    }),
    input: provided => ({
        ...provided,
        color: 'var(--input-color)',
        margin: '0 0.125rem',
        padding: 0,
    }),
    menu: provided => ({
        ...provided,
        background: 'var(--dropdown-bg)',
        padding: '0.25rem 0',
        margin: '0.125rem 0 0',
        dropShadow: 'var(--dropdown-shadow)',
    }),
    menuList: provided => ({
        ...provided,
        padding: 0,
    }),
    multiValueRemove: (provided, state) => ({
        ...provided,
        backgroundColor: 'transparent',
        boxShadow: state.isFocused ? 'var(--input-focus-box-shadow)' : undefined,
        ':hover': {
            ...provided[':hover'],
            backgroundColor: 'transparent',
            color: undefined,
        },
    }),
    noOptionsMessage: provided => ({
        ...provided,
        color: 'var(--input-placeholder-color)',
    }),
    option: (provided, state) => ({
        ...provided,
        backgroundColor: state.isSelected
            ? state.theme.colors.primary
            : state.isFocused
            ? 'var(--dropdown-link-hover-bg)'
            : 'transparent',
        color: state.isSelected ? 'var(--light-text)' : undefined,
        ':hover': {
            cursor: 'pointer',
        },
        ':active': {
            backgroundColor: !state.isDisabled
                ? state.isSelected
                    ? state.theme.colors.primary
                    : 'var(--dropdown-link-hover-bg)'
                : undefined,
        },
    }),
    placeholder: (provided, state) => ({
        ...provided,
        color: state.isDisabled ? 'var(--gray-06)' : 'var(--input-placeholder-color)',
    }),
    valueContainer: provided => ({
        ...provided,
        padding: '0.125rem 0.125rem 0.125rem 0.75rem',
    }),
}

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
