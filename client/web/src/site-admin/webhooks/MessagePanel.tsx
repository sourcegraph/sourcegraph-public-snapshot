import { getReasonPhrase } from 'http-status-codes'
import React, { useMemo } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'

import { WebhookLogMessageFields, WebhookLogRequestFields } from '../../graphql-operations'

import styles from './MessagePanel.module.scss'

export interface Props {
    className?: string
    message: WebhookLogMessageFields
    request?: WebhookLogRequestFields
    statusCode?: number
}

export const MessagePanel: React.FunctionComponent<Props> = ({ className, message, request, statusCode }) => {
    const headers = useMemo(() => {
        const headers: Map<string, { name: string; values: string[] }> = new Map()

        for (const header of message.headers) {
            headers.set(header.name.toLowerCase(), {
                name: header.name,
                values: header.values,
            })
        }

        return headers
    }, [message.headers])

    const [language, body] = useMemo((): [string, string] => {
        const contentType = headers.get('content-type')
        if (contentType) {
            // We only really ever expect JSON here, so let's just look for that
            // for now.
            if (contentType.values[0].includes('/json')) {
                try {
                    // Let's reindent the JSON, since it probably came over the
                    // wire in the minimal form.
                    return ['json', JSON.stringify(JSON.parse(message.body), null, 2)]
                } catch {
                    // Fall through to the fallback case without highlighting,
                    // since this apparently isn't JSON after all.
                }
            }
        }
        return ['nohighlight', message.body]
    }, [headers, message.body])

    const rawHeaders = useMemo(() => {
        const raw = []

        for (const { name, values } of headers.values()) {
            for (const value of values) {
                raw.push(`${name}: ${value}`)
            }
        }

        raw.sort()
        if (statusCode !== undefined) {
            let reason
            try {
                reason = ' ' + getReasonPhrase(statusCode)
            } catch {
                reason = ''
            }

            raw.unshift(`HTTP/1.1 ${statusCode}${reason}`)
        } else if (request !== undefined) {
            raw.unshift(`${request.method} ${request.url} ${request.version}`)
        }

        return raw.join('\n')
    }, [headers, request, statusCode])

    return (
        <div className={className}>
            <CodeSnippet language="http" code={rawHeaders} />
            <CodeSnippet className={styles.body} language={language} code={body} />
        </div>
    )
}
