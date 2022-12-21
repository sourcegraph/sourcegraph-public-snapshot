import { useContext, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { useIsMounted } from '@sourcegraph/wildcard'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'initial' | 'loaded' | 'error'
const MISSING_CLIENT_ERROR =
    '[useFeatureFlag]: No FeatureFlagClient is configured. All feature flags will default to "false" value.'

/**
 * Returns an evaluated feature flag for the current user
 *
 * @returns [flagValue, fetchStatus, error]
 */
export function useFeatureFlag(flagName: FeatureFlagName, defaultValue = false): [boolean, FetchStatus, any?] {
    const isMounted = useIsMounted()
    const { client } = useContext(FeatureFlagsContext)
    const [{ value, status, error }, setResult] = useState<{ value: boolean | null; status: FetchStatus; error?: any }>(
        {
            status: 'initial',
            value: defaultValue,
        }
    )

    if (!client) {
        if (status !== 'error') {
            logger.warn(MISSING_CLIENT_ERROR)
            setResult(({ value }) => ({ value, status: 'error', error: new Error(MISSING_CLIENT_ERROR) }))
        }
    } else {
        // We want to `client.get(flagName)` on every render and update the state only
        // on the value change. We won't be sending an API request on every render
        // because evaluated feature flags are cached in memory for a short period of time.
        async function getValue(): Promise<void> {
            const newValue = await client!.get(flagName)

            if (newValue === value && status !== 'initial') {
                return
            }

            if (isMounted()) {
                setResult({ value: newValue, status: 'loaded' })
            }
        }

        getValue().catch(error => {
            if (isMounted()) {
                setResult(({ value }) => ({ value, status: 'error', error }))
            }
        })
    }

    return [typeof value === 'boolean' ? value : defaultValue, status, error]
}
