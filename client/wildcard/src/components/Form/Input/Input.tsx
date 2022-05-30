import { useRef, forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'

import { Typography } from '../..'
import { useAutoFocus } from '../../../hooks/useAutoFocus'
import { ForwardReferenceComponent } from '../../../types'

import styles from './Input.module.scss'

export enum InputStatus {
    initial = 'initial',
    error = 'error',
    loading = 'loading',
    valid = 'valid',
}

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** text label of input. */
    label?: ReactNode
    /** Description block shown below the input. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /** Custom class name for input element. */
    inputClassName?: string
    /** Input icon (symbol) which render right after the input element. */
    inputSymbol?: ReactNode
    /** Exclusive status */
    status?: InputStatus | `${InputStatus}`
    error?: ReactNode
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
        status = InputStatus.initial,
        error,
        autoFocus,
        ...otherProps
    } = props

    const localReference = useRef<HTMLInputElement>(null)
    const mergedReference = useMergeRefs([localReference, reference])

    useAutoFocus({ autoFocus, reference: localReference })

    const messageClassName = 'form-text font-weight-normal mt-2'
    const inputWithMessage = (
        <>
            <LoaderInput
                className={classNames('d-flex loader-input', !label && className)}
                loading={status === InputStatus.loading}
            >
                <Component
                    disabled={disabled}
                    type={type}
                    className={classNames(
                        inputClassName,
                        status === InputStatus.loading && styles.inputLoading,
                        'form-control',
                        'with-invalid-icon',
                        {
                            'is-valid': status === InputStatus.valid,
                            'is-invalid': error || status === InputStatus.error,
                            'form-control-sm': variant === 'small',
                        }
                    )}
                    {...otherProps}
                    ref={mergedReference}
                    autoFocus={autoFocus}
                />

                {inputSymbol}
            </LoaderInput>

            {error && (
                <small role="alert" className={classNames('text-danger', messageClassName)}>
                    {error}
                </small>
            )}
            {!error && message && <small className={classNames('text-muted', messageClassName)}>{message}</small>}
        </>
    )

    if (label) {
        return (
            <Typography.Label className={classNames('w-100', className)}>
                {label && <div className="mb-2">{variant === 'regular' ? label : <small>{label}</small>}</div>}
                {inputWithMessage}
            </Typography.Label>
        )
    }

    return inputWithMessage
}) as ForwardReferenceComponent<'input', InputProps>

Input.displayName = 'Input'
