import { type FC, type PropsWithChildren, useEffect } from 'react'

import { logger } from '@sourcegraph/common'

import { updateOverrideCounter } from '../stores'

import { removeFeatureFlagOverride, setFeatureFlagOverride } from './lib/feature-flag-local-overrides'
import { parseUrlOverrideFeatureFlags } from './lib/parseUrlOverrideFeatureFlags'

/**
 * Overrides feature flag based on initial URL query parameters
 * @description
 * Enable: "/?feat=my-feature"
 * Disable: "/?feat=-my-feature"
 * Remove/reset: "/?feat=~my-feature"
 * Multiple values: "/?feat=my-feature1,-my-feature2"
 */
export const FeatureFlagsLocalOverrideAgent: FC<PropsWithChildren<{}>> = ({ children }) => {
    useEffect(() => {
        try {
            const overrideFeatureFlags = parseUrlOverrideFeatureFlags(location.search)
            for (const [flagName, value] of overrideFeatureFlags) {
                if (value !== null) {
                    setFeatureFlagOverride(flagName, value)
                } else {
                    removeFeatureFlagOverride(flagName)
                }
            }
            // Update override counter to notify/update the developer settings
            // dialog.
            updateOverrideCounter()
        } catch (error) {
            logger.error(error)
        }
    }, [])

    return <>{children}</>
}
