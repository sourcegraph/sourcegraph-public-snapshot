import classNames from 'classnames'
import React from 'react'

import { FormFieldLabel } from './FormFieldLabel'
import { FormFieldMessage } from './FormFieldMessage'
import { getValidStyle } from './utils'

export const BASE_CONTROL_TYPES = ['radio', 'checkbox'] as const

export interface BaseControlInputProps
    extends React.InputHTMLAttributes<HTMLInputElement>,
        React.RefAttributes<HTMLInputElement> {
    className?: string
    /**
     * A unique ID for the <input> element. This is required to correctly associate the rendered <label> with the <input>
     */
    id: string
    /**
     * Used to control the styling of the <input> and surrounding elements.
     * Set this value to `false` to show invalid styling.
     * Set this value to `true` to show valid styling.
     */
    isValid?: boolean
    /**
     * Descriptive text rendered within a <label> element.
     */
    label: React.ReactNode
    /**
     * Optional message to display below the <input>.
     * This should typically be used to display additional information to the user.
     * It will be styled differently if `isValid` is truthy or falsy.
     */
    message?: React.ReactNode
    /**
     * The <input> type. Use one of the currently supported types.
     */
    type?: typeof BASE_CONTROL_TYPES[number]
}

export const BaseControlInput: React.FunctionComponent<BaseControlInputProps> = React.forwardRef(
    ({ children, id, className, label, message, isValid, type, ...inputProps }, reference) => (
        <div className="form-check">
            <input
                id={id}
                ref={reference}
                type={type}
                className={classNames('form-check-input', getValidStyle(isValid), className)}
                {...inputProps}
            />
            <FormFieldLabel htmlFor={id} className="form-check-label">
                {label}
            </FormFieldLabel>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
