import classNames from 'classnames'
import React from 'react'

import { FormFieldLabel } from '../internal/FormFieldLabel'
import { FormFieldMessage } from '../internal/FormFieldMessage'
import { getValidStyle } from '../internal/utils'

export interface SelectProps
    extends React.SelectHTMLAttributes<HTMLSelectElement>,
        React.RefAttributes<HTMLSelectElement> {
    className?: string
    /**
     * A unique ID for the <input> element. This is required to correctly associate the rendered <label> with the <input>
     */
    id: string
    /**
     * Used to control the styling of the <select> and surrounding elements.
     * Set this value to `false` to show invalid styling.
     * Set this value to `true` to show valid styling.
     */
    isValid?: boolean
    /**
     * Descriptive text rendered within a <label> element.
     */
    label: React.ReactNode
    /**
     * Optional message to display below the <select>.
     * This should typically be used to display additional information (perhap error/success states) to the user.
     * It will be styled differently if `isValid` is truthy or falsy.
     */
    message?: React.ReactNode
    /**
     * Use the Bootstrap custom <select> styles
     */
    isCustomStyle?: boolean
}

/**
 * Returns the Boostrap specific style to differentiate between native and custom <select> styles.
 */
export const getSelectStyles = (isCustomStyle?: boolean): string => {
    if (isCustomStyle) {
        return 'custom-select'
    }

    return 'form-control'
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
    ({ children, id, className, label, message, isValid, isCustomStyle, ...selectProps }, reference) => (
        <div className="form-group">
            <FormFieldLabel htmlFor={id}>{label}</FormFieldLabel>
            <select
                id={id}
                ref={reference}
                className={classNames(getSelectStyles(isCustomStyle), getValidStyle(isValid), className)}
                {...selectProps}
            >
                {children}
            </select>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
