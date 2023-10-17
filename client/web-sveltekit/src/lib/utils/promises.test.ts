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
        it('correctly updates each store for resolved initial promises', async () => {
            const { pending, value, error, set } = createPromiseStore<number>()
            set(Promise.resolve(1))

            expect(get(pending)).toBe(true)
            expect(get(value)).toBe(null)
            expect(get(error)).toBe(null)

            await vi.runOnlyPendingTimersAsync()

            expect(get(pending)).toBe(false)
            expect(get(value)).toBe(1)
            expect(get(error)).toBe(null)
        })

        it('correctly updates each store for rejected initial promises', async () => {
            const { pending, value, error, set } = createPromiseStore<number>()
            set(Promise.reject(1))

            expect(get(pending)).toBe(true)
            expect(get(value)).toBe(null)
            expect(get(error)).toBe(null)

            await vi.runOnlyPendingTimersAsync()

            expect(get(pending)).toBe(false)
            expect(get(value)).toBe(null)
            expect(get(error)).toBe(1)
        })
    })

    describe('updates', () => {
        it('updates the store values when a new promise is set', async () => {
            const { pending, value, error, set } = createPromiseStore<number>()
            set(Promise.resolve(1))
            await vi.runOnlyPendingTimersAsync()
            expect(get(pending)).toBe(false)

            set(Promise.reject(2))
            expect(get(pending)).toBe(true)

            await vi.runOnlyPendingTimersAsync()

            expect(get(pending)).toBe(false)
            expect(get(value)).toBe(null)
            expect(get(error)).toBe(2)

            set(Promise.resolve(3))
            expect(get(pending)).toBe(true)

            await vi.runOnlyPendingTimersAsync()

            expect(get(pending)).toBe(false)
            expect(get(value)).toBe(3)
            expect(get(error)).toBe(null)
        })

        it('updates the store with the latest resolved promise', async () => {
            const { pending, value, set } = createPromiseStore<number>()
            set(Promise.resolve(1))
            set(Promise.resolve(2))

            await vi.runOnlyPendingTimersAsync()

            expect(get(pending)).toBe(false)
            expect(get(value)).toBe(2)
        })

        it('retains the old value while a new promise is resolved', async () => {
            const { pending, value, latestValue, set } = createPromiseStore<number>()
            set(Promise.resolve(1))
            await vi.runOnlyPendingTimersAsync()

            set(Promise.resolve(2))

            expect(get(pending)).toBe(true)
            expect(get(value)).toBe(null)
            expect(get(latestValue)).toBe(1)
        })

        it('retains the old error while a new promise is resolved', async () => {
            const { pending, error, latestError, set } = createPromiseStore<number>()
            set(Promise.reject(1))
            await vi.runOnlyPendingTimersAsync()

            set(Promise.resolve(2))

            expect(get(pending)).toBe(true)
            expect(get(error)).toBe(null)
            expect(get(latestError)).toBe(1)
        })
    })
})
