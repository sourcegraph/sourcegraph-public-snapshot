import { renderHook } from '@testing-library/react-hooks'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { useFeatureFlag } from './useFeatureFlag'

describe('useFeatureFlag', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName
    const ERROR_FLAG = 'error-flag' as FeatureFlagName
    const setup = (initialFlagName: FeatureFlagName, refetchInterval?: number) =>
        renderHook(({ flagName }) => useFeatureFlag(flagName), {
            wrapper: function Wrapper({ children, overrides }) {
                return (
                    <MockedFeatureFlagsProvider
                        overrides={overrides as Partial<Record<FeatureFlagName, boolean>>}
                        refetchInterval={refetchInterval}
                    >
                        {children}
                    </MockedFeatureFlagsProvider>
                )
            },
            initialProps: {
                flagName: initialFlagName,
                overrides: {
                    [ENABLED_FLAG]: true,
                    [DISABLED_FLAG]: false,
                    [ERROR_FLAG]: new Error('Some error'),
                },
            },
        })

    it('returns [false] value correctly', async () => {
        const state = setup(DISABLED_FLAG)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([false, 'loaded', undefined])
    })

    it('returns [true] value correctly', async () => {
        const state = setup(ENABLED_FLAG)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([true, 'loaded', undefined])
        expect(state.result.all.length).toBe(2)
    })

    it('updates on value change', async () => {
        const state = setup(ENABLED_FLAG, 100)
        // Initial state
        expect(state.result.current).toStrictEqual([false, 'initial', undefined])

        // Loaded state
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([true, 'loaded', undefined])

        // Rerender and wait for new state
        state.rerender({ overrides: { [ENABLED_FLAG]: false }, flagName: ENABLED_FLAG })
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([false, 'loaded', undefined])
    })

    it('updates when feature flag prop changes', async () => {
        const state = setup(ENABLED_FLAG)
        // Initial state
        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])
        // Loaded state
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([true, 'loaded', undefined])

        // Rerender and wait for new state
        state.rerender({ overrides: {}, flagName: DISABLED_FLAG })
        await state.waitForNextUpdate()
        expect(state.result.current).toStrictEqual([false, 'loaded', undefined])
    })

    it('returns "error" when no context parent', () => {
        const state = renderHook(() => useFeatureFlag(ENABLED_FLAG))
        // Initial state
        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])
        // Loaded state
        expect(state.result.current).toEqual(expect.arrayContaining([false, 'error']))
    })

    it('returns "error" when unhandled error', async () => {
        const state = setup(ERROR_FLAG)
        await state.waitForNextUpdate()
        expect(state.result.current).toEqual(expect.arrayContaining([false, 'error']))
    })
})
