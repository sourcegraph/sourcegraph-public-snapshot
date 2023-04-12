import http from 'http'
import https from 'https'

import { isError } from '../../utils'

import { SourcegraphCompletionsClient } from './client'
import { parseEvents } from './parse'
import { CompletionParameters, CompletionCallbacks } from './types'

export class SourcegraphNodeCompletionsClient extends SourcegraphCompletionsClient {
    public stream(params: CompletionParameters, cb: CompletionCallbacks): () => void {
        const requestFn = this.completionsEndpoint.startsWith('https://') ? https.request : http.request

        const request = requestFn(
            this.completionsEndpoint,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...(this.accessToken ? { Authorization: `token ${this.accessToken}` } : null),
                    ...this.customHeaders,
                },
                // So we can send requests to the Sourcegraph local development instance, which has an incompatible cert.
                rejectUnauthorized: this.mode === 'production',
            },
            (res: http.IncomingMessage) => {
                let buffer = ''

                res.on('data', chunk => {
                    if (!(chunk instanceof Buffer)) {
                        throw new TypeError('expected chunk to be a Buffer')
                    }
                    buffer += chunk.toString()

                    const parseResult = parseEvents(buffer)
                    if (isError(parseResult)) {
                        console.error(parseResult)
                        return
                    }

                    this.sendEvents(parseResult.events, cb)
                    buffer = parseResult.remainingBuffer
                })

                res.on('error', e => cb.onError(e.message))
            }
        )

        request.write(JSON.stringify(params))
        request.end()

        return () => request.destroy()
    }
}
