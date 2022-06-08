import { useContext, useMemo } from 'react'

import { throwError } from 'rxjs'

import { useObservableWithStatus } from '@sourcegraph/wildcard'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'initial' | 'loaded' | 'error'

/**
 * Returns an evaluated feature flag for the current user
 *
 * @returns [flagValue, fetchStatus, error]
 */
export function useFeatureFlag(flagName: FeatureFlagName): [boolean, FetchStatus, any] {
    const { client } = useContext(FeatureFlagsContext)
    const [value = false, observableStatus, error] = useObservableWithStatus(
        useMemo(() => client?.get(flagName) ?? throwError(new Error('No FeatureFlagClient set in context')), [
            client,
            flagName,
        ])
    )

    const status: FetchStatus = useMemo(() => {
        if (['completed', 'next'].includes(observableStatus)) {
            return 'loaded'
        }
        if (observableStatus === 'error') {
            return 'error'
        }
        return 'initial'
    }, [observableStatus])

    return [value, status, error]
}
