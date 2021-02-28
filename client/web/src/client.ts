import { ApolloClient, InMemoryCache } from '@apollo/client'

export const client = new ApolloClient({
    uri: 'https:/sourcegraph.test:3443/.api/graphql',
    cache: new InMemoryCache(),
})
