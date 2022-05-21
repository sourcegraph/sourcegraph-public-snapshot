import { InputProps, FormInputStatus, FormInputProps, FormInputStatusType } from '@sourcegraph/wildcard'

import { useFieldAPI } from './hooks/useField'

function getDefaultInputStatus<T>({ meta }: useFieldAPI<T>): FormInputStatusType {
    const isValidated = meta.initialValue || meta.touched

    if (meta.validState === 'CHECKING') {
        return FormInputStatus.loading
    }

    if (isValidated && meta.validState === 'VALID') {
        return FormInputStatus.valid
    }

    if (isValidated && meta.error) {
        return FormInputStatus.error
    }

    return FormInputStatus.initial
}

function getDefaultInputError<T>({ meta }: useFieldAPI<T>): Pick<InputProps, 'error'> {
    return meta.touched && meta.error
}

type GetDefaultInputPropsResult<T> = Pick<FormInputProps, 'error' | 'status'> & useFieldAPI<T>['input']

export function getDefaultInputProps<T>(field: useFieldAPI<T>): GetDefaultInputPropsResult<T> {
    return {
        status: getDefaultInputStatus(field),
        error: getDefaultInputError(field),
        ...field.input,
    }
}
