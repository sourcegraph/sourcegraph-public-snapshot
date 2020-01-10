import * as GQL from '../../../../shared/src/graphql/schema'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { queryGraphQL } from '../../backend/graphql'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'

/**
 * Fetch a single LSIF upload by id.
 */
export function fetchLsifUpload({ id }: { id: string }): Observable<GQL.ILSIFUpload | null> {
    return queryGraphQL(
        gql`
            query LsifUpload($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on LSIFUpload {
                        id
                        projectRoot {
                            commit {
                                oid
                                abbreviatedOID
                                url
                                repository {
                                    name
                                    url
                                }
                            }
                            path
                            url
                        }
                        inputRepoName
                        inputCommit
                        inputRoot
                        state
                        failure {
                            summary
                        }
                        uploadedAt
                        startedAt
                        finishedAt
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'LSIFUpload') {
                throw new Error(`The given ID is a ${node.__typename}, not an LSIFUpload`)
            }

            return node
        })
    )
}
