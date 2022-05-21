import { InputStatus } from '..'

interface GetInputStatusProps {
    isValid?: boolean
    isError?: boolean
}

export function getInputStatus(props: GetInputStatusProps): InputStatus {
    const { isError, isValid } = props

    if (isError) {
        return InputStatus.error
    }

    if (isValid) {
        return InputStatus.valid
    }

    return InputStatus.initial
}
