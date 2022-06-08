import React from 'react'

import { renderHook, waitFor } from '@testing-library/react'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { useFeatureFlag } from './useFeatureFlag'

describe('useFeatureFlag', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName
    const ERROR_FLAG = 'error-flag' as FeatureFlagName
    const NON_EXISTING_FLAG = 'non-existing-flag' as FeatureFlagName
    interface WrapperProps {
        flagName: FeatureFlagName
    }
    const setup = (
        initialFlagName: FeatureFlagName,
        defaultValue = false,
        refetchInterval?: number,
        nextOverrides?: Partial<Record<FeatureFlagName, boolean | Error>>
    ) =>
        renderHook<ReturnType<typeof useFeatureFlag>, WrapperProps>(
            ({ flagName }) => useFeatureFlag(flagName, defaultValue),
            {
                wrapper: ({ children }: React.PropsWithChildren<{}>) => {
                    // eslint-disable-next-line react-hooks/rules-of-hooks
                    const [overrides, setOverrides] = React.useState({
                        [ENABLED_FLAG]: true,
                        [DISABLED_FLAG]: false,
                        [ERROR_FLAG]: new Error('Some error'),
                    })

                    // eslint-disable-next-line react-hooks/rules-of-hooks
                    React.useEffect(() => {
                        setTimeout(() => {
                            setOverrides(current => nextOverrides ?? current)
                        }, refetchInterval)
                    }, [])

                    return (
                        <MockedFeatureFlagsProvider overrides={overrides} refetchInterval={refetchInterval}>
                            {children}
                        </MockedFeatureFlagsProvider>
                    )
                },
                initialProps: {
                    flagName: initialFlagName,
                },
            }
        )

    it('returns [false] value correctly', async () => {
        const state = setup(DISABLED_FLAG)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('returns [defaultValue=true] correctly', async () => {
        const state = setup(NON_EXISTING_FLAG, true)

        await waitFor(() => expect(state.result.current).toEqual(expect.arrayContaining([true, 'loaded'])))
    })

    it('returns [defaultValue=false] correctly', async () => {
        const state = setup(NON_EXISTING_FLAG, false)

        await waitFor(() => expect(state.result.current).toEqual(expect.arrayContaining([false, 'loaded'])))
    })

    it('returns [true] value correctly', async () => {
        const state = setup(ENABLED_FLAG)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))
    })

    it('updates on value change', async () => {
        const state = setup(ENABLED_FLAG, false, 100, { [ENABLED_FLAG]: false })
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))

        // Rerender and wait for new state
        state.rerender({ flagName: ENABLED_FLAG })

        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('updates when feature flag prop changes', async () => {
        const state = setup(ENABLED_FLAG, false, undefined, {})
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])
        // Loaded state
        await waitFor(() => expect(state.result.current).toStrictEqual([true, 'loaded', undefined]))

        // Rerender and wait for new state
        state.rerender({ flagName: DISABLED_FLAG })
        await waitFor(() => expect(state.result.current).toStrictEqual([false, 'loaded', undefined]))
    })

    it('returns "error" when no context parent', () => {
        const state = renderHook(() => useFeatureFlag(ENABLED_FLAG))

        expect(state.result.current).toEqual(expect.arrayContaining([false, 'error']))
    })

    it('returns "error" when unhandled error', async () => {
        const state = setup(ERROR_FLAG)

        await waitFor(() => expect(state.result.current).toEqual(expect.arrayContaining([false, 'error'])))
    })
})
