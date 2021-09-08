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
         * The id used to match the label to the input.
         */
        id: string
        /**
         * The label for this input.
         */
        label: React.ReactNode
        /**
         * The <input> type. Use one of the currently supported types.
         */
        type?: typeof BASE_CONTROL_TYPES[number]
        /**
         * Used to pass props through to the input element.
         */
        inputProps?: React.InputHTMLAttributes<HTMLInputElement>
        /**
         * Used to pass props through to the label element.
         */
        labelProps?: React.LabelHTMLAttributes<HTMLLabelElement>
    }

export const BaseControlInput: React.FunctionComponent<ControlInputProps> = React.forwardRef(
    ({ id, label, className, message, isValid, type, labelProps, inputProps, ...props }, reference) => (
        <div {...props} className={classNames('form-check', className)}>
            <input
                id={id}
                ref={reference}
                type={type}
                className={classNames('form-check-input', getValidStyle(isValid), inputProps?.className)}
                {...inputProps}
            />
            <FormFieldLabel
                {...labelProps}
                htmlFor={id}
                className={classNames('form-check-label', labelProps?.className)}
            >
                {label}
            </FormFieldLabel>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
)
