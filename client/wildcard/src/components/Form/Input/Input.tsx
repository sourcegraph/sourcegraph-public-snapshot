import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'

import styles from './Input.module.scss'
import { ForwardReferenceComponent } from './types'

interface FormInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of input. */
    title?: string
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
    disable?: boolean
    /** Determines the size of the input */
    size?: 'regular' | 'small'
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 */
export const Input = forwardRef((props, reference) => {
    const {
        as: Component = 'input',
        type = 'text',
        size = 'regular',
        title,
        message,
        className,
        inputClassName,
        inputSymbol,
        disable,
        status,
        ...otherProps
    } = props

    return (
        <label className={classnames('w-100', className)}>
            {title && <div className="mb-2">{size === 'regular' ? title : <small>{title}</small>}</div>}

            <LoaderInput className="d-flex" loading={status === 'loading'}>
                <Component
                    disabled={disable}
                    type={type}
                    className={classnames(styles.input, inputClassName, 'form-control', 'with-invalid-icon', {
                        'is-valid': status === 'valid',
                        'is-invalid': status === 'error',
                        'form-control-sm': size === 'small',
                    })}
                    {...otherProps}
                    ref={reference}
                />

                {inputSymbol}
            </LoaderInput>

            {message && (
                <small
                    className={classnames(
                        status === 'error' ? 'text-danger' : 'text-muted',
                        'form-text font-weight-normal mt-2'
                    )}
                >
                    {message}
                </small>
            )}
        </label>
    )
}) as ForwardReferenceComponent<'input', FormInputProps>

Input.displayName = 'Input'
