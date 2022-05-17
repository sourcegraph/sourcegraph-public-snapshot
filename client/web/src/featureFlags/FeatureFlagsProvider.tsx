import React, { createContext, useEffect, useRef } from 'react'

import { FeatureFlagClient } from './featureFlags'
import { getOverrideKey } from './lib/getOverrideKey'
import { parseUrlOverrideFeatureFlags } from './lib/parseUrlOverrideFeatureFlags'

export const FeatureFlagsContext = createContext<{ client?: FeatureFlagClient }>({})

interface FeatureFlagsProviderProps {
    isLocalOverrideEnabled?: boolean
}

/**
 * Overrides feature flag based on initial URL query parameters
 *
 * @description
 * Enable: "/?feature-flag-key=my-feature&feature-flag-value=true"
 * Disable: "/?feature-flag-key=my-feature&feature-flag-value=false"
 * Remove/reset local override: "/?feature-flag-key=my-feature"
 * Multiple values: /?feature-flag-key=my-feature-one,my-feature-two&feature-flag-value=false,true
 */
const FeatureFlagsLocalOverrideAgent = React.memo(() => {
    useEffect(() => {
        try {
            const overrideFeatureFlags = parseUrlOverrideFeatureFlags(location.search) || {}
            for (const [overrideFeatureKey, overrideFeatureValue] of Object.entries(overrideFeatureFlags)) {
                if (!overrideFeatureValue) {
                    localStorage.removeItem(getOverrideKey(overrideFeatureKey))
                } else {
                    localStorage.setItem(
                        getOverrideKey(overrideFeatureKey),
                        (overrideFeatureValue === 'true').toString()
                    )
                }
            }
        } catch (error) {
            console.error(error)
        }
    }, [])
    return null
})

export const FeatureFlagsProvider: React.FunctionComponent<FeatureFlagsProviderProps> = ({
    isLocalOverrideEnabled = true,
    children,
}) => {
    const clientReference = useRef(new FeatureFlagClient(true))

    return (
    <FeatureFlagsContext.Provider value={{ client: clientReference.current }}>
        {isLocalOverrideEnabled && <FeatureFlagsLocalOverrideAgent />}
        {children}
    </FeatureFlagsContext.Provider>
)}
