import { memoize } from 'lodash'
import { writable, type Readable, type Writable } from 'svelte/store'

import { createEmptySingleSelectTreeState, type TreeState } from '$lib/TreeView'

import type { HistoryPanel_HistoryConnection } from './HistoryPanel.gql'

export const sidebarOpen = writable(true)

/**
 * Persistent, global state for the file sidebar. By keeping the state in memory we can
 * properly restore the UI when the user closes/opens the sidebar or navigates up the repo.
 */
export const getSidebarFileTreeStateForRepo = memoize(
    (_repoName: string): Writable<TreeState> => writable<TreeState>(createEmptySingleSelectTreeState()),
    repoName => repoName
)

interface HistoryPanelStoreValue {
    loading: boolean
    history?: HistoryPanel_HistoryConnection
    error?: Error
}

interface HistoryPanelStore extends Readable<HistoryPanelStoreValue> {
    capture(): HistoryPanel_HistoryConnection | null
    restore(result: HistoryPanel_HistoryConnection | null): void
    loadMore(fetch: (afterCursor: string | null) => Promise<HistoryPanel_HistoryConnection | null>): void
}

/**
 * Creates a store for properly handling history panel state. Having this logic in a separate
 * store makes it easier to handle promises.
 */
export function createHistoryPanelStore(
    initialHistory: Promise<HistoryPanel_HistoryConnection | null>
): HistoryPanelStore {
    let loading = true
    let history: HistoryPanel_HistoryConnection | null = null
    let currentPromise: Promise<HistoryPanel_HistoryConnection | null> | null = initialHistory

    const { subscribe, set, update } = writable<HistoryPanelStoreValue>({ loading })

    function processPromise(promise: Promise<HistoryPanel_HistoryConnection | null>): void {
        currentPromise = promise
        loading = true
        update(state => ({ ...state, loading, error: undefined }))
        promise.then(result => {
            if (promise === currentPromise) {
                // Don't update data when promise is "stale"
                history = result
                    ? { pageInfo: result.pageInfo, nodes: [...(history?.nodes ?? []), ...result.nodes] }
                    : { pageInfo: { hasNextPage: false, endCursor: null }, nodes: [] }
                loading = false
                set({ history, loading })
            }
        })
    }

    processPromise(initialHistory)

    return {
        subscribe,
        restore(result) {
            if (result) {
                // When we restore the list of commits from a previous state, we
                // "abort" any commit loading in progress
                currentPromise = null
                loading = false
                history = result
                set({ loading, history })
            }
        },
        loadMore(fetch) {
            if (loading || !history || !history.pageInfo.hasNextPage) {
                return
            }
            processPromise(fetch(history.pageInfo.endCursor))
        },
        capture() {
            return history
        },
    }
}
