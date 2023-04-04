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
    text: string
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
