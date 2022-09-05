import { InputStatus } from '../Input'

interface GetInputStatusProps {
    isValid?: boolean
    isError?: boolean
    isLoading?: boolean
}

export function getInputStatus(props: GetInputStatusProps): InputStatus {
    const { isLoading, isError, isValid } = props

    if (isLoading) {
        return InputStatus.loading
    }

    if (isError) {
        return InputStatus.error
    }

    if (isValid) {
        return InputStatus.valid
    }

    return InputStatus.initial
}
