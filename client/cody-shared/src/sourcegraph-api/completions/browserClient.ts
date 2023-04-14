import { fetchEventSource } from '@microsoft/fetch-event-source'

import { SourcegraphCompletionsClient } from './client'
import type { Event, CompletionParameters, CompletionCallbacks, CodeCompletionResponse } from './types'

export class SourcegraphBrowserCompletionsClient extends SourcegraphCompletionsClient {
    public complete(): Promise<CodeCompletionResponse> {
        throw new Error('SourcegraphBrowserCompletionsClient.complete not implemented')
    }

    public stream(params: CompletionParameters, cb: CompletionCallbacks): () => void {
        const abort = new AbortController()
        const headersInstance = new Headers(this.customHeaders as HeadersInit)
        headersInstance.set('Content-Type', 'application/json; charset=utf-8')
        if (this.accessToken) {
            headersInstance.set('Authorization', `token ${this.accessToken}`)
        }
        fetchEventSource(this.completionsEndpoint, {
            method: 'POST',
            headers: Object.fromEntries(headersInstance.entries()),
            body: JSON.stringify(params),
            signal: abort.signal,
            onmessage: message => {
                const data: Event = { ...JSON.parse(message.data), type: message.event }
                this.sendEvents([data], cb)
            },
            onerror(error) {
                console.error(error)
            },
        }).catch(error => {
            console.error(error)
        })
        return () => abort.abort()
    }
}
