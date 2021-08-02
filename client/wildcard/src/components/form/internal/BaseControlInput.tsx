import classNames from 'classnames'
import React from 'react'

import { AccessibleFieldProps } from './AccessibleFieldType'
import { FormFieldLabel } from './FormFieldLabel'
import { FormFieldMessage } from './FormFieldMessage'
import { getValidStyle } from './utils'

export const BASE_CONTROL_TYPES = ['radio', 'checkbox'] as const

export type ControlInputProps = AccessibleFieldProps<React.InputHTMLAttributes<HTMLInputElement>> &
    React.RefAttributes<HTMLInputElement> & {
        /**
         * The <input> type. Use one of the currently supported types.
         */
        type?: typeof BASE_CONTROL_TYPES[number]
    }

export const BaseControlInput: React.FunctionComponent<ControlInputProps> = React.forwardRef(
    ({ children, className, message, isValid, type, ...props }, reference) => (
        <div className="form-check">
            <input
                ref={reference}
                type={type}
                className={classNames('form-check-input', getValidStyle(isValid), className)}
                {...props}
            />
            {'label' in props && (
                <FormFieldLabel htmlFor={props.id} className="form-check-label">
                    {props.label}
                </FormFieldLabel>
            )}
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
