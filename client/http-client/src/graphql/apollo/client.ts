import {
    ApolloClient,
    createHttpLink,
    from,
    type HttpOptions,
    type InMemoryCache,
    type NormalizedCacheObject,
} from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import type { PersistenceMapperFunction } from 'apollo3-cache-persist/lib/types'
import { once } from 'lodash'

import { defined } from '@sourcegraph/common'

import { checkOk } from '../../http-status-error'
import { buildGraphQLUrl } from '../graphql'
import { ConcurrentRequestsLink } from '../links/concurrent-requests-link'

interface GetGraphqlClientOptions {
    getHeaders?: () => Record<string, string>
    cache: InMemoryCache
    baseUrl?: string
    credentials?: 'include' | 'omit' | 'same-origin'
    persistenceMapper?: PersistenceMapperFunction
}

export type GraphQLClient = ApolloClient<NormalizedCacheObject>

export const getGraphQLClient = once(async (options: GetGraphqlClientOptions): Promise<GraphQLClient> => {
    const { getHeaders, baseUrl, credentials, cache } = options
    const uri = buildGraphQLUrl({ baseUrl })

    const apolloClient = new ApolloClient({
        uri,
        cache,
        defaultOptions: {
            /**
             * The default `fetchPolicy` is `cache-first`, which returns a cached response
             * and doesn't trigger cache update. This is undesirable default behavior because
             * we want to keep our cache updated to avoid confusing the user with stale data.
             * `cache-and-network` allows us to return a cached result right away and then update
             * all consumers with the fresh data from the network request.
             */
            watchQuery: {
                fetchPolicy: 'cache-and-network',
            },
            /**
             * `client.query()` returns promise, so it can only resolve one response.
             * Meaning we cannot return the cached result first and then update it with
             * the response from the network as it's done in `client.watchQuery()`.
             * So we always need to make a network request to get data unless another
             * `fetchPolicy` is specified in the `client.query()` call.
             */
            query: {
                fetchPolicy: 'network-only',
            },
        },
        link: from(
            defined([
                new ConcurrentRequestsLink(),
                getHeaders
                    ? setContext((_request, previousContext) => ({
                          ...previousContext,
                          headers: {
                              ...previousContext.headers,
                              ...getHeaders(),
                          },
                      }))
                    : null,
                createHttpLink({
                    uri: ({ operationName }) => `${uri}?${operationName}`,
                    credentials,
                    fetch: customFetch,
                }),
            ])
        ),
    })

    return Promise.resolve(apolloClient)
})

const customFetch: HttpOptions['fetch'] = (uri, options) => fetch(uri, options).then(checkOk)
