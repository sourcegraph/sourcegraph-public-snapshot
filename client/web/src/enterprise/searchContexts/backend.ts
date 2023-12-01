import { ApolloQueryResult, gql, ApolloClient } from '@apollo/client'

import type { InputMaybe, RepositoriesByNamesResult, RepositoriesByNamesVariables } from '../../graphql-operations'

const query = gql`
    query RepositoriesByNames($names: [String!]!, $first: Int!, $after: String) {
        repositories(names: $names, first: $first, after: $after) {
            pageInfo {
                hasNextPage
                endCursor
            }
            nodes {
                id
                name
            }
        }
    }
`
export async function fetchRepositoriesByNames(
    names: string[],
    apolloClient: ApolloClient<object>
): Promise<RepositoriesByNamesResult['repositories']['nodes']> {
    let repos: RepositoriesByNamesResult['repositories']['nodes'] = []
    const first = names.length
    let after: InputMaybe<string> = null

    while (true) {
        const result: ApolloQueryResult<RepositoriesByNamesResult> = await apolloClient.query<
            RepositoriesByNamesResult,
            RepositoriesByNamesVariables
        >({
            query,
            variables: {
                names,
                first,
                after,
            },
        })

        if (result.error) {
            throw new Error(result.error.message)
        }

        repos = repos.concat(result.data.repositories.nodes)
        if (!result.data.repositories.pageInfo.hasNextPage) {
            break
        }
        after = result.data.repositories.pageInfo.endCursor
    }
    return repos
}
