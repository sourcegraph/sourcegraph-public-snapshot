import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { createInvalidGraphQLQueryResponseError, dataOrThrowErrors, gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'

/**
 * Fetches commits.
 */
export function fetchCommits(
    repo: GQL.ID,
    rev: string,
    args: { first?: number; currentPath?: string; query?: string }
): Observable<GQL.IGitCommitConnection> {
    return queryGraphQL(
        gql`
            query FetchCommits($repo: ID!, $rev: String!, $first: Int, $currentPath: String, $query: String) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $rev) {
                            ancestors(first: $first, query: $query, path: $currentPath) {
                                nodes {
                                    oid
                                    abbreviatedOID
                                    message
                                    author {
                                        person {
                                            avatarURL
                                            name
                                            email
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        { ...args, repo, rev }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node || !(data.node as GQL.IRepository).commit) {
                throw createInvalidGraphQLQueryResponseError('FetchCommits')
            }
            return (data.node as GQL.IRepository).commit!.ancestors
        })
    )
}
