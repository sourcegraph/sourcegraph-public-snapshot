import { MockedProvider, MockedProviderProps } from '@apollo/client/testing'
import { act } from '@testing-library/react'
import React, { useMemo } from 'react'

import { generateCache } from '../graphql/cache'

/*
 * Wait one tick to load the next response from Apollo
 * https://www.apollographql.com/docs/react/development-testing/testing/#testing-the-success-state
 */
export const waitForNextApolloResponse = (): Promise<void> => act(() => new Promise(resolve => setTimeout(resolve, 0)))

export const MockedTestProvider: React.FunctionComponent<MockedProviderProps> = ({ children, ...props }) => {
    /**
     * Generate a fresh cache for each instance of MockedTestProvider.
     * Important to ensure tests don't share cached data.
     */
    const cache = useMemo(() => generateCache(), [])

    return (
        <MockedProvider
            cache={cache}
            defaultOptions={{
                mutate: {
                    // Fix errors being thrown globally https://github.com/apollographql/apollo-client/issues/7167
                    errorPolicy: 'all',
                },
            }}
            {...props}
        >
            {children}
        </MockedProvider>
    )
}
