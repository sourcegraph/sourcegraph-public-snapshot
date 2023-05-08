import { fetchEventSource } from '@microsoft/fetch-event-source'

import { SourcegraphCompletionsClient } from './client'
import type { Event, CompletionParameters, CompletionCallbacks, CodeCompletionResponse } from './types'

export class SourcegraphBrowserCompletionsClient extends SourcegraphCompletionsClient {
    public complete(): Promise<CodeCompletionResponse> {
        throw new Error('SourcegraphBrowserCompletionsClient.complete not implemented')
    }

    public stream(params: CompletionParameters, cb: CompletionCallbacks): () => void {
        const abort = new AbortController()
        const headersInstance = new Headers(this.config.customHeaders as HeadersInit)
        headersInstance.set('Content-Type', 'application/json; charset=utf-8')
        if (this.config.accessToken) {
            headersInstance.set('Authorization', `token ${this.config.accessToken}`)
        }
        fetchEventSource(this.completionsEndpoint, {
            method: 'POST',
            headers: Object.fromEntries(headersInstance.entries()),
            body: JSON.stringify(params),
            signal: abort.signal,
            async onopen(response) {
                if (!response.ok && response.headers.get('content-type') !== 'text/event-stream') {
                    let errorMessage: null | string = null
                    try {
                        errorMessage = await response.text()
                    } catch (error) {
                        // We show the generic error message in this case
                        console.error(error)
                    }
                    cb.onError(
                        errorMessage === null || errorMessage.length === 0
                            ? `Request failed with status code ${response.status}`
                            : errorMessage,
                        response.status
                    )
                    abort.abort()
                    return
                }
            },
            onmessage: message => {
                const data: Event = { ...JSON.parse(message.data), type: message.event }
                this.sendEvents([data], cb)
            },
            onerror(error) {
                cb.onError(error.message)
                abort.abort()
                console.error(error)
            },
        }).catch(error => {
            cb.onError(error.message)
            abort.abort()
            console.error(error)
        })
        return () => abort.abort()
    }
}
