import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react'

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
        errorInputState,
        ...otherProps
    } = props

    return (
        <label className={classnames(className)}>
            {title && <div className="mb-2">{title}</div>}

            <div className="d-flex">
                <input
                    type={type}
                    className={classnames(inputClassName, 'form-control', {
                        'is-valid': valid,
                        'is-invalid': !!error || errorInputState,
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {inputSymbol}
            </div>

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
