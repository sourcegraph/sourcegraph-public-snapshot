import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { requestGraphQL } from '../../../backend/graphql'
import {
    LsifUploadConnectionFields,
    LsifUploadsForRepoResult,
    LsifUploadsForRepoVariables,
    LsifUploadsResult,
    LsifUploadsVariables,
} from '../../../graphql-operations'

/**
 * Return LSIF uploads. If a repository is given, only uploads for that repository will be returned. Otherwise,
 * uploads across all repositories are returned.
 */
export function fetchLsifUploads({
    repository,
    query,
    state,
    isLatestForRepo,
    dependencyOf,
    dependentOf,
    first,
    after,
}: { repository?: string } & GQL.ILsifUploadsOnRepositoryArguments): Observable<LsifUploadConnectionFields> {
    const vars: LsifUploadsVariables = {
        query: query ?? null,
        state: state ?? null,
        isLatestForRepo: isLatestForRepo ?? null,
        dependencyOf: dependencyOf ?? null,
        dependentOf: dependentOf ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    if (repository) {
        const gqlQuery = gql`
            query LsifUploadsForRepo(
                $repository: ID!
                $state: LSIFUploadState
                $isLatestForRepo: Boolean
                $dependencyOf: ID
                $dependentOf: ID
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
                            dependencyOf: $dependencyOf
                            dependentOf: $dependentOf
                            first: $first
                            after: $after
                        ) {
                            ...LsifUploadConnectionFields
                        }
                    }
                }
            }

            fragment LsifUploadConnectionFields on LSIFUploadConnection {
                nodes {
                    ...LsifUploadFields
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }

            ${lsifUploadFieldsFragment}
        `

        return requestGraphQL<LsifUploadsForRepoResult, LsifUploadsForRepoVariables>(gqlQuery, {
            ...vars,
            repository,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error('Invalid repository')
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`The given ID is ${node.__typename}, not Repository`)
                }

                return node.lsifUploads
            })
        )
    }

    const gqlQuery = gql`
        query LsifUploads(
            $state: LSIFUploadState
            $isLatestForRepo: Boolean
            $dependencyOf: ID
            $dependentOf: ID
            $first: Int
            $after: String
            $query: String
        ) {
            lsifUploads(
                query: $query
                state: $state
                isLatestForRepo: $isLatestForRepo
                dependencyOf: $dependencyOf
                dependentOf: $dependentOf
                first: $first
                after: $after
            ) {
                nodes {
                    ...LsifUploadFields
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }
        }

        ${lsifUploadFieldsFragment}
    `

    return requestGraphQL<LsifUploadsResult, LsifUploadsVariables>(gqlQuery, vars).pipe(
        map(dataOrThrowErrors),
        map(({ lsifUploads }) => lsifUploads)
    )
}

export const lsifUploadFieldsFragment = gql`
    fragment LsifUploadFields on LSIFUpload {
        __typename
        id
        inputCommit
        inputRoot
        inputIndexer
        projectRoot {
            url
            path
            repository {
                url
                name
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        state
        failure
        isLatestForRepo
        uploadedAt
        startedAt
        finishedAt
        placeInQueue
        associatedIndex {
            id
            state
            queuedAt
            startedAt
            finishedAt
            placeInQueue
        }
    }
`

export const lsifIndexFieldsFragment = gql`
    fragment LsifIndexFields on LSIFIndex {
        __typename
        id
        inputCommit
        inputRoot
        inputIndexer
        projectRoot {
            url
            path
            repository {
                url
                name
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        steps {
            ...LsifIndexStepsFields
        }
        state
        failure
        queuedAt
        startedAt
        finishedAt
        placeInQueue
        associatedUpload {
            id
            state
            uploadedAt
            startedAt
            finishedAt
            placeInQueue
        }
    }
    fragment LsifIndexStepsFields on IndexSteps {
        setup {
            ...ExecutionLogEntryFields
        }
        preIndex {
            root
            image
            commands
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        index {
            indexerArgs
            outfile
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        upload {
            ...ExecutionLogEntryFields
        }
        teardown {
            ...ExecutionLogEntryFields
        }
    }
    fragment ExecutionLogEntryFields on ExecutionLogEntry {
        key
        command
        startTime
        exitCode
        out
        durationMilliseconds
    }
`
