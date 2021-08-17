import classNames from 'classnames'
import React from 'react'

import { AccessibleFieldProps } from '../internal/AccessibleFieldType'
import { FormFieldLabel } from '../internal/FormFieldLabel'
import { FormFieldMessage } from '../internal/FormFieldMessage'
import { getValidStyle } from '../internal/utils'

export const SELECT_SIZES = ['sm', 'lg'] as const

export type SelectProps = AccessibleFieldProps<React.SelectHTMLAttributes<HTMLSelectElement>> &
    React.RefAttributes<HTMLSelectElement> & {
        /**
         * Use the Bootstrap custom <select> styles
         */
        isCustomStyle?: boolean
        /**
         * Optional size modifier to render a smaller or larger <select> variant
         */
        selectSize?: typeof SELECT_SIZES[number]
    }

/**
 * Returns the Boostrap specific style to differentiate between native and custom <select> styles.
 */
export const getSelectStyles = ({
    isCustomStyle,
    selectSize,
}: Pick<SelectProps, 'isCustomStyle' | 'selectSize'>): string => {
    if (isCustomStyle) {
        return classNames('custom-select', selectSize && `custom-select-${selectSize}`)
    }

    return classNames('form-control', selectSize && `form-control-${selectSize}`)
}

/**
 * A wrapper around the <select> element.
 * Supports both native and custom styling.
 *
 * Select should be used to provide a user with a list of options within a form.
 *
 * Please note that this component takes <option> elements as children. This is to easily support advanced functionality such as usage of <optgroup>.
 */
export const Select: React.FunctionComponent<SelectProps> = React.forwardRef(
    ({ children, className, message, isValid, isCustomStyle, selectSize, ...props }, reference) => (
        <div className="form-group">
            {'label' in props && <FormFieldLabel htmlFor={props.id}>{props.label}</FormFieldLabel>}
            <select
                ref={reference}
                className={classNames(
                    getSelectStyles({ isCustomStyle, selectSize }),
                    getValidStyle(isValid),
                    className
                )}
                {...props}
            >
                {children}
            </select>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
