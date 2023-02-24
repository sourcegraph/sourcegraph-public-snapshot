import { FC, forwardRef, ReactNode, HTMLAttributes } from 'react'

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

export { Label }

export interface InputProps {
    /**
     * Text label of input.
     * @deprecated Use <Label /> composition components instead
     */
    label?: ReactNode

    /** Description block shown below the input. */
    message?: ReactNode

    /** To allow multiple files in file type inputs. */
    multiple?: boolean

    /** Custom class name for input element. */
    inputClassName?: string

    /** Input icon (symbol) which render right after the input element. */
    inputSymbol?: ReactNode

    /** Exclusive status */
    status?: InputStatus | `${InputStatus}`

    /** Optional error (validation) message. Rendered as Markdown. */
    error?: string

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
        multiple,
        message,
        className,
        inputClassName,
        inputSymbol,
        status = InputStatus.initial,
        error,
        autoFocus,
        ...attributes
    } = props

    const inputWithMessage = (
        <>
            <LoaderInput
                className={classNames('loader-input', styles.loaderInput, !label && className)}
                loading={status === InputStatus.loading}
            >
                <InputElement
                    as={Component}
                    {...attributes}
                    ref={reference}
                    variant={variant}
                    status={error ? InputStatus.error : status}
                    type={type}
                    multiple={multiple ?? false}
                    autoFocus={autoFocus}
                    className={inputClassName}
                />

                {inputSymbol}
            </LoaderInput>

            {error && <InputErrorMessage message={error} className="mt-2" />}
            {!error && message && <InputDescription message={message} className="mt-2" />}
        </>
    )

    if (label) {
        return (
            <Label className={classNames(styles.label, className)}>
                {label && <div className="mb-2">{variant === 'regular' ? label : <small>{label}</small>}</div>}
                {inputWithMessage}
            </Label>
        )
    }

    return inputWithMessage
}) as ForwardReferenceComponent<'input', InputProps>

interface InputElementProps {
    variant?: 'regular' | 'small'
    status?: InputStatus | `${InputStatus}`
}

export const InputElement = forwardRef(function InputElement(props, ref) {
    const {
        status = InputStatus.initial,
        variant = 'regular',
        as: Component = 'input',
        autoFocus,
        className,
        'aria-invalid': ariaInvalid,
        ...attributes
    } = props

    const mergedReference = useMergeRefs([ref])

    useAutoFocus({ autoFocus, reference: mergedReference })

    return (
        <Component
            {...attributes}
            ref={mergedReference}
            aria-invalid={ariaInvalid ?? status === InputStatus.error ? true : undefined}
            className={classNames(
                className,
                status === InputStatus.loading && styles.inputLoading,
                'form-control',
                'with-invalid-icon',
                {
                    'is-valid': status === InputStatus.valid,
                    'is-invalid': status === InputStatus.error,
                    'form-control-sm': variant === 'small',
                }
            )}
        />
    )
}) as ForwardReferenceComponent<'input', InputElementProps>

interface InputDescriptionProps extends HTMLAttributes<HTMLElement> {
    message?: ReactNode
}

export const InputDescription: FC<InputDescriptionProps> = props => {
    const { message, children, className, ...attributes } = props

    return (
        <small
            {...attributes}
            className={classNames('text-muted form-text font-weight-normal', styles.descriptionBlock, className)}
        >
            {message ?? children}
        </small>
    )
}

interface InputErrorMessageProps extends HTMLAttributes<HTMLElement> {
    message?: string
}

export const InputErrorMessage: FC<InputErrorMessageProps> = props => {
    const { message, className, ...attributes } = props

    return (
        <small
            {...attributes}
            role="alert"
            aria-live="polite"
            className={classNames('text-danger form-text font-weight-normal', className)}
        >
            <ErrorMessage error={message} />
        </small>
    )
}
