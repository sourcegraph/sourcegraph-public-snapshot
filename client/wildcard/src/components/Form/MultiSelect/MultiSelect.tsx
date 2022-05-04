import { ReactElement } from 'react'

import classNames from 'classnames'
import Select, { Props as SelectProps, StylesConfig, GroupBase } from 'react-select'

import { AccessibleFieldProps } from '../internal/AccessibleFieldType'
import { FormFieldLabel } from '../internal/FormFieldLabel'
import { FormFieldMessage } from '../internal/FormFieldMessage'
import { getValidStyle } from '../internal/utils'

import { ClearIndicator } from './ClearIndicator'
import { DropdownIndicator } from './DropdownIndicator'
import { MultiValueContainer } from './MultiValueContainer'
import { MultiValueLabel } from './MultiValueLabel'
import { MultiValueRemove } from './MultiValueRemove'
import { STYLES } from './styles'
import { THEME } from './theme'
import { MultiSelectOption } from './types'

import selectStyles from '../Select/Select.module.scss'
import styles from './MultiSelect.module.scss'

export type MultiSelectProps<Option = unknown> = AccessibleFieldProps<
    SelectProps<Option, true> & { options: SelectProps<Option, true>['options'] } & {
        // Require options
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
