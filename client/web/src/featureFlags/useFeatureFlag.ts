import { useContext, useMemo } from 'react'

import { of } from 'rxjs'

import { useObservableWithStatus } from '@sourcegraph/wildcard'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'loading' | 'finished' | 'error'

/**
 * Returns an evaluated feature flag for the current user
 *
 * @returns [flagValue, fetchStatus, error]
 */
export function useFeatureFlag(flagName: FeatureFlagName): [boolean, FetchStatus, any] {
    const { client } = useContext(FeatureFlagsContext)
    const [value = false, observableStatus, error] = useObservableWithStatus(
        useMemo(() => client?.get(flagName) ?? of(false), [client, flagName])
    )
    const status: FetchStatus = useMemo(
        () =>
            ['completed', 'next'].includes(observableStatus)
                ? 'finished'
                : observableStatus === 'error'
                ? 'error'
                : 'loading',
        [observableStatus]
    )

    return [value, status, error]
}
