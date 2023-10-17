import React, { useMemo } from 'react'

import { getReasonPhrase } from 'http-status-codes'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'

import type { WebhookLogFields } from '../../graphql-operations'

import styles from './MessagePanel.module.scss'

type WebhookLogMessageFields = Pick<WebhookLogFields['request'], 'headers' | 'body'>
type WebhookLogRequestFields = Pick<WebhookLogFields['request'], 'method' | 'url' | 'version'>

export interface Props {
    className?: string
    message: WebhookLogMessageFields
    // A HTTP message can be either a request or a response; if it's a response,
    // then we're only interested in the status code here to render the first
    // line of the "response".
    requestOrStatusCode: WebhookLogRequestFields | number
}

export const MessagePanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    message,
    requestOrStatusCode,
}) => {
    const [headers, language, body] = useMemo(() => {
        const headers = []
        let language = 'nohighlight'
        let body = message.body

        for (const header of message.headers) {
            if (
                header.name.toLowerCase() === 'content-type' &&
                header.values.find(value => value.includes('/json')) !== undefined
            ) {
                language = 'json'
                body = JSON.stringify(JSON.parse(message.body), null, 2)
            }

            headers.push(...header.values.map(value => `${header.name}: ${value}`))
        }

        // Since the headers aren't in any useful order when they're returned
        // from the backend, let's just sort them alphabetically.
        headers.sort()

        // We want to prepend either the request line or the status line,
        // depending on what type of message this is.
        if (typeof requestOrStatusCode === 'number') {
            let reason
            try {
                reason = ' ' + getReasonPhrase(requestOrStatusCode)
            } catch {
                reason = ''
            }

            headers.unshift(`HTTP/1.1 ${requestOrStatusCode}${reason}`)
        } else {
            headers.unshift(`${requestOrStatusCode.method} ${requestOrStatusCode.url} ${requestOrStatusCode.version}`)
        }

        return [headers.join('\n'), language, body]
    }, [message.body, message.headers, requestOrStatusCode])

    return (
        <div className={className}>
            <CodeSnippet language="http" code={headers} />
            <CodeSnippet className={styles.messageBody} language={language} code={body} />
        </div>
    )
}
