import { useContext, useEffect, useState } from 'react'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'

type FetchStatus = 'loading' | 'finished' | 'error'

export function useFeatureFlag(flagName: FeatureFlagName): [boolean, FetchStatus, Error | null] {
    const { client } = useContext(FeatureFlagsContext)
    const [value, setValue] = useState<boolean>(false)
    const [status, setStatus] = useState<FetchStatus>('loading')
    const [error, setError] = useState<Error | null>(null)

    useEffect(() => {
        if (!client) {
            console.warn('No FeatureFlagClient not configured. All feature flags will default to "false" value.')
            return
        }
        return client.on(flagName, (value, error) => {
            if (error) {
                setError(error)
                setStatus('error')
                return
            }
            setStatus('finished')
            setValue(value)
        })
    })
    return [value, status, error]
}
