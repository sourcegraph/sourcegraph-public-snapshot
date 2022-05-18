import { useContext, useEffect, useState } from 'react'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'loading' | 'finished' | 'error'

/**
 * Returns an evaluated feature flag for the current user
 */
export function useFeatureFlag(flagName: FeatureFlagName): [boolean, FetchStatus, Error | null] {
    const { client } = useContext(FeatureFlagsContext)
    const [value, setValue] = useState<boolean>(false)
    const [status, setStatus] = useState<FetchStatus>('loading')
    const [error, setError] = useState<Error | null>(null)

    useEffect(() => {
        let isMounted = true
        if (!client) {
            console.warn(
                '[useFeatureFlag]: No FeatureFlagClient is configured. All feature flags will default to "false" value.'
            )
            return
        }

        const cleanup = client.on(flagName, (value, error) => {
            if (!isMounted) {
                return
            }
            if (error) {
                setError(error)
                setStatus('error')
                return
            }
            setStatus('finished')
            setValue(value)
        })

        return () => {
            isMounted = false
            cleanup()
        }
    })
    return [value, status, error]
}
