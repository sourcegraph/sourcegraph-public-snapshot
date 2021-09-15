/* eslint-disable unicorn/filename-case */
import { ObservableQuery } from '@apollo/client'
import { mock } from 'jest-mock-extended'
import { of, Subscriber } from 'rxjs'

import { GraphQLClient } from '../../graphql/graphql'
import { PlatformContext } from '../../platform/context'

interface CreateGraphQLClientGetterOptions {
    /** Responses emitted by watch query sequentially for each call. */
    watchQueryMocks: object[]
}

/**
 * Helper to mock Apollo-Client with subsequent `watchQuery` calls.
 * It would be possible to use `import { createMockClient } from '@apollo/client/testing'`
 * but it requires a lot of monkey patching because there's no straightforward way to mock
 * `watchQuery` results for the returned client mock.
 */
export function createGraphQLClientGetter({
    watchQueryMocks,
}: CreateGraphQLClientGetterOptions): PlatformContext['getGraphQLClient'] {
    const observableQuery = mock<ObservableQuery<unknown, unknown>>()
    const graphQlClient = mock<GraphQLClient>()

    graphQlClient.watchQuery.mockReturnValue(observableQuery)

    for (const mockResponse of watchQueryMocks) {
        observableQuery.subscribe.mockImplementationOnce((subscriber: unknown) =>
            of(mockResponse).subscribe(subscriber as Subscriber<unknown>)
        )
    }

    return () => Promise.resolve(graphQlClient)
}
