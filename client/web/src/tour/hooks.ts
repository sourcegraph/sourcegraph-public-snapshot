import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

import type { AuthenticatedUser } from '../auth'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { useRecentSearches } from '../search/input/useRecentSearches'

export function useShowOnboardingTour({
    authenticatedUser,
    isSourcegraphDotCom,
}: {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}): boolean {
    const [enduserOnboardingEnabled] = useFeatureFlag('end-user-onboarding', false)
    return enduserOnboardingEnabled && !!authenticatedUser && !isSourcegraphDotCom
}

// If a user has made less than this number of queries we will show them
// the tour setup questions.
const MIN_QUERIES = 5

export function useShowOnboardingSetup(): boolean {
    // The user onboarding setup is only shown if the user hasn't
    // completed or skipped the process and has performed less than
    // MIN_QUERIES search queries.
    // While the data is loaded this function returns false.
    const [config, , configStatus] = useTemporarySetting('onboarding.userconfig')
    const { recentSearches, state: searchesStatus } = useRecentSearches()
    return (
        configStatus === 'loaded' &&
        searchesStatus === 'success' &&
        !!recentSearches &&
        recentSearches.length < MIN_QUERIES &&
        !(config?.skipped || config?.userinfo)
    )
}
