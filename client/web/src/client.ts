import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client'

const uri = 'https://sourcegraph.test:3443/.api/graphql'

export const client = new ApolloClient({
    uri,
    cache: new InMemoryCache(),
    link: createHttpLink({
        uri,
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
            'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
        },
    }),
    connectToDevTools: true,
})
