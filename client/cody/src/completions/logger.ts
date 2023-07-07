import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { logEvent } from '../event-logger'

interface CompletionEvent {
    params: {
        type: 'inline' | 'manual'
        multilineMode: null | 'block'
        providerIdentifier: string
        contextSummary: {
            embeddings?: number
            local?: number
            duration: number
        }
        languageId: string
    }
    // The timestamp when the request started
    startedAt: number
    // The timestamp of when the suggestion was first displayed to a users
    // screen
    suggestedAt: number | null
    // The timestamp of when the suggestion was logged to our analytics backend
    // This is to avoid double-logging
    suggestionLoggedAt: number | null
    // The timestamp of when a completion was accepted and logged to our backend
    acceptedAt: number | null
    // When set, the completion will always be marked as `read`. This helps us
    // to avoid not counting a suggested event in case where the user accepts
    // the completion below the default timeout
    forceRead: boolean
}

const READ_TIMEOUT = 750

const displayedCompletions = new LRUCache<string, CompletionEvent>({
    max: 100, // Maximum number of completions that we are keeping track of
})

export function logCompletionEvent(name: string, params?: unknown): void {
    logEvent(`CodyVSCodeExtension:completion:${name}`, params, params)
}

export function start(params: CompletionEvent['params']): string {
    const id = createId()
    displayedCompletions.set(id, {
        params,
        startedAt: Date.now(),
        suggestedAt: null,
        suggestionLoggedAt: null,
        acceptedAt: null,
        forceRead: false,
    })

    logCompletionEvent('started', params)

    return id
}

// Suggested completions will not logged individually. Instead, we log them when
// we either hide them again (they are NOT accepted) or when they ARE accepted.
// This way, we can calculate the duration they were actually visible.
export function suggest(id: string): void {
    const event = displayedCompletions.get(id)
    if (event) {
        event.suggestedAt = Date.now()
    }
}

export function accept(id: string, lines: number): void {
    const completionEvent = displayedCompletions.get(id)
    if (!completionEvent || completionEvent.acceptedAt) {
        return
    }
    completionEvent.forceRead = true
    completionEvent.acceptedAt = Date.now()

    logSuggestionEvent()
    logCompletionEvent('accepted', {
        ...completionEvent.params,
        lines,
        otherCompletionProviderEnabled: otherCompletionProviderEnabled(),
    })
}

export function noResponse(id: string): void {
    const completionEvent = displayedCompletions.get(id)
    logCompletionEvent('noResponse', completionEvent?.params ?? {})
}

/**
 * This callback should be triggered whenever VS Code tries to highlight a new
 * completion and it's
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
    // eslint-disable-next-line ban/ban
    displayedCompletions.forEach(completionEvent => {
        const { suggestedAt, suggestionLoggedAt, startedAt, params, forceRead } = completionEvent

        // Only log events that were already suggested to the user and have not
        // been logged yet.
        if (!suggestedAt || suggestionLoggedAt) {
            return
        }
        completionEvent.suggestionLoggedAt = now

        const latency = suggestedAt - startedAt
        const displayDuration = now - suggestedAt
        const read = displayDuration >= READ_TIMEOUT

        logCompletionEvent('suggested', {
            ...params,
            latency,
            displayDuration,
            read: forceRead || read,
            otherCompletionProviderEnabled: otherCompletionProviderEnabled(),
        })
    })

    // Completions are kept in the LRU cache for longer. This is because they
    // can still become visible if e.g. they are served from the cache and we
    // need to retain the ability to mark them as seen
}

const otherCompletionProviders = [
    'GitHub.copilot',
    'GitHub.copilot-nightly',
    'TabNine.tabnine-vscode',
    'TabNine.tabnine-vscode-self-hosted-updater',
    'AmazonWebServices.aws-toolkit-vscode', // Includes CodeWhisperer
    'Codeium.codeium',
    'Codeium.codeium-enterprise-updater',
    'CodeComplete.codecomplete-vscode',
    'Venthe.fauxpilot',
    'TabbyML.vscode-tabby',
]
function otherCompletionProviderEnabled(): boolean {
    return !!otherCompletionProviders.find(id => vscode.extensions.getExtension(id)?.isActive)
}
