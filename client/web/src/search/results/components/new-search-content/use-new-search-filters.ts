import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'

export function useIsNewSearchFiltersEnabled(): boolean {
    const newFiltersEnabled = useExperimentalFeatures(features => features.newSearchResultFiltersPanel)

    return newFiltersEnabled ?? true
}
