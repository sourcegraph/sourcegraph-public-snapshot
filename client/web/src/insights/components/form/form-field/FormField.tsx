import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode, useEffect, useRef } from 'react'
import { useMergeRefs } from 'use-callback-ref'

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
    /** Turn on or turn off autofocus for input. */
    autofocus?: boolean
    /** Custom class name for input element. */
    inputClassName?: string;
    /** Input icon (symbol) which render right after the input element. */
    inputSymbol?: ReactNode;
}

/** Displays input with description, error message, visual invalid and valid states. */
export const InputField = forwardRef<HTMLInputElement, InputFieldProps>((props, reference) => {
    const { title, description, className, inputClassName, inputSymbol, valid, error, autofocus = false, ...otherProps } = props
    const localInputReference = useRef<HTMLInputElement>(null)

    useEffect(() => {
        if (autofocus) {
            localInputReference.current?.focus()
        }
    }, [autofocus])

    return (
        <label className={classnames(styles.formField, className)}>
            {title && <h4>{title}</h4>}

            <div className={styles.formFieldInputBlock}>
                <input
                    type="text"
                    className={classnames(inputClassName, 'form-control', {
                        'is-valid': valid,
                        'is-invalid': !!error,
                    })}
                    {...otherProps}
                    ref={useMergeRefs([localInputReference, reference])}
                />

                { inputSymbol }
            </div>

            {error && <span className={styles.formFieldError}>*{error}</span>}
            {!error && description && (
                <span className={classnames(styles.formFieldDescription, 'text-muted')}>{description}</span>
            )}
        </label>
    )
})
