import { fetchEventSource } from '@microsoft/fetch-event-source'

export interface CompletionRequest {
    messages: { speaker: 'human' | 'assistant'; text: string }[]
    temperature: number
    maxTokensToSample: number
    topK: number
    topP: number
}

export const DEFAULT_CHAT_COMPLETION_PARAMETERS: Omit<CompletionRequest, 'messages'> = {
    temperature: 0.2,
    maxTokensToSample: 1000,
    topK: -1,
    topP: -1,
}

export function getCodyCompletionOneShot(params: CompletionRequest, abortSignal: AbortSignal | null): Promise<string> {
    return new Promise<string>((resolve, reject) => {
        let lastCompletion: string | undefined
        fetchEventSource('/.api/completions/stream', {
            method: 'POST',
            headers: { 'X-Requested-With': 'Sourcegraph', 'Content-Type': 'application/json; charset=utf-8' },
            body: JSON.stringify(params),
            onmessage(message) {
                if (message.event === 'completion') {
                    const data = JSON.parse(message.data) as { completion: string }
                    lastCompletion = data.completion
                }
            },
            onclose() {
                if (lastCompletion) {
                    resolve(lastCompletion)
                } else {
                    reject(new Error('no completion received'))
                }
            },
            onerror(error) {
                reject(error)
            },
            signal: abortSignal,
        }).catch(error => reject(error))
    })
}
