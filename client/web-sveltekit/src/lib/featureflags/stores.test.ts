import { describe, test, vi, expect } from 'vitest'

import { mockFeatureFlags, unmockFeatureFlags, useFakeTimers, useRealTimers } from '$mocks'

import { createFeatureFlagStore, featureFlag } from './stores'

describe('featureflags', () => {
    describe('createFeatureFlagStore()', () => {
        test('update feature flags periodically', async () => {
            useFakeTimers()

            const store = createFeatureFlagStore(
                [{ name: 'sentinel', value: true }],
                vi
                    .fn()
                    .mockResolvedValueOnce([{ name: 'sentinel', value: false }])
                    .mockResolvedValueOnce([{ name: 'sentinel', value: true }])
            )

            const sub = vi.fn()
            store.subscribe(sub)
            expect(sub).toHaveBeenLastCalledWith([{ name: 'sentinel', value: true }])

            await vi.advanceTimersToNextTimerAsync()
            expect(sub).toHaveBeenLastCalledWith([{ name: 'sentinel', value: false }])

            await vi.advanceTimersToNextTimerAsync()
            expect(sub).toHaveBeenLastCalledWith([{ name: 'sentinel', value: true }])

            useRealTimers()
        })
    })

    describe('featureFlag()', () => {
        test('returns the current feature flag value', () => {
            mockFeatureFlags({ sentinel: false })

            const store = featureFlag('sentinel')

            const sub = vi.fn()
            store.subscribe(sub)
            expect(sub).toHaveBeenLastCalledWith(false)

            mockFeatureFlags({ sentinel: true })
            expect(sub).toHaveBeenLastCalledWith(true)

            unmockFeatureFlags()
        })
    })
})
