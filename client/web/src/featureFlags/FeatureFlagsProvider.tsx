import React, { createContext, useEffect, useMemo } from 'react'

import { Observable, of } from 'rxjs'

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
    overrides: Record<FeatureFlagName, boolean>
    refetchInterval?: number
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
    refetchInterval,
    children,
}) => {
    const mockRequestGraphQL = useMemo(
        () => (
            query: string,
            variables: any
        ): Observable<{
            data: { evaluateFeatureFlag: boolean }
        }> =>
            of({
                data: { evaluateFeatureFlag: overrides[variables.flagName as FeatureFlagName] ?? false },
            }),
        [overrides]
    )

    const client = useMemo(
        () => new FeatureFlagClient(mockRequestGraphQL as typeof requestGraphQL, refetchInterval),
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    useEffect(() => {
        client.setRequestGraphQLFn(mockRequestGraphQL as typeof requestGraphQL)
    }, [client, mockRequestGraphQL])

    return <FeatureFlagsContext.Provider value={{ client }}>{children}</FeatureFlagsContext.Provider>
}
