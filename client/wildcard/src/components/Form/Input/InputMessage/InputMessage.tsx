import React from 'react'

import classNames from 'classnames'

export interface InputMessageProps {
    isValid?: boolean
    className?: string
}

/**
 * A styled message to render below an input field.
 * Can be styled differently based on the field's validity.
 */
export const InputMessage: React.FunctionComponent<React.PropsWithChildren<InputMessageProps>> = props => {
    const { isValid, className, children, ...rest } = props

    const isError = isValid === false
    const messageClassName = classNames(
        'form-text font-weight-normal mt-2',
        isError ? 'text-danger' : 'text-muted',
        className
    )

    return (
        <small role={isError ? 'alert' : undefined} className={messageClassName} {...rest}>
            {children}
        </small>
    )
}
