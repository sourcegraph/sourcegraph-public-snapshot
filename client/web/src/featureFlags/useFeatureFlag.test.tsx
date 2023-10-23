import { describe, expect, it } from '@jest/globals'
import { renderHook, waitFor } from '@testing-library/react'
import delay from 'delay'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { createFlagMock } from './createFlagMock'
import type { FeatureFlagName } from './featureFlags'
import { useFeatureFlag } from './useFeatureFlag'

describe('useFeatureFlag', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName
    const ERROR_FLAG = 'error-flag' as FeatureFlagName
    const MOCKS = [
        createFlagMock(ENABLED_FLAG, true),
        createFlagMock(DISABLED_FLAG, false),
        createFlagMock(ENABLED_FLAG, false),
        createFlagMock(ERROR_FLAG, new Error('oops')),
    ]

    const setup = (initialFlagName: FeatureFlagName, defaultValue = false, cacheTTL?: number) => {
        const flagStates = new Map()
        const allResults: ReturnType<typeof useFeatureFlag>[] = []

        const result = renderHook(
            ({ flagName }) => {
                const result = useFeatureFlag(flagName, defaultValue, cacheTTL, flagStates)
                allResults.push(result)

                return result
            },
            {
                wrapper: ({ children }) => <MockedTestProvider mocks={MOCKS}>{children}</MockedTestProvider>,
                initialProps: {
                    flagName: initialFlagName,
                },
            }
        )

        return {
            ...result,
            allResults,
        }
    }

    it('returns [false] value correctly', async () => {
        const state = setup(DISABLED_FLAG)
        // Initial state
        expect(state.allResults[0]).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('returns [true] value correctly', async () => {
        const state = setup(ENABLED_FLAG)
        // Initial state
        expect(state.allResults[0]).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))
    })

    it('updates on value change', async () => {
        const state = setup(ENABLED_FLAG, false, 100)
        // Initial state
        expect(state.allResults[0]).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))

        // The cached value is used because of TTL.
        await delay(50)
        state.rerender({ flagName: ENABLED_FLAG })
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))

        // The new value is fetched because the cache is stale.
        await delay(50)
        state.rerender({ flagName: ENABLED_FLAG })
        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('updates when feature flag prop changes', async () => {
        const state = setup(ENABLED_FLAG, false, undefined)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])
        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))

        // Rerender and wait for new state
        state.rerender({ flagName: DISABLED_FLAG })
        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('returns "error" when unhandled error', async () => {
        const state = setup(ERROR_FLAG)

        await waitFor(() => expect(state.result.current).toEqual(expect.arrayContaining([false, 'error'])))
    })
})
