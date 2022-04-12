import { InputStatus, InputProps } from '@sourcegraph/wildcard'

import { Field } from './hooks/useField'

function getDefaultInputStatus({ meta }: Field): InputStatus {
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

function getDefaultInputError({ meta }: Field): Pick<InputProps, 'error'> {
    return meta.touched && meta.error
}

type GetDefaultInputPropsResult = Pick<InputProps, 'error' | 'status'> & Field['input']

export function getDefaultInputProps(field: Field): GetDefaultInputPropsResult {
    return {
        status: getDefaultInputStatus(field),
        error: getDefaultInputError(field),
        ...field.input,
    }
}
