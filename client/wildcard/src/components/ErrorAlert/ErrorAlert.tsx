import React, { type HTMLAttributes } from 'react'

import { Alert, type AlertProps } from '../Alert/Alert'
import { ErrorMessage } from '../ErrorMessage'

export type ErrorAlertProps = {
    /**
     * An Error-like object or a string.
     */
    error: unknown

    /**
     * Optional prefix for the message
     */
    prefix?: string

    className?: string
    style?: React.CSSProperties

    /**
     * The Alert variant to display. Defaults to "danger"
     */
    variant?: AlertProps['variant']
} & HTMLAttributes<HTMLDivElement>

/**
 * Renders a given `Error` object as a Wildcard alert.
 *
 * The error message is optimistically formatted as markdown to enrich links,
 * bullet points, respect line breaks, code and bolded elements.
 * Made to work with Go `multierror`.
 */
export const ErrorAlert: React.FunctionComponent<React.PropsWithChildren<ErrorAlertProps>> = ({
    error,
    className,
    prefix,
    variant = 'danger',
    ...rest
}) => {
    prefix = prefix?.trim().replace(/:+$/, '')
    return (
        <Alert className={className} variant={variant} {...rest}>
            {prefix && <strong>{prefix}:</strong>} <ErrorMessage error={error} />
        </Alert>
    )
}
