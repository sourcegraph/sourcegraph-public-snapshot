import React from 'react'

export interface FormFieldMessageProps {
    isValid?: boolean
}

/**
 * Returns the global CSS classes to apply to the message element based on associated validity.
 */
export const getMessageStyle = ({ isValid }: FormFieldMessageProps): string => {
    if (isValid === undefined) {
        return 'field-message'
    }

    if (isValid) {
        return 'valid-feedback'
    }

    return 'invalid-feedback'
}

/**
 * A simple message to render alongside a form field.
 * Can be styled differently based on the field's validity.
 */
export const FormFieldMessage: React.FunctionComponent<React.PropsWithChildren<FormFieldMessageProps>> = ({
    isValid,
    children,
}) => <small className={getMessageStyle({ isValid })}>{children}</small>
