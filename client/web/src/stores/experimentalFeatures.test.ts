import { setExperimentalFeaturesFromSettings, useExperimentalFeatures } from './experimentalFeatures'

describe('experimentalFeatures store', () => {
    // NOTE: This test is not using '@testing-library/react-hooks' because using
    // only want to test the interaction between
    // 'setExperimentalFeaturesFromSettings' and the store.

    it('returns experimental feature flags', () => {
        setExperimentalFeaturesFromSettings({
            subjects: null,
            final: { experimentalFeatures: { fuzzyFinder: true, showSearchContext: false } },
        })

        expect(useExperimentalFeatures.getState()).toHaveProperty('fuzzyFinder', true)
        expect(useExperimentalFeatures.getState()).toHaveProperty('showSearchContext', false)
    })
})
