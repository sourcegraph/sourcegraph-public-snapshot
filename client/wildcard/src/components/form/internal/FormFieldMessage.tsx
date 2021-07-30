import React from 'react'

export interface FormFieldMessageProps {
    isValid?: boolean
}

export const getMessageStyle = ({ isValid }: FormFieldMessageProps): string => {
    if (isValid === undefined) {
        return 'field-message'
    }

    if (isValid) {
        return 'valid-feedback'
    }

    return 'invalid-feedback'
}

export const FormFieldMessage: React.FunctionComponent<FormFieldMessageProps> = ({ isValid, children }) => (
    <small className={getMessageStyle({ isValid })}>{children}</small>
)
