import { useContext, useEffect, useState } from 'react'

import { logger } from '@sourcegraph/common'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'initial' | 'loaded' | 'error'

/**
 * Returns an evaluated feature flag for the current user
 *
 * @returns [flagValue, fetchStatus, error]
 */
export function useFeatureFlag(flagName: FeatureFlagName, defaultValue = false): [boolean, FetchStatus, any?] {
    const { client } = useContext(FeatureFlagsContext)
    const [{ value, status, error }, setResult] = useState<{ value: boolean | null; status: FetchStatus; error?: any }>(
        {
            status: 'initial',
            value: defaultValue,
        }
    )

    useEffect(() => {
        let isMounted = true
        if (!client) {
            const errorMessage =
                '[useFeatureFlag]: No FeatureFlagClient is configured. All feature flags will default to "false" value.'
            logger.warn(errorMessage)
            setResult(({ value }) => ({ value, status: 'error', error: new Error(errorMessage) }))
            return
        }

        const subscription = client.get(flagName).subscribe(
            value => {
                if (!isMounted) {
                    return
                }
                setResult({ value, status: 'loaded' })
            },
            error => {
                if (!isMounted) {
                    return
                }
                setResult(({ value }) => ({ value, status: 'error', error }))
            }
        )

        return () => {
            isMounted = false
            setResult(({ value }) => ({ value, status: 'initial' }))
            subscription.unsubscribe()
        }
    }, [client, flagName])

    return [typeof value === 'boolean' ? value : defaultValue, status, error]
}
