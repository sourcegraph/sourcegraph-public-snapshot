import React, { useMemo } from 'react'

import { ApolloLink } from '@apollo/client'
import { MockedProvider, type MockedProviderProps, type MockedResponse, MockLink } from '@apollo/client/testing'
import { getOperationName } from '@apollo/client/utilities'

import { logger } from '@sourcegraph/common'

import { generateCache } from '../backend/apolloCache'

/**
 * Intercept each mocked Apollo request and ensure that any request variables match the specified mock.
 * This effectively means we are mocking against the operationName of the query being fired.
 */
const forceMockVariablesLink = (mocks: readonly MockedResponse[]): ApolloLink =>
    new ApolloLink((operation, forward) => {
        const mock = mocks.find(mock => getOperationName(mock.request.query) === operation.operationName)
        if (mock) {
            operation.variables = mock.request.variables || {}
        } else {
            logger.warn(`Unable to find a mock for query: ${operation.operationName}. Did you mean to mock this?`)
        }
        return forward(operation)
    })

export interface MockedStoryProviderProps extends MockedProviderProps {
    /**
     * Set this to `true` to preserve the default behavior of MockedProvider.
     * Requests will require that both the `operationName` **and** `variables` match the mock to be resolved.
     */
    useStrictMocking?: boolean
}

/**
 * A wrapper around MockedProvider with a custom ApolloLink to ensure flexible request mocking.
 *
 * MockedProvider does not support dynamic variable matching for mocks.
 * This wrapper **only** mocks against the operation name, the specific provided variables are not used to match against a mock.
 */
export const MockedStoryProvider: React.FunctionComponent<React.PropsWithChildren<MockedStoryProviderProps>> = ({
    children,
    mocks = [],
    useStrictMocking,
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
            mocks={mocks}
            link={ApolloLink.from(
                useStrictMocking ? [new MockLink(mocks)] : [forceMockVariablesLink(mocks), new MockLink(mocks)]
            )}
            {...props}
        >
            {children}
        </MockedProvider>
    )
}
