import React, { useEffect, useMemo } from 'react'

import { Observable, of, throwError } from 'rxjs'

import { requestGraphQL } from '../backend/graphql'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'
import { FeatureFlagClient } from './lib/FeatureFlagClient'

interface MockedFeatureFlagsProviderProps {
    overrides: Partial<Record<FeatureFlagName, boolean | Error>>
    cacheTimeToLive?: number
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
> = ({ overrides, cacheTimeToLive, children }) => {
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

    const providerValue = useMemo(
        () => ({ client: new FeatureFlagClient(mockRequestGraphQL as typeof requestGraphQL, cacheTimeToLive) }),
        // eslint-disable-next-line react-hooks/exhaustive-deps
        []
    )

    useEffect(() => {
        providerValue.client.setRequestGraphQLFunction(mockRequestGraphQL as typeof requestGraphQL)
    }, [providerValue, mockRequestGraphQL])

    return <FeatureFlagsContext.Provider value={providerValue}>{children}</FeatureFlagsContext.Provider>
}
