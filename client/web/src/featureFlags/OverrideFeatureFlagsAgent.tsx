import React, { useEffect } from 'react'

import { getOverrideKey } from './lib/getOverrideKey'
import { parseUrlOverrideFeatureFlags } from './lib/parseUrlOverrideFeatureFlags'

/**
 * Overrides feature flag based on initial URL query parameters
 *
 * @description
 * Enable: "/?feature-flag-key=my-feature&feature-flag-value=true"
 * Disable: "/?feature-flag-key=my-feature&feature-flag-value=false"
 * Remove/reset local override: "/?feature-flag-key=my-feature"
 * Multiple values: /?feature-flag-key=my-feature-one,my-feature-two&feature-flag-value=false,true
 */
export const OverrideFeatureFlagsAgent = React.memo(() => {
    useEffect(() => {
        try {
            const overrideFeatureFlags = parseUrlOverrideFeatureFlags(location.search) || {}
            for (const [overrideFeatureKey, overrideFeatureValue] of Object.entries(overrideFeatureFlags)) {
                if (!overrideFeatureValue) {
                    localStorage.removeItem(getOverrideKey(overrideFeatureKey))
                } else {
                    localStorage.setItem(
                        getOverrideKey(overrideFeatureKey),
                        Boolean(overrideFeatureValue === 'true').toString()
                    )
                }
            }
        } catch (error) {
            console.error(error)
        }
    }, [])
    return null
})
