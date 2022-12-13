import React, { useEffect, useMemo } from 'react'

import { Observable, of, throwError } from 'rxjs'

import { requestGraphQL } from '../backend/graphql'

import { FeatureFlagName } from './featureFlags'
import { FeatureFlagsContext } from './FeatureFlagsProvider'
import { FeatureFlagClient } from './lib/FeatureFlagClient'

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
        () => (
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
