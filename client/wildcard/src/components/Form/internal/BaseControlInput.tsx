import React from 'react'

import classNames from 'classnames'

import type { AccessibleFieldProps } from './AccessibleFieldType'
import { FormFieldLabel } from './FormFieldLabel'
import { FormFieldMessage } from './FormFieldMessage'
import { getValidStyle } from './utils'

import styles from './BaseControlInput.module.scss'

export const BASE_CONTROL_TYPES = ['radio', 'checkbox'] as const

export type ControlInputProps = AccessibleFieldProps<React.InputHTMLAttributes<HTMLInputElement>> &
    React.RefAttributes<HTMLInputElement> & {
        /**
         * The <input> type. Use one of the currently supported types.
         */
        type?: typeof BASE_CONTROL_TYPES[number]
        /**
         * CSS class name for the wrapper
         */
        wrapperClassName?: string
    }

export const BaseControlInput: React.FunctionComponent<React.PropsWithChildren<ControlInputProps>> = React.forwardRef(
    function BaseControlInput({ children, className, message, isValid, type, wrapperClassName, ...props }, reference) {
        return (
            <div className={classNames('form-check', wrapperClassName)}>
                {/* eslint-disable-next-line react/forbid-elements */}
                <input
                    ref={reference}
                    type={type}
                    className={classNames('form-check-input', getValidStyle(isValid), className)}
                    {...props}
                />
                {'label' in props && (
                    <FormFieldLabel htmlFor={props.id} className={classNames('form-check-label', styles.label)}>
                        {props.label}
                    </FormFieldLabel>
                )}
                {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
            </div>
        )
    }
)
