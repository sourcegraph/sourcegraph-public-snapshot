import React, { createContext, useEffect, useMemo } from 'react'

import { Observable, of, throwError } from 'rxjs'

import { logger } from '@sourcegraph/common'

import { requestGraphQL } from '../backend/graphql'

import { FeatureFlagName } from './featureFlags'
import { removeFeatureFlagOverride, setFeatureFlagOverride } from './lib/feature-flag-local-overrides'
import { FeatureFlagClient } from './lib/FeatureFlagClient'
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
export const FeatureFlagsProvider: React.FunctionComponent<React.PropsWithChildren<FeatureFlagsProviderProps>> = ({
    isLocalOverrideEnabled = true,
    children,
}) => {
    const client = useMemo(() => new FeatureFlagClient(requestGraphQL, MINUTE), [])

    return (
        <FeatureFlagsContext.Provider value={{ client }}>
            {isLocalOverrideEnabled && <FeatureFlagsLocalOverrideAgent />}
            {children}
        </FeatureFlagsContext.Provider>
    )
}

interface MockedFeatureFlagsProviderProps {
    overrides: Partial<Record<FeatureFlagName, boolean | Error>>
    refetchInterval?: number
}

/**
 * Provides mocked feature flag value for testing purposes.
 *
 * @example
 * return (<MockedFeatureFlagsProvider overrides={{'my-feature-flag': true}}>
 *              <ComponentUsingFeatureFlag />
 *         </MockedFeatureFlagsProvider>)
 */
export const MockedFeatureFlagsProvider: React.FunctionComponent<
    React.PropsWithChildren<MockedFeatureFlagsProviderProps>
> = ({ overrides, refetchInterval, children }) => {
    const mockRequestGraphQL = useMemo(
        () =>
            (
                query: string,
                variables: any
            ): Observable<{
                data: { evaluateFeatureFlag: boolean | null }
            }> => {
                const value = overrides[variables.flagName as FeatureFlagName]
                if (value instanceof Error) {
                    return throwError(value)
                }

                return of({
                    data: { evaluateFeatureFlag: value ?? null },
                })
            },
        [overrides]
    )

    const client = useMemo(
        () => new FeatureFlagClient(mockRequestGraphQL as typeof requestGraphQL, refetchInterval),
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    useEffect(() => {
        client.setRequestGraphQLFunction(mockRequestGraphQL as typeof requestGraphQL)
    }, [client, mockRequestGraphQL])

    return <FeatureFlagsContext.Provider value={{ client }}>{children}</FeatureFlagsContext.Provider>
}
