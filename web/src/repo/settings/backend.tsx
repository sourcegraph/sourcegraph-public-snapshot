import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

/**
 * Fetches a repository.
 */
export function fetchRepository(name: string): Observable<GQL.IRepository> {
    return queryGraphQL(
        gql`
            query Repository($name: String!) {
                repository(name: $name) {
                    id
                    name
                    viewerCanAdminister
                    mirrorInfo {
                        remoteURL
                        cloneInProgress
                        cloneProgress
                        cloned
                        updatedAt
                        updateSchedule {
                            due
                            index
                            total
                        }
                        updateQueue {
                            updating
                            index
                            total
                        }
                    }
                    externalServices {
                        nodes {
                            id
                            kind
                            displayName
                        }
                    }
                }
            }
        `,
        { name }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repository) {
                throw createAggregateError(errors)
            }
            return data.repository
        })
    )
}

/**
 * Fetch of LSIF dumps for a repository.
 */
export function fetchLsifDumps(repository: string): Observable<GQL.ILSIFDumpConnection> {
    return queryGraphQL(
        gql`
            query LsifDumps($repository: String!) {
                lsifDumps(repository: $repository) {
                    nodes {
                        id
                        repository
                        commit
                        root
                    }
                }
            }
        `,
        { repository }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifDumps)
    )
}
