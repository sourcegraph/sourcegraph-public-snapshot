import { setExperimentalFeaturesFromSettings, useExperimentalFeatures } from './experimentalFeatures'

describe('experimentalFeatures store', () => {
    // NOTE: This test is not using '@testing-library/react-hooks' because using
    // 'renderHook' shows a warning in the test output about using the wrong
    // 'act' function (because our zustand mock uses a different 'act'
    // function). Since we only want to test the interaction between
    // 'setExperimentalFeaturesFromSettings' and the store, that's OK (we assume
    // that zustand itself works correctly)

    it('returns experimental feature flags', () => {
        setExperimentalFeaturesFromSettings({
            subjects: null,
            final: { experimentalFeatures: { fuzzyFinder: true, showSearchContext: false } },
        })

        expect(useExperimentalFeatures.getState()).toStrictEqual({ fuzzyFinder: true, showSearchContext: false })
    })
})
