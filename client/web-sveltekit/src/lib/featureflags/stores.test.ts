import { describe, test, vi, expect } from 'vitest'

import { mockFeatureFlags, unmockFeatureFlags, useFakeTimers, useRealTimers } from '$testing/mocks'

import { createFeatureFlagStore, featureFlag } from './stores'

describe('featureflags', () => {
    describe('createFeatureFlagStore()', () => {
        test('update feature flags periodically', async () => {
            useFakeTimers()

            const store = createFeatureFlagStore(
                [{ name: 'search-debug', value: true }],
                vi
                    .fn()
                    .mockResolvedValueOnce([{ name: 'search-debug', value: false }])
                    .mockResolvedValueOnce([{ name: 'search-debug', value: true }])
            )

            const sub = vi.fn()
            store.subscribe(sub)
            expect(sub).toHaveBeenLastCalledWith([{ name: 'search-debug', value: true }])

            await vi.advanceTimersToNextTimerAsync()
            expect(sub).toHaveBeenLastCalledWith([{ name: 'search-debug', value: false }])

            await vi.advanceTimersToNextTimerAsync()
            expect(sub).toHaveBeenLastCalledWith([{ name: 'search-debug', value: true }])

            useRealTimers()
        })
    })

    describe('featureFlag()', () => {
        test('returns the current feature flag value', () => {
            mockFeatureFlags({ 'search-debug': false })

            const store = featureFlag('search-debug')

            const sub = vi.fn()
            store.subscribe(sub)
            expect(sub).toHaveBeenLastCalledWith(false)

            mockFeatureFlags({ 'search-debug': true })
            expect(sub).toHaveBeenLastCalledWith(true)

            unmockFeatureFlags()
        })
    })
})
