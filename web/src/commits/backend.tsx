import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { createAggregateError } from '../util/errors'

/**
 * Fetches commits.
 */
export function fetchCommits(
    repo: GQLID,
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
        map(({ data, errors }) => {
            if (
                !data ||
                !data.node ||
                !(data.node as GQL.IRepository).commit ||
                !(data.node as GQL.IRepository).commit!.ancestors ||
                !(data.node as GQL.IRepository).commit!.ancestors.nodes
            ) {
                throw createAggregateError(errors)
            }
            return (data.node as GQL.IRepository).commit!.ancestors
        })
    )
}
