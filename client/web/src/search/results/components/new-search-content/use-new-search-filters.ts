import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'

import { useFeatureFlag } from '../../../../featureFlags/useFeatureFlag'

export function useIsNewSearchFiltersEnabled(): boolean {
    const newFiltersEnabled = useExperimentalFeatures(features => features.newSearchResultFiltersPanel)
    const [newFiltersFeatureFlagEnabled] = useFeatureFlag('search.newFilters', false)

    return newFiltersEnabled || newFiltersFeatureFlagEnabled
}
