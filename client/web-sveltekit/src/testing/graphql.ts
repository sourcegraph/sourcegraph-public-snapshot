import { ApolloClient, InMemoryCache, type NormalizedCacheObject } from '@apollo/client/core'

export function createTestGraphqlClient(): ApolloClient<NormalizedCacheObject> {
    return new ApolloClient({
        cache: new InMemoryCache(),
    })
}
