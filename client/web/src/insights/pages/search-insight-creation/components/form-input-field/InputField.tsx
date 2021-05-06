import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react'

interface InputFieldProps extends InputHTMLAttributes<HTMLInputElement> {
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
}

/** Displays input with description, error message, visual invalid and valid states. */
export const InputField = forwardRef<HTMLInputElement, InputFieldProps>((props, reference) => {
    const { type = 'text', title, description, className, valid, error, errorInputState, ...otherProps } = props

    return (
        <label className={classnames(className)}>
            {title && <div className="mb-2">{title}</div>}

            <input
                type={type}
                className={classnames('form-control', {
                    'is-valid': valid,
                    'is-invalid': !!error || errorInputState,
                })}
                {...otherProps}
                ref={reference}
            />

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
