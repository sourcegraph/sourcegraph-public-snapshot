import http from 'http'
import https from 'https'

import { isError } from '../../utils'
import { toPartialUtf8String } from '../utils'

import { SourcegraphCompletionsClient } from './client'
import { parseEvents } from './parse'
import { CompletionParameters, CompletionCallbacks, CodeCompletionParameters, CodeCompletionResponse } from './types'

export class SourcegraphNodeCompletionsClient extends SourcegraphCompletionsClient {
    public async complete(params: CodeCompletionParameters, abortSignal: AbortSignal): Promise<CodeCompletionResponse> {
        const requestFn = this.codeCompletionsEndpoint.startsWith('https://') ? https.request : http.request
        const headersInstance = new Headers(this.config.customHeaders as HeadersInit)
        headersInstance.set('Content-Type', 'application/json')
        if (this.config.accessToken) {
            headersInstance.set('Authorization', `token ${this.config.accessToken}`)
        }
        const completion = await new Promise<CodeCompletionResponse>((resolve, reject) => {
            const req = requestFn(
                this.codeCompletionsEndpoint,
                {
                    method: 'POST',
                    headers: Object.fromEntries(headersInstance.entries()),
                    // So we can send requests to the Sourcegraph local development instance, which has an incompatible cert.
                    rejectUnauthorized: !this.config.debug,
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

            abortSignal.addEventListener('abort', () => {
                req.destroy()
                reject(new Error('aborted'))
            })
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
                    ...(this.config.accessToken ? { Authorization: `token ${this.config.accessToken}` } : null),
                    ...this.config.customHeaders,
                },
                // So we can send requests to the Sourcegraph local development instance, which has an incompatible cert.
                rejectUnauthorized: !this.config.debug,
            },
            (res: http.IncomingMessage) => {
                if (res.statusCode === undefined) {
                    throw new Error('no status code present')
                }
                // For failed requests, we just want to read the entire body and
                // ultimately return it to the error callback.
                if (res.statusCode >= 400) {
                    // Bytes which have not been decoded as UTF-8 text
                    let bufferBin = Buffer.of()
                    // Text which has not been decoded as a server-sent event (SSE)
                    let errorMessage = ''
                    res.on('data', chunk => {
                        if (!(chunk instanceof Buffer)) {
                            throw new TypeError('expected chunk to be a Buffer')
                        }
                        // Messages are expected to be UTF-8, but a chunk can terminate
                        // in the middle of a character
                        const { str, buf } = toPartialUtf8String(Buffer.concat([bufferBin, chunk]))
                        errorMessage += str
                        bufferBin = buf
                    })

                    res.on('error', e => cb.onError(e.message))
                    res.on('end', () => cb.onError(errorMessage))
                    return
                }

                // Bytes which have not been decoded as UTF-8 text
                let bufferBin = Buffer.of()
                // Text which has not been decoded as a server-sent event (SSE)
                let bufferText = ''

                res.on('data', chunk => {
                    if (!(chunk instanceof Buffer)) {
                        throw new TypeError('expected chunk to be a Buffer')
                    }
                    // text/event-stream messages are always UTF-8, but a chunk
                    // may terminate in the middle of a character
                    const { str, buf } = toPartialUtf8String(Buffer.concat([bufferBin, chunk]))
                    bufferText += str
                    bufferBin = buf

                    const parseResult = parseEvents(bufferText)
                    if (isError(parseResult)) {
                        console.error(parseResult)
                        return
                    }

                    this.sendEvents(parseResult.events, cb)
                    bufferText = parseResult.remainingBuffer
                })

                res.on('error', e => cb.onError(e.message))
            }
        )

        request.write(JSON.stringify(params))
        request.end()

        return () => request.destroy()
    }
}
