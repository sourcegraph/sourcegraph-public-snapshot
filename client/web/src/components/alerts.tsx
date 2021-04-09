import classNames from 'classnames'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import { useHistory } from 'react-router'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

const renderError = (error: unknown): string =>
    renderMarkdown(upperFirst((asError(error).message || 'Unknown Error').replace(/\t/g, '')), { breaks: true })
        .trim()
        .replace(/^<p>/, '')
        .replace(/<\/p>$/, '')

export const ErrorMessage: React.FunctionComponent<{ error: unknown }> = ({ error }) => {
    const history = useHistory()

    return <Markdown wrapper="span" dangerousInnerHTML={renderError(error)} history={history} />
}

/**
 * Renders a given `Error` object in a Bootstrap danger alert.
 *
 * The error message is optimistically formatted as markdown to enrich links,
 * bullet points, respect line breaks, code and bolded elements.
 * Made to work with Go `multierror`.
 */
export const ErrorAlert: React.FunctionComponent<{
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
}> = ({ error, className, icon = true, prefix, ...rest }) => {
    prefix = prefix?.trim().replace(/:+$/, '')
    return (
        <div className={classNames('alert', 'alert-danger', className)} {...rest}>
            {icon && <AlertCircleIcon className="icon icon-inline" />} {prefix && <strong>{prefix}:</strong>}{' '}
            <ErrorMessage error={error} />
        </div>
    )
}
