import React from 'react'

import { upperFirst } from 'lodash'

import { asError, renderMarkdown } from '@sourcegraph/common'

import { Markdown } from '../Markdown'

export const renderError = (error: unknown): string =>
    renderMarkdown(upperFirst((asError(error).message || 'Unknown Error').replaceAll('\t', '')), { breaks: true })
        .trim()
        .replace(/^<p>/, '')
        .replace(/<\/p>$/, '')

interface ErrorMessageProps {
    className?: string
    error: unknown
}

export const ErrorMessage: React.FunctionComponent<React.PropsWithChildren<ErrorMessageProps>> = ({
    className,
    error,
}) => <Markdown className={className} wrapper="span" dangerousInnerHTML={renderError(error)} />
