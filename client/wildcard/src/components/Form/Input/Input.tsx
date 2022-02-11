import classNames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'

import { ForwardReferenceComponent } from '../../../types'

import styles from './Input.module.scss'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** text label of input. */
    label?: string
    /** Description block shown below the input. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Custom class name for input element. */
    inputClassName?: string
    /** Input icon (symbol) which render right after the input element. */
    inputSymbol?: ReactNode
    /** Exclusive status */
    status?: 'error' | 'loading' | 'valid'
    /** Disable input behavior */
    disabled?: boolean
    /** Determines the size of the input */
    variant?: 'regular' | 'small'
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 */
export const Input = forwardRef((props, reference) => {
    const {
        as: Component = 'input',
        type = 'text',
        variant = 'regular',
        label,
        message,
        className,
        inputClassName,
        inputSymbol,
        disabled,
        status,
        ...otherProps
    } = props

    return (
        <label className={classNames('w-100', className)}>
            {label && <div className="mb-2">{variant === 'regular' ? label : <small>{label}</small>}</div>}

            <LoaderInput className="d-flex" loading={status === 'loading'}>
                <Component
                    disabled={disabled}
                    type={type}
                    className={classNames(styles.input, inputClassName, 'form-control', 'with-invalid-icon', {
                        'is-valid': status === 'valid',
                        'is-invalid': status === 'error',
                        'form-control-sm': variant === 'small',
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {inputSymbol}
            </LoaderInput>

            {message && (
                <small
                    className={classNames(
                        status === 'error' ? 'text-danger' : 'text-muted',
                        'form-text font-weight-normal mt-2'
                    )}
                >
                    {message}
                </small>
            )}
        </label>
    )
}) as ForwardReferenceComponent<'input', InputProps>

Input.displayName = 'Input'
