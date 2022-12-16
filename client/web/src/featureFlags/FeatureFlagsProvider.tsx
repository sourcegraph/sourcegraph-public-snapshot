import React, { createContext, useEffect } from 'react'

import { logger } from '@sourcegraph/common'

import { requestGraphQL } from '../backend/graphql'

import { removeFeatureFlagOverride, setFeatureFlagOverride } from './lib/feature-flag-local-overrides'
import { FeatureFlagClient } from './lib/FeatureFlagClient'
import { parseUrlOverrideFeatureFlags } from './lib/parseUrlOverrideFeatureFlags'

interface FeatureFlagsContextValue {
    client?: FeatureFlagClient
}

export const FeatureFlagsContext = createContext<FeatureFlagsContextValue>({})

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
            for (const [flagName, value] of Object.entries(overrideFeatureFlags)) {
                if (!value) {
                    removeFeatureFlagOverride(flagName)
                } else if ([1, 'true'].includes(value)) {
                    setFeatureFlagOverride(flagName, true)
                } else if ([0, 'false'].includes(value)) {
                    setFeatureFlagOverride(flagName, false)
                } else {
                    logger.warn(
                        `[FeatureFlagsLocalOverrideAgent]: can not override feature flag "${flagName}" with value "${value}". Only boolean values are supported.`
                    )
                }
            }
        } catch (error) {
            logger.error(error)
        }
    }, [])
    return null
})

const MINUTE = 60000

const featureFlagsContextValue = {
    client: new FeatureFlagClient(requestGraphQL, MINUTE * 10),
} satisfies FeatureFlagsContextValue

interface FeatureFlagsProviderProps {
    isLocalOverrideEnabled?: boolean
}

export const FeatureFlagsProvider: React.FunctionComponent<React.PropsWithChildren<FeatureFlagsProviderProps>> = ({
    isLocalOverrideEnabled = true,
    children,
}) => (
    <FeatureFlagsContext.Provider value={featureFlagsContextValue}>
        {isLocalOverrideEnabled && <FeatureFlagsLocalOverrideAgent />}
        {children}
    </FeatureFlagsContext.Provider>
)
