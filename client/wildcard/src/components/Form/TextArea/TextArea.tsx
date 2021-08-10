import classnames from 'classnames'
import React, { forwardRef, ForwardRefExoticComponent, ReactNode, RefAttributes, TextareaHTMLAttributes } from 'react'

import styles from './TextArea.module.scss'

export interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
    /** Title of textarea. Used as label */
    label?: string
    /** Description block shown below the textarea. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Define an error in the textarea */
    isError?: boolean
    /** Disable textarea behavior */
    disabled?: boolean
    /** Determines the size of the textarea */
    size?: 'regular' | 'small'
}

/**
 * Displays a textarea with description, error message, visual invalid and valid states.
 */
export const TextArea: ForwardRefExoticComponent<TextAreaProps & RefAttributes<HTMLTextAreaElement>> = forwardRef(
    (props, reference) => {
        const { label, message, className, disabled, isError, size, ...otherProps } = props

        return (
            <label className={classnames('w-100', className)}>
                {label && <div className="mb-2">{size === 'regular' ? label : <small>{label}</small>}</div>}

                <textarea
                    disabled={disabled}
                    className={classnames(styles.textarea, 'form-control', {
                        'is-invalid': isError,
                        'form-control-sm': size === 'small',
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {message && (
                    <small className={classnames(isError ? 'text-danger' : 'text-muted', 'form-text')}>{message}</small>
                )}
            </label>
        )
    }
)

TextArea.displayName = 'TextArea'
