import { useContext, useEffect, useState } from 'react'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'initial' | 'loaded' | 'error'

/**
 * Returns an evaluated feature flag for the current user
 */
export function useFeatureFlag(flagName: FeatureFlagName): [boolean, FetchStatus, any?] {
    const { client } = useContext(FeatureFlagsContext)
    const [{ value, status, error }, setResult] = useState<{ value: boolean; status: FetchStatus; error?: any }>({
        status: 'initial',
        value: false,
    })

    useEffect(() => {
        let isMounted = true
        if (!client) {
            const errorMessage =
                '[useFeatureFlag]: No FeatureFlagClient is configured. All feature flags will default to "false" value.'
            console.warn(errorMessage)
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

    return [value, status, error]
}
