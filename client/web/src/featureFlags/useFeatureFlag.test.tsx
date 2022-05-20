import { renderHook, WrapperComponent } from '@testing-library/react-hooks'

import { FeatureFlagName } from './featureFlags'
import { MockedFeatureFlagsProvider } from './FeatureFlagsProvider'
import { useFeatureFlag } from './useFeatureFlag'

describe('useFeatureFlag', () => {
    it('returns state correctly', () => {
        const FLAG_NAME = 'test-flag'
        const wrapper: WrapperComponent<Record<string, boolean>> = ({ children, ...overrides }) => (
            <MockedFeatureFlagsProvider overrides={new Map(Object.entries(overrides) as [FeatureFlagName, boolean][])}>
                {children}
            </MockedFeatureFlagsProvider>
        )
        const { result, rerender } = renderHook(() => useFeatureFlag(FLAG_NAME as FeatureFlagName), {
            wrapper,
            initialProps: {
                [FLAG_NAME]: false,
            },
        })

        expect(result.all[0]).toStrictEqual([false, 'loading', null])
        expect(result.current).toStrictEqual([false, 'finished', null])
        rerender({ [FLAG_NAME]: true })
        expect(result.current).toStrictEqual([true, 'finished', null])
    })
})
