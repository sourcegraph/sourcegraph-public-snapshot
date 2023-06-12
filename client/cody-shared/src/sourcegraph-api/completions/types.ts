export interface DoneEvent {
    type: 'done'
}

export interface CompletionEvent extends CompletionResponse {
    type: 'completion'
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

export interface CompletionResponse {
    completion: string
    stopReason: string
}

export interface CompletionParameters {
    fast?: boolean
    messages: Message[]
    maxTokensToSample: number
    temperature?: number
    stopSequences?: string[]
    topK?: number
    topP?: number
    model?: string
}

export interface CompletionCallbacks {
    onChange: (text: string) => void
    onComplete: () => void
    onError: (message: string, statusCode?: number) => void
}
