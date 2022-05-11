import { InputStatus, InputProps } from '@sourcegraph/wildcard'

import { useFieldAPI } from './hooks/useField'

function getDefaultInputStatus<T>({ meta }: useFieldAPI<T>): InputStatus {
    const isValidated = meta.initialValue || meta.touched

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

function getDefaultInputError<T>({ meta }: useFieldAPI<T>): Pick<InputProps, 'error'> {
    return meta.touched && meta.error
}

type GetDefaultInputPropsResult<T> = Pick<InputProps, 'error' | 'status'> & useFieldAPI<T>['input']

export function getDefaultInputProps<T>(field: useFieldAPI<T>): GetDefaultInputPropsResult<T> {
    return {
        status: getDefaultInputStatus(field),
        error: getDefaultInputError(field),
        ...field.input,
    }
}
