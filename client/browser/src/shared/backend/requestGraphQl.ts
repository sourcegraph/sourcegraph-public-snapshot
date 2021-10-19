import { print } from 'graphql'
import { once } from 'lodash'
import { from, Observable } from 'rxjs'

import {
    GraphQLResult,
    getGraphQLClient,
    GraphQLClient,
    requestGraphQLCommon,
} from '@sourcegraph/shared/src/graphql/graphql'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { background } from '../../browser-extension/web-extension-api/runtime'
import { isBackground } from '../context'

import { getHeaders } from './headers'

interface RequestGraphQLOptions<V> {
    request: string
    variables: V
    mightContainPrivateInfo: boolean
    privateCloudErrors?: Observable<boolean>
}

interface GraphQLHelpers {
    getBrowserGraphQLClient: PlatformContext['getGraphQLClient']
    requestGraphQL: <T, V = object>(options: RequestGraphQLOptions<V>) => Observable<GraphQLResult<T>>
}

function createMainThreadExtensionGraphQLHelpers(sourcegraphURL: string): GraphQLHelpers {
    /**
     * Forward GraphQL request to the background script for execution.
     */
    const requestGraphQLInBackground = <T, V = object>(
        options: RequestGraphQLOptions<V>
        // Keep both helpers inside of the factory function.
        // eslint-disable-next-line unicorn/consistent-function-scoping
    ): Observable<GraphQLResult<T>> =>
        from(
            background.requestGraphQL<T, V>({ ...options, sourcegraphURL })
        )

    /**
     * Apollo-Client is not configured yet to execute requests in the background script.
     * Fallback to `requestGraphQLInBackground` when `client.watchQuery` is called.
     *
     * The implementation should forward Apollo-Client method calls to the background script
     * in the same manner as it's done for `requestGraphQL` method. Or we can consider executing
     * API requests in the main thread.
     */
    const getBrowserGraphQLClient = once(() => {
        if (process.env.NODE_ENV === 'development') {
            console.warn(
                'Apollo-Client mock is used in browser extension to forward GraphQL requests to the background script!'
            )
        }

        const graphqlClient: Pick<GraphQLClient, 'watchQuery'> = {
            watchQuery: ({ variables, query }) =>
                // Temporary implementation till Apollo-Client is configured in the background script.
                // eslint-disable-next-line @typescript-eslint/no-unsafe-return
                requestGraphQLInBackground({
                    request: print(query),
                    variables,
                    mightContainPrivateInfo: false,
                }) as any,
        }

        return Promise.resolve(graphqlClient) as Promise<GraphQLClient>
    })

    return { getBrowserGraphQLClient, requestGraphQL: requestGraphQLInBackground }
}

/**
 * Returns a platform-appropriate implementation of the function used to make requests to our GraphQL API.
 *
 * In the browser extension, the returned function will make all requests from the background page.
 * In the native integration, the returned function will rely on the `requestGraphQL` implementation from `/shared`.
 */
export function createGraphQLHelpers(sourcegraphURL: string, isExtension: boolean): GraphQLHelpers {
    if (isExtension && !isBackground) {
        if (process.env.NODE_ENV === 'development') {
            console.warn('GraphQL requests initiated in the main thread are forwarded to the background script!')
            console.warn('Check out the implementation of the `requestGraphQLInBackground` function above.')
        }

        return createMainThreadExtensionGraphQLHelpers(sourcegraphURL)
    }

    const requestGraphQL = <T, V = object>({
        request,
        variables,
    }: RequestGraphQLOptions<V>): Observable<GraphQLResult<T>> =>
        requestGraphQLCommon<T, V>({
            request,
            variables,
            baseUrl: sourcegraphURL,
            credentials: 'include',
        })
    /**
     * Memoized Apollo Client getter. It should be executed once to restore the cache from the local storage.
     * After that, the same instance should be used by all consumers.
     */
    const getBrowserGraphQLClient = once(() =>
        getGraphQLClient({ headers: getHeaders(), baseUrl: sourcegraphURL, isAuthenticated: false })
    )

    return { getBrowserGraphQLClient, requestGraphQL }
}
