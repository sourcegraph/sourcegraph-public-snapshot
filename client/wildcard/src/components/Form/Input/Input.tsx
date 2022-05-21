import { useRef, forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { Typography } from '../..'
import { useAutoFocus } from '../../../hooks/useAutoFocus'
import { ForwardReferenceComponent } from '../../../types'

export enum InputStatus {
    initial = 'initial',
    error = 'error',
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
    /** Exclusive status */
    status?: InputStatus | `${InputStatus}`
    error?: ReactNode
    /** Disable input behavior */
    disabled?: boolean
    /** Determines the size of the input */
    variant?: 'regular' | 'small'
    /** Supports appending an element to input */
    inputSymbol?: ReactNode
    /** Custom class name for input element and symbol wrapper. */
    inputSymbolWrapperClassName?: string
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 * Does not support Loader icon and status=loadind (user FormInput to get support for loading state)
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
        disabled,
        status = InputStatus.initial,
        error,
        autoFocus,
        inputSymbol,
        inputSymbolWrapperClassName,
        ...otherProps
    } = props

    const localReference = useRef<HTMLInputElement>(null)
    const mergedReference = useMergeRefs([localReference, reference])

    useAutoFocus({ autoFocus, reference: localReference })
    const inputElement = (
        <Component
            disabled={disabled}
            type={type}
            className={classNames(inputClassName, 'form-control', 'with-invalid-icon', {
                'is-valid': status === InputStatus.valid,
                'is-invalid': error || status === InputStatus.error,
                'form-control-sm': variant === 'small',
            })}
            {...otherProps}
            ref={mergedReference}
            autoFocus={autoFocus}
        />
    )

    const inputWithSymbol = (
        <div className={classNames('d-flex', inputSymbolWrapperClassName)}>
            {inputElement}
            {inputSymbol}
        </div>
    )

    const messageClassName = 'form-text font-weight-normal mt-2'
    const inputWithMessage = (
        <>
            {inputSymbol ? inputWithSymbol : inputElement}
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
