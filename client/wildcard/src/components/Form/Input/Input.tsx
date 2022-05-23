/* eslint-disable react/display-name */
import { useRef, forwardRef, InputHTMLAttributes, ReactNode } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { useAutoFocus } from '../../../hooks/useAutoFocus'
import { ForwardReferenceComponent } from '../../../types'

import { InputLabel } from './InputLabel'
import { InputMessage } from './InputMessage'

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
    /** Error block shown below the input. */
    error?: ReactNode
    /** Disable input behavior */
    disabled?: boolean
    /** Determines the size of the input */
    variant?: 'regular' | 'small'
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
        ...otherProps
    } = props

    const localReference = useRef<HTMLInputElement>(null)
    const mergedReference = useMergeRefs([localReference, reference])

    useAutoFocus({ autoFocus, reference: localReference })

    const inputWithMessage = (
        <>
            <Component
                disabled={disabled}
                type={type}
                className={classNames(inputClassName, !label && className, 'form-control', 'with-invalid-icon', {
                    'is-valid': status === InputStatus.valid,
                    'is-invalid': error || status === InputStatus.error,
                    'form-control-sm': variant === 'small',
                })}
                {...otherProps}
                ref={mergedReference}
                autoFocus={autoFocus}
            />
            {(error || message) && <InputMessage isValid={!error}>{error || message}</InputMessage>}
        </>
    )

    if (label) {
        return (
            <InputLabel className={className} variant={variant} label={label}>
                {inputWithMessage}
            </InputLabel>
        )
    }

    return inputWithMessage
}) as ForwardReferenceComponent<'input', InputProps> & {
    Label: typeof InputLabel
    Message: typeof InputMessage
}

Input.displayName = 'Input'
Input.Label = InputLabel
Input.Message = InputMessage
