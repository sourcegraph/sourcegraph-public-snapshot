export interface DoneEvent {
    type: 'done'
}

export interface CompletionEvent {
    type: 'completion'
    completion: string
}

export interface ErrorEvent {
    type: 'error'
    error: string
}

export type Event = DoneEvent | CompletionEvent | ErrorEvent

export interface Message {
    speaker: 'human' | 'assistant'
    text?: string
}

export interface CodeCompletionResponse {
    completion: string
    stop: string | null
    stopReason: string
    truncated: boolean
    exception: string | null
    logID: string
}

export interface CodeCompletionParameters {
    prompt: string
    temperature: number
    maxTokensToSample: number
    stopSequences: string[]
    topK: number
    topP: number
    model?: string
}

export interface CompletionParameters {
    messages: Message[]
    temperature: number
    maxTokensToSample: number
    topK: number
    topP: number
}

export interface CompletionCallbacks {
    onChange: (text: string) => void
    onComplete: () => void
    onError: (message: string) => void
}
