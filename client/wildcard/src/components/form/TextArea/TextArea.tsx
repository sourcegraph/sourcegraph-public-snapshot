import classnames from 'classnames'
import React, { forwardRef, ForwardRefExoticComponent, InputHTMLAttributes, ReactNode, RefAttributes } from 'react'

import styles from './TextArea.module.scss'

export interface FormTextAreaProps extends InputHTMLAttributes<HTMLTextAreaElement> {
    /** Title of input. */
    title?: string
    /** Description block shown below the input. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Turn on or turn off autofocus for input. */
    autofocus?: boolean
    /** Custom class for text area element. */
    textAreaClassName?: string
    /** Exclusive status */
    status?: 'error'
    /** Disable input behavior */
    disable?: boolean
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 */
export const TextArea: ForwardRefExoticComponent<FormTextAreaProps & RefAttributes<unknown>> = forwardRef(
    (props, reference) => {
        const { title, message, className, textAreaClassName, disable, status, ...otherProps } = props

        return (
            <label className={classnames('w-100', className)}>
                {title && <div className="mb-2">{title}</div>}

                <TextArea
                    disabled={disable}
                    className={classnames(/* styles.textarea,*/ 'form-control', {
                        'is-invalid': status === 'error',
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {message && (
                    <small
                        className={classnames(
                            status === 'error' ? 'text-danger' : 'text-muted',
                            'form-text'
                            // styles.message
                        )}
                    >
                        {message}
                    </small>
                )}
            </label>
        )
    }
)
