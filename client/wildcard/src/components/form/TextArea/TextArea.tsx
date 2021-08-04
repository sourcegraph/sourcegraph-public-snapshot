import classnames from 'classnames'
import React, { forwardRef, ForwardRefExoticComponent, InputHTMLAttributes, ReactNode, RefAttributes } from 'react'

import styles from './TextArea.module.scss'

export interface FormTextAreaProps extends InputHTMLAttributes<HTMLTextAreaElement> {
    /** Title of textarea. Used as label */
    title?: string
    /** Description block shown below the textarea. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Define an error in the textarea */
    isError?: boolean
    /** Disable textarea behavior */
    disabled?: boolean
}

/**
 * Displays a textarea with description, error message, visual invalid and valid states.
 */
export const TextArea: ForwardRefExoticComponent<FormTextAreaProps & RefAttributes<HTMLTextAreaElement>> = forwardRef(
    (props, reference) => {
        const { title, message, className, disabled, isError, ...otherProps } = props

        return (
            <label className={classnames('w-100', className)}>
                {title && <div className="mb-2">{title}</div>}

                <textarea
                    disabled={disabled}
                    className={classnames(styles.textarea, 'form-control', {
                        'is-invalid': isError,
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
