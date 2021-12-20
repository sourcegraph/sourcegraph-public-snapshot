import React, { useEffect } from 'react'

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
                localStorage.removeItem(featureFlagKey)
            } else {
                localStorage.setItem(featureFlagKey, Boolean(featureFlagValue).toString())
            }
        } catch (error) {
            console.error(error)
        }
    }, [])
    return null
})
