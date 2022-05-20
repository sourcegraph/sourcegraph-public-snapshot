import React from 'react'

import { FeatureFlagName } from './featureFlags'
import { useFeatureFlag } from './useFeatureFlag'

/**
 * HOC. Renders component when a certain feature flag value equals to a desired value.
 */
export const withFeatureFlag = <P extends object>(
    Component: React.ComponentType<React.PropsWithChildren<P>>,
    flagName: FeatureFlagName,
    flagValue: boolean = true
) => (props: P): React.ReactElement | null => {
    const [value] = useFeatureFlag(flagName)
    if (value !== flagValue) {
        return null
    }
    return <Component {...props} />
}
