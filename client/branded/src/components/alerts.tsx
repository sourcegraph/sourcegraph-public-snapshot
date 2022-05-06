import React, { HTMLAttributes } from 'react'

import { upperFirst } from 'lodash'

import { asError, renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Alert, AlertProps } from '@sourcegraph/wildcard'

export const renderError = (error: unknown): string =>
    renderMarkdown(upperFirst((asError(error).message || 'Unknown Error').replace(/\t/g, '')), { breaks: true })
        .trim()
        .replace(/^<p>/, '')
        .replace(/<\/p>$/, '')

export const ErrorMessage: React.FunctionComponent<React.PropsWithChildren<{ className?: string; error: unknown }>> = ({
    className,
    error,
}) => <Markdown className={className} wrapper="span" dangerousInnerHTML={renderError(error)} />

export type ErrorAlertProps = {
    /**
     * An Error-like object or a string.
     */
    error: unknown

    /**
     * Whether to show an icon.
     *
     * @default true
     */
    icon?: boolean

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
    icon = true,
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
