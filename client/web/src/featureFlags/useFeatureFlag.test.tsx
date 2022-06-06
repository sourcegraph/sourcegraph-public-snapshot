import { renderHook } from '@testing-library/react-hooks'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { useFeatureFlag } from './useFeatureFlag'

describe('useFeatureFlag', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName
    const setup = (initialFlagName: FeatureFlagName) =>
        renderHook(({ flagName }) => useFeatureFlag(flagName), {
            wrapper: function Wrapper({ children, overrides }) {
                return (
                    <MockedFeatureFlagsProvider
                        overrides={
                            new Map(
                                Object.entries({ [ENABLED_FLAG]: true, ...overrides }) as [FeatureFlagName, boolean][]
                            )
                        }
                    >
                        {children}
                    </MockedFeatureFlagsProvider>
                )
            },
            initialProps: {
                flagName: initialFlagName,
                overrides: {
                    [DISABLED_FLAG]: false,
                },
            },
        })

    it('returns [false] value correctly', () => {
        const state = setup(DISABLED_FLAG)

        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])

        expect(state.result.current).toStrictEqual([false, 'loaded', undefined])

        expect(state.result.all.length).toBe(2)
    })

    it('returns [true] value correctly', () => {
        const state = setup(ENABLED_FLAG)

        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])

        expect(state.result.current).toStrictEqual([true, 'loaded', undefined])

        expect(state.result.all.length).toBe(2)
    })

    it('updates when feature flag prop changes', () => {
        const state = setup(ENABLED_FLAG)

        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])
        expect(state.result.current).toStrictEqual([true, 'loaded', undefined])

        state.rerender({ overrides: {}, flagName: DISABLED_FLAG })
        expect(state.result.current).toStrictEqual([false, 'loaded', undefined])
    })

    it('returns "error" when no context parent', () => {
        const state = renderHook(() => useFeatureFlag(ENABLED_FLAG))

        expect(state.result.all[0]).toStrictEqual([false, 'initial', undefined])

        expect(state.result.current).toEqual(expect.arrayContaining([false, 'error']))

        expect(state.result.all.length).toBe(2)
    })
})
