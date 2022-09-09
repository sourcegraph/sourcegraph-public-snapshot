import { QueryGraphQLFn } from '../util/graphql'
import { parseGitURI } from '../util/uri'

/** The response envelope for all LSIF queries. */
export interface GenericLSIFResponse<R> {
    repository: { commit: { blob: { lsif: R | null } | null } | null }
}

/**
 * Perform an LSIF request to the GraphQL API.
 *
 * @param args Parameter bag.
 * @param queryGraphQL The function used to query the GraphQL API.
 */
export async function queryLSIF<P extends { query: string; uri: string }, R>(
    {
        /** The GraphQL request query. */
        query,
        /** The current text document uri. */
        uri,
        /** Additional query parameters. */
        ...rest
    }: P,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<R>>
): Promise<R | null> {
    const { repo, commit, path } = parseGitURI(new URL(uri))
    const queryArguments = { repository: repo, commit, path, ...rest }
    const data = await queryGraphQL(query, queryArguments)
    return data.repository?.commit?.blob?.lsif || null
}
