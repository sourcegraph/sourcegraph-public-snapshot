import { GraphQLParams } from 'graphiql'
import { GraphQLClient } from 'graphql-request'
import { GraphQLResponse, Variables } from 'graphql-request/dist/types'
import { QueryFunction } from 'react-query'

const uri = 'https://sourcegraph.test:3443/.api/graphql'

// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
const fetcher: QueryFunction = async (context: { queryKey: [string, Variables] }) => {
    const [query, variables] = context.queryKey

    const graphQLClient = new GraphQLClient(uri, {
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
            'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
        },
    })

    const data = await graphQLClient.request<GraphQLResponse>(query, variables)
    return data
}

export default fetcher
