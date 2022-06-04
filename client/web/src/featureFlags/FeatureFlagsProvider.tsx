import React, { createContext, useEffect, useMemo } from 'react'

import { Observable, of, throwError } from 'rxjs'

import { requestGraphQL } from '../backend/graphql'

import { FeatureFlagName } from './featureFlags'
import { removeFeatureFlagOverride, setFeatureFlagOverride } from './lib/feature-flag-local-overrides'
import { FeatureFlagClient, IFeatureFlagClient } from './lib/FeatureFlagClient'
import { parseUrlOverrideFeatureFlags } from './lib/parseUrlOverrideFeatureFlags'

export const FeatureFlagsContext = createContext<{ client?: IFeatureFlagClient }>({})

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
            for (const [flagName, value] of Object.entries(overrideFeatureFlags)) {
                if (!value) {
                    removeFeatureFlagOverride(flagName)
                } else if ([1, 'true'].includes(value)) {
                    setFeatureFlagOverride(flagName, true)
                } else if ([0, 'false'].includes(value)) {
                    setFeatureFlagOverride(flagName, false)
                } else {
                    console.warn(
                        `[FeatureFlagsLocalOverrideAgent]: can not override feature flag "${flagName}" with value "${value}". Only boolean values are supported.`
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
    const client = useMemo(() => new FeatureFlagClient(requestGraphQL), [])

    return (
        <FeatureFlagsContext.Provider value={{ client }}>
            {isLocalOverrideEnabled && <FeatureFlagsLocalOverrideAgent />}
            {children}
        </FeatureFlagsContext.Provider>
    )
}

interface MockedFeatureFlagsProviderProps {
    overrides: Map<FeatureFlagName, boolean | Error>
}

/**
 * Provides mocked feature flag value for testing purposes.
 *
 * @example
 * const overrides = new Map([['my-feature-flag', true]]);
 * return (<MockedFeatureFlagsProvider overrides={overrides}>
 *              <ComponentUsingFeatureFlag />
 *         </MockedFeatureFlagsProvider>)
 */
export const MockedFeatureFlagsProvider: React.FunctionComponent<MockedFeatureFlagsProviderProps> = ({
    overrides,
    children,
}) => {
    const client = useMemo(() => new MockFeatureFlagClient(overrides), [overrides])
    return <FeatureFlagsContext.Provider value={{ client }}>{children}</FeatureFlagsContext.Provider>
}

class MockFeatureFlagClient implements IFeatureFlagClient {
    constructor(private overrides: Map<FeatureFlagName, boolean | Error>) {}

    public get(flagName: FeatureFlagName): Observable<boolean> {
        const value = this.overrides.get(flagName)
        return value instanceof Error ? throwError(value) : of(value || false)
    }
}
