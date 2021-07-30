import classNames from 'classnames'
import React from 'react'

import { FormFieldMessage } from '../internal/FormFieldMessage'

export interface SelectProps
    extends React.SelectHTMLAttributes<HTMLSelectElement>,
        React.RefAttributes<HTMLSelectElement> {
    className?: string
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

export const getSelectStyles = (isCustomStyle?: boolean): string => {
    if (isCustomStyle) {
        return 'custom-select'
    }

    return 'form-control'
}

export const Select: React.FunctionComponent<SelectProps> = React.forwardRef(
    ({ children, className, label, message, isValid, isCustomStyle, ...selectProps }, reference) => (
        <div className="form-group">
            <label>
                <select
                    ref={reference}
                    className={classNames(getSelectStyles(isCustomStyle), className)}
                    {...selectProps}
                >
                    {children}
                </select>
                {label}
            </label>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
