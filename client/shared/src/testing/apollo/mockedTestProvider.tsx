import React, { useMemo } from 'react'

import { MockedProvider, type MockedProviderProps } from '@apollo/client/testing'

import { generateCache } from '../../backend/apolloCache'

export const MockedTestProvider: React.FunctionComponent<React.PropsWithChildren<MockedProviderProps>> = ({
    children,
    ...props
}) => {
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
