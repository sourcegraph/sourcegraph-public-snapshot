import { InputStatus, type InputProps } from '../Input'

import type { useFieldAPI } from './hooks'

export function getDefaultInputStatus<T>({ meta }: useFieldAPI<T>, getValue?: (value: T) => unknown): InputStatus {
    const initialValue = getValue
        ? getValue(meta.initialValue)
        : Array.isArray(meta.initialValue)
        ? meta.initialValue.length
        : meta.initialValue
    const isValidated = initialValue || meta.touched

    if (meta.validState === 'CHECKING') {
        return InputStatus.loading
    }

    if (isValidated && meta.validState === 'VALID') {
        return InputStatus.valid
    }

    if (isValidated && meta.error) {
        return InputStatus.error
    }

    return InputStatus.initial
}

export function getDefaultInputError<T>({ meta }: useFieldAPI<T>): InputProps['error'] {
    return (meta.touched && meta.error) || undefined
}

type GetDefaultInputPropsResult<T> = Pick<InputProps, 'error' | 'status'> & useFieldAPI<T>['input']

export function getDefaultInputProps<T>(field: useFieldAPI<T>): GetDefaultInputPropsResult<T> {
    return {
        status: getDefaultInputStatus(field),
        error: getDefaultInputError(field),
        ...field.input,
    }
}
