import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../backend/graphql'

/**
 * Fetch LSIF uploads for a repository.
 */
export function fetchLsifUploads({
    repository,
    query,
    state,
    isLatestForRepo,
    first,
    after,
}: { repository: string } & GQL.ILsifUploadsOnRepositoryArguments): Observable<GQL.ILSIFUploadConnection> {
    return queryGraphQL(
        gql`
            query LsifUploads(
                $repository: ID!
                $state: LSIFUploadState
                $isLatestForRepo: Boolean
                $first: Int
                $after: String
                $query: String
            ) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        lsifUploads(
                            query: $query
                            state: $state
                            isLatestForRepo: $isLatestForRepo
                            first: $first
                            after: $after
                        ) {
                            nodes {
                                id
                                state
                                projectRoot {
                                    commit {
                                        abbreviatedOID
                                    }
                                    path
                                    url
                                }
                                inputRepoName
                                inputCommit
                                inputRoot
                                uploadedAt
                                startedAt
                                finishedAt
                            }

                            totalCount
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                        }
                    }
                }
            }
        `,
        { repository, query, state, isLatestForRepo, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is a ${node.__typename}, not a Repository`)
            }

            return node.lsifUploads
        })
    )
}
