import { FormInputStatus, FormInputStatusType } from '..'

interface GetFormInputStatusProps {
    isValid?: boolean
    isError?: boolean
    isLoading?: boolean
}

export function getFormInputStatus(props: GetFormInputStatusProps): FormInputStatusType {
    const { isLoading, isError, isValid } = props

    if (isLoading) {
        return FormInputStatus.loading
    }

    if (isError) {
        return FormInputStatus.error
    }

    if (isValid) {
        return FormInputStatus.valid
    }

    return FormInputStatus.initial
}
