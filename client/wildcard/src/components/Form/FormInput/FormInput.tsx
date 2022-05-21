import { forwardRef } from 'react'

import classNames from 'classnames'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'

import { ForwardReferenceComponent } from '../../../types'
import { Label } from '../../Typography/Label'
import { Input, InputProps, InputStatus } from '../Input'

import styles from './FormInput.module.scss'

enum LoadingStatus {
    loading = 'loading',
}

export type FormInputStatusType = LoadingStatus | InputStatus
export const FormInputStatus = { ...LoadingStatus, ...InputStatus }

export interface FormInputProps extends Omit<InputProps, 'status'> {
    status?: FormInputStatusType | `${FormInputStatusType}`
}

/**
 * Displays the input with description, error message, visual invalid and valid states.
 * Renders Input component within LoaderInput to display loader icon with status=loading.
 */
export const FormInput = forwardRef((props, reference) => {
    const {
        variant = 'regular',
        label,
        className,
        status = FormInputStatus.initial,
        inputClassName,
        ...otherProps
    } = props

    const inputWithMessage = (
        <>
            <LoaderInput className={classNames(!label && className)} loading={status === FormInputStatus.loading}>
                <Input
                    ref={reference}
                    className={className}
                    status={status === 'loading' ? 'initial' : status}
                    inputClassName={classNames(styles.input, inputClassName)}
                    {...otherProps}
                />
            </LoaderInput>
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
}) as ForwardReferenceComponent<'input', FormInputProps>

FormInput.displayName = 'FormInput'
