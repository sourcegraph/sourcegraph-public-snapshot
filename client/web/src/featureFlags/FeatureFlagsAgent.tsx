import React, { useEffect } from 'react'

import { getOverrideKey } from './lib/getOverrideKey'

/**
 * Overrides feature flag based on initial URL query parameters
 *
 * @description
 * To force enabled: "/?feature-flag-key=my-feature&feature-flag-value=true"
 * To force disable: "/?feature-flag-key=my-feature&feature-flag-value=false"
 * To remove local override: "/?feature-flag-key=my-feature"
 */
export const FeatureFlagsAgent = React.memo(() => {
    useEffect(() => {
        try {
            const queryString = location.search
            const urlParameters = new URLSearchParams(queryString)
            const featureFlagKey = urlParameters.get('feature-flag-key')
            if (!featureFlagKey) {
                return
            }
            const featureFlagValue = urlParameters.get('feature-flag-value')
            if (!featureFlagValue) {
                localStorage.removeItem(getOverrideKey(featureFlagKey))
            } else {
                localStorage.setItem(getOverrideKey(featureFlagKey), Boolean(featureFlagValue === 'true').toString())
            }
        } catch (error) {
            console.error(error)
        }
    }, [])
    return null
})
