import React from 'react'

import { upperFirst } from 'lodash'


import { Markdown } from '../Markdown'
import { asError, renderMarkdown } from '../../utils'

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
