import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'

import styles from './FormInput.module.scss'

interface FormInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of input. */
    title?: string
    /** Description block for field. */
    description?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Error massage for input. */
    error?: string
    errorInputState?: boolean
    /** Valid sign to show valid state on input. */
    valid?: boolean
    loading?: boolean
    /** Turn on or turn off autofocus for input. */
    autofocus?: boolean
    /** Custom class name for input element. */
    inputClassName?: string
    /** Input icon (symbol) which render right after the input element. */
    inputSymbol?: ReactNode
}

/** Displays input with description, error message, visual invalid and valid states. */
export const FormInput = forwardRef<HTMLInputElement, FormInputProps>((props, reference) => {
    const {
        type = 'text',
        title,
        description,
        className,
        inputClassName,
        inputSymbol,
        valid,
        error,
        loading = false,
        errorInputState,
        ...otherProps
    } = props

    return (
        <label className={classnames(className)}>
            {title && <div className="mb-2">{title}</div>}

            <LoaderInput className="d-flex" loading={loading}>
                <input
                    type={type}
                    className={classnames(styles.input, inputClassName, 'form-control', {
                        'is-valid': valid,
                        'is-invalid': !!error || errorInputState,
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {inputSymbol}
            </LoaderInput>

            {error && (
                <small className="text-danger form-text" role="alert">
                    {error}
                </small>
            )}
            {!error && description && (
                <small role="alert" className={classnames('text-muted', 'form-text')}>
                    {description}
                </small>
            )}
        </label>
    )
})
