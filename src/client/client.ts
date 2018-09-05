import { ClientKey } from '@sourcegraph/sourcegraph.proposed/module/environment/controller'
import { Trace } from '@sourcegraph/sourcegraph.proposed/module/jsonrpc2/trace'
import { isEqual } from 'lodash'

/** Client options persisted in localStorage so that they survive across page reloads. */
interface SavedClientOptions {
    /** Unique identifier for this client. */
    key: ClientKey

    /** Timestamp of when this record was last updated. */
    at: number

    trace: Trace
}

const STORAGE_KEY = 'savedClientSettings'

function getAll(): SavedClientOptions[] {
    try {
        const data = localStorage.getItem(STORAGE_KEY)
        if (data) {
            const all = JSON.parse(data)
            if (!Array.isArray(all)) {
                throw new Error('invalid')
            }
            return all
        }
    } catch (err) {
        localStorage.removeItem(STORAGE_KEY)
    }
    return []
}

/** Read the saved trace settings for a client. */
export function getSavedClientTrace(key: ClientKey): Trace {
    const entry = getAll().find(e => isEqual(e.key, key))
    return entry && typeof entry.trace !== 'undefined' ? entry.trace : Trace.Off
}

/** Update the saved trace settings for a client. */
export function updateSavedClientTrace(key: ClientKey, value: Trace): void {
    const entries = [...getAll().filter(e => !isEqual(e.key, key)), { key, at: Date.now(), trace: value }]
    localStorage.setItem(STORAGE_KEY, JSON.stringify(entries))
}
