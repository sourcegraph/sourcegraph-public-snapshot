import { logEvent } from '../event-logger'

interface CompletionEvent {
    params: {
        type: 'inline' | 'manual'
        multilineMode: null | 'block'
        providerIdentifier: string
        contextSummary: {
            embeddings?: number
            local?: number
        }
        languageId: string
    }
    startedAt: number
    suggestedAt: number | null
    // When set, the completion will always be marked as `read`. This helps us
    // to avoid not counting a suggested event in case where the user accepts
    // the completion below the default timeout
    forceRead: boolean
}

const READ_TIMEOUT = 750

const displayedCompletions: Map<string, CompletionEvent> = new Map()

export function logCompletionEvent(name: string, params?: unknown): void {
    logEvent(`CodyVSCodeExtension:completion:${name}`, params, params)
}

export function start(params: CompletionEvent['params']): string {
    const id = createId()
    displayedCompletions.set(id, {
        params,
        startedAt: Date.now(),
        suggestedAt: null,
        forceRead: false,
    })

    logCompletionEvent('started', params)

    return id
}

// Suggested completions will not logged individually. Instead, we log them when we either hide them
// again (they are NOT accepted) or when they ARE accepted. This way, we can calculate the duration
// they were actually visible.
export function suggest(id: string): void {
    const event = displayedCompletions.get(id)
    if (event) {
        event.suggestedAt = Date.now()
    }
}

export function accept(id: string): void {
    const completionEvent = displayedCompletions.get(id)
    if (!completionEvent) {
        return
    }
    completionEvent.forceRead = true

    logSuggestionEvent()
    logCompletionEvent('accepted', completionEvent.params)
}

export function noResponse(id: string): void {
    const completionEvent = displayedCompletions.get(id)
    logCompletionEvent('noResponse', completionEvent?.params ?? {})
}

/**
 * This callback should be triggered whenever VS Code tries to highlight a new completion and it's
 * used to measure how long previous completions were visible.
 */
export function clear(): void {
    logSuggestionEvent()
}

function createId(): string {
    return Math.random().toString(36).slice(2, 11)
}

function logSuggestionEvent(): void {
    const now = Date.now()
    for (const completionEvent of displayedCompletions.values()) {
        const { suggestedAt, startedAt, params, forceRead } = completionEvent

        if (!suggestedAt) {
            continue
        }

        const latency = suggestedAt - startedAt
        const displayDuration = now - suggestedAt
        const read = displayDuration >= READ_TIMEOUT

        logCompletionEvent('suggested', {
            ...params,
            latency,
            displayDuration,
            read: forceRead || read,
        })
    }
    displayedCompletions.clear()
}
