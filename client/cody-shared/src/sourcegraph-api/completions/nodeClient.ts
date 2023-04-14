import http from 'http'
import https from 'https'

import { isError } from '../../utils'

import { SourcegraphCompletionsClient } from './client'
import { parseEvents } from './parse'
import { CompletionParameters, CompletionCallbacks, CodeCompletionParameters, CodeCompletionResponse } from './types'

export class SourcegraphNodeCompletionsClient extends SourcegraphCompletionsClient {
    public async complete(params: CodeCompletionParameters): Promise<CodeCompletionResponse> {
        const requestFn = this.codeCompletionsEndpoint.startsWith('https://') ? https.request : http.request
        const headersInstance = new Headers(this.customHeaders as HeadersInit)
        headersInstance.set('Content-Type', 'application/json')
        if (this.accessToken) {
            headersInstance.set('Authorization', `token ${this.accessToken}`)
        }
        const completion = await new Promise<CodeCompletionResponse>((resolve, reject) => {
            const req = requestFn(
                this.codeCompletionsEndpoint,
                {
                    method: 'POST',
                    headers: Object.fromEntries(headersInstance.entries()),
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
                    })
                    res.on('end', () => {
                        req.destroy()
                        try {
                            const resp = JSON.parse(buffer) as CodeCompletionResponse
                            if (typeof resp.completion !== 'string' || typeof resp.stopReason !== 'string') {
                                reject(new Error(`response does not satisfy CodeCompletionResponse: ${buffer}`))
                            } else {
                                resolve(resp)
                            }
                        } catch (error) {
                            reject(
                                new Error(
                                    `error parsing response CodeCompletionResponse: ${error}, response text: ${buffer}`
                                )
                            )
                        }
                    })

                    res.on('error', e => {
                        req.destroy()
                        reject(e)
                    })
                }
            )
            req.write(JSON.stringify(params))
            req.end()
        })
        return completion
    }

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
