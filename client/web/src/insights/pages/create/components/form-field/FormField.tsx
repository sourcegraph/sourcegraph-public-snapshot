import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes } from 'react'

import styles from './FormField.module.scss'

interface InputFieldProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of input. */
    title?: string
    /** Description block for field. */
    description?: string
    /** Custom class name for root label element. */
    className?: string
    /** Error massage for input. */
    error?: string
    /** Valid sign to show valid state on input. */
    valid?: boolean
}

/** Displays input with description, error message, visual invalid and valid states. */
export const InputField = forwardRef<HTMLInputElement, InputFieldProps>((props, reference) => {
    const { title, description, className, valid, error, ...otherProps } = props

    return (
        <label className={classnames(styles.formField, className)}>
            {title && <h4>{title}</h4>}

            <input
                type="text"
                className={classnames(styles.formFieldInput, 'form-control', {
                    'is-valid': valid,
                    'is-invalid': !!error,
                })}
                {...otherProps}
                ref={reference}
            />

            {error && <span className={styles.formFieldError}>*{error}</span>}
            {!error && description && (
                <span className={classnames(styles.formFieldDescription, 'text-muted')}>{description}</span>
            )}
        </label>
    )
})
