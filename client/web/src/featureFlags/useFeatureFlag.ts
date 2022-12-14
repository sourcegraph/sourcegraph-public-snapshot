import { useContext, useEffect, useState } from 'react'

import { logger } from '@sourcegraph/common'

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
            if (status !== 'error') {
                logger.warn(MISSING_CLIENT_ERROR)
                setResult(({ value }) => ({ value, status: 'error', error: new Error(MISSING_CLIENT_ERROR) }))
            }
            return
        }

        async function getValue(): Promise<void> {
            const newValue = await client!.get(flagName).toPromise()

            if (newValue === value && status !== 'initial') {
                return
            }

            if (isMounted) {
                setResult({ value: newValue, status: 'loaded' })
            }
        }

        getValue().catch(error => {
            if (isMounted) {
                setResult(({ value }) => ({ value, status: 'error', error }))
            }
        })

        return () => {
            isMounted = false
        }
    })

    return [typeof value === 'boolean' ? value : defaultValue, status, error]
}
