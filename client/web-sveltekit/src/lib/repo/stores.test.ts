import { get } from 'svelte/store'
import { expect, describe, it, vi, beforeAll, afterAll } from 'vitest'

import type { HistoryResult } from '$lib/graphql-operations'
import { createHistoryResults } from '$testdata'

import { createHistoryPanelStore } from './stores'

describe('createHistoryPanelStore', () => {
    const historyResults: HistoryResult[] = createHistoryResults(2, 2)

    beforeAll(() => {
        vi.useFakeTimers()
    })

    afterAll(() => {
        vi.useRealTimers()
    })

    it('resolves the initial promise', async () => {
        const [initial] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))

        expect(get(store)).toEqual({ loading: true })

        await vi.runAllTimersAsync()
        expect(get(store)).toEqual({ loading: false, history: initial })
    })

    it('prefers restored data over initial data before it loaded', async () => {
        const [initial, restore] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))

        store.restore(restore)
        await vi.runAllTimersAsync()

        expect(get(store)).toEqual({ loading: false, history: restore })
    })

    it('prefers restored data over initial data after it loaded', async () => {
        const [initial, restore] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))

        await vi.runAllTimersAsync()
        store.restore(restore)

        expect(get(store).history).toEqual(restore)
    })

    it('appends "load more" data to existing data', async () => {
        const [initial, more] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))
        await vi.runAllTimersAsync()

        store.loadMore(async () => more)
        await vi.runAllTimersAsync()

        expect(get(store).history).toEqual({ nodes: [...initial.nodes, ...more.nodes], pageInfo: more.pageInfo })
    })

    it('does not append more data if there is no next page', async () => {
        const [initial, more] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))
        await vi.runAllTimersAsync()

        store.loadMore(async () => more)
        await vi.runAllTimersAsync()

        store.loadMore(async () => createHistoryResults(1, 2)[0])
        await vi.runAllTimersAsync()

        expect(get(store).history).toEqual({ nodes: [...initial.nodes, ...more.nodes], pageInfo: more.pageInfo })
    })

    it('does not append more data if the initial data is still loading', async () => {
        const [initial, more] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))
        store.loadMore(async () => more)
        await vi.runAllTimersAsync()

        expect(get(store).history).toEqual(initial)
    })

    it('does not append more data if other data is still loading', async () => {
        const [initial, more, evenMore] = historyResults
        const store = createHistoryPanelStore(Promise.resolve(initial))
        await vi.runAllTimersAsync()

        store.loadMore(async () => more)
        store.loadMore(async () => evenMore)
        await vi.runAllTimersAsync()

        expect(get(store).history).toEqual({ nodes: [...initial.nodes, ...more.nodes], pageInfo: more.pageInfo })
    })
})
