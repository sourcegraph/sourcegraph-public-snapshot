import { get } from 'svelte/store'
import { describe, it, vi, beforeAll, afterAll, expect } from 'vitest'

import { createPromiseStore } from './promises'

beforeAll(() => {
    vi.useFakeTimers()
})
afterAll(() => {
    vi.useRealTimers()
})

describe('createPromiseStore', () => {
    describe('initial promise', () => {
        it('correctly updates store for resolved initial promises', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.resolve(1))

            expect(get(store)).toMatchObject({
                pending: true,
                value: null,
                error: null,
            })

            await vi.runOnlyPendingTimersAsync()

            expect(get(store)).toMatchObject({
                pending: false,
                value: 1,
                error: null,
            })
        })

        it('correctly updates each store for rejected initial promises', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.reject(1))

            expect(get(store)).toMatchObject({
                pending: true,
                value: null,
                error: null,
            })

            await vi.runOnlyPendingTimersAsync()

            expect(get(store)).toMatchObject({
                pending: false,
                value: null,
                error: 1,
            })
        })
    })

    describe('updates', () => {
        it('updates the store when a new promise is set', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.resolve(1))
            await vi.runOnlyPendingTimersAsync()
            expect(get(store).pending).toBe(false)

            store.set(Promise.reject(2))
            expect(get(store).pending).toBe(true)

            await vi.runOnlyPendingTimersAsync()

            expect(get(store)).toMatchObject({
                pending: false,
                value: null,
                error: 2,
            })

            store.set(Promise.resolve(3))
            expect(get(store).pending).toBe(true)

            await vi.runOnlyPendingTimersAsync()

            expect(get(store)).toMatchObject({
                pending: false,
                value: 3,
                error: null,
            })
        })

        it('updates the store with the latest resolved promise', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.resolve(1))
            store.set(Promise.resolve(2))

            await vi.runOnlyPendingTimersAsync()

            expect(get(store)).toMatchObject({
                pending: false,
                value: 2,
                error: null,
            })
        })

        it('retains the old value while a new promise is resolved', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.resolve(1))
            await vi.runOnlyPendingTimersAsync()

            store.set(Promise.resolve(2))

            expect(get(store)).toMatchObject({
                pending: true,
                value: 1,
                error: null,
            })
        })

        it('retains the old error while a new promise is resolved', async () => {
            const store = createPromiseStore<number>()
            store.set(Promise.reject(1))
            await vi.runOnlyPendingTimersAsync()

            store.set(Promise.resolve(2))

            expect(get(store)).toMatchObject({
                pending: true,
                value: null,
                error: 1,
            })

            await vi.runOnlyPendingTimersAsync()
            expect(get(store)).toMatchObject({
                pending: false,
                value: 2,
                error: null,
            })
        })
    })
})
