import { useRef, forwardRef, ReactNode } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { Label } from '../..'
import { useAutoFocus } from '../../../hooks'
import { ForwardReferenceComponent } from '../../../types'
import { ErrorMessage } from '../../ErrorMessage'
import { LoaderInput } from '../LoaderInput'

import styles from './Input.module.scss'

export enum InputStatus {
    initial = 'initial',
    error = 'error',
    loading = 'loading',
    valid = 'valid',
}

export interface InputProps {
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
    /** Optional error (validation) message. Rendered as Markdown. */
    error?: string
    /** Disable input behavior */
    disabled?: boolean
    /** Determines the size of the input */
    variant?: 'regular' | 'small'
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 */
export const Input = forwardRef(function Input(props, reference) {
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
                className={classNames('loader-input', styles.loaderInput, !label && className)}
                loading={status === InputStatus.loading}
            >
                <Component
                    {...otherProps}
                    type={type}
                    disabled={disabled}
                    ref={mergedReference}
                    autoFocus={autoFocus}
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
                />

                {inputSymbol}
            </LoaderInput>

            {error && (
                <small role="alert" aria-live="polite" className={classNames('text-danger', messageClassName)}>
                    <ErrorMessage error={error} />
                </small>
            )}

            {!error && message && <small className={classNames('text-muted', messageClassName)}>{message}</small>}
        </>
    )

    if (label) {
        return (
            <Label className={classNames('w-100', className)}>
                {label && <div className="mb-2">{variant === 'regular' ? label : <small>{label}</small>}</div>}
                {inputWithMessage}
            </Label>
        )
    }

    return inputWithMessage
}) as ForwardReferenceComponent<'input', InputProps>

Input.displayName = 'Input'
