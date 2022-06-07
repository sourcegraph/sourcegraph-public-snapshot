import React from 'react'

import { FeatureFlagName } from './featureFlags'
import { useFeatureFlag } from './useFeatureFlag'

/**
 * HOC. Renders component when a certain feature flag value equals to a desired value.
 */
export const withFeatureFlag = <P extends object>(
    flagName: FeatureFlagName,
    TrueComponent: React.ComponentType<React.PropsWithChildren<P>>,
    FalseComponent?: React.ComponentType<React.PropsWithChildren<P>> | null
) =>
    function WithFeatureFlagHOC(props: P): React.ReactElement | null {
        const [value, status] = useFeatureFlag(flagName)

        if (status !== 'loaded') {
            return null
        }

        if (value === true) {
            return <TrueComponent {...props} />
        }

        return FalseComponent ? <FalseComponent {...props} /> : null
    }
