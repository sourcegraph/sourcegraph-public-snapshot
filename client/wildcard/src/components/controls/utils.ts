export const getValidStyle = (isValid?: boolean): string => {
    if (isValid === undefined) {
        return ''
    }

    if (isValid) {
        return 'is-valid'
    }

    return 'is-invalid'
}

export const getMessageStyle = (isValid?: boolean): string => {
    if (isValid === undefined) {
        return 'field-message'
    }

    if (isValid) {
        return 'valid-feedback'
    }

    return 'invalid-feedback'
}
