import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../backend/graphql'
import type {
    RepoBatchChangesResult,
    RepoBatchChangesVariables,
    RepoBatchChangeStatsVariables,
    RepoBatchChangeStatsResult,
} from '../../../graphql-operations'
import { changesetFieldsFragment } from '../detail/backend'

const repoBatchChangeStatsFragment = gql`
    fragment RepoBatchChangeStats on Repository {
        batchChangesDiffStat {
            added
            deleted
        }
        changesetsStats {
            unpublished
            draft
            open
            merged
            closed
            total
        }
    }
`

export const queryRepoBatchChangeStats = ({
    name,
}: RepoBatchChangeStatsVariables): Observable<NonNullable<RepoBatchChangeStatsResult['repository']>> =>
    requestGraphQL<RepoBatchChangeStatsResult, RepoBatchChangeStatsVariables>(
        gql`
            query RepoBatchChangeStats($name: String!) {
                repository(name: $name) {
                    ...RepoBatchChangeStats
                }
            }

            ${repoBatchChangeStatsFragment}
        `,
        { name }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository) {
                throw new Error(`Repository "${name}" not found`)
            }
            return data.repository
        })
    )

export const MAX_CHANGESETS_COUNT = 10

const repoBatchChangeFragment = gql`
    fragment RepoBatchChange on BatchChange {
        id
        url
        name
        namespace {
            namespaceName
            url
        }
        description
        createdAt
        closedAt
        changesets(first: 10, repo: $repoID) {
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
            nodes {
                ...ChangesetFields
            }
        }
        changesetsStats {
            open
            closed
            merged
        }
    }

    ${changesetFieldsFragment}
`

export const queryRepoBatchChanges = ({
    name,
    repoID,
    first = null,
    after = null,
    state = null,
    viewerCanAdminister = null,
}: Partial<RepoBatchChangesVariables> & Pick<RepoBatchChangesVariables, 'name' | 'repoID'>): Observable<
    NonNullable<RepoBatchChangesResult['repository']>
> =>
    requestGraphQL<RepoBatchChangesResult, RepoBatchChangesVariables>(
        gql`
            query RepoBatchChanges(
                $name: String!
                $repoID: ID!
                $first: Int
                $after: String
                $state: BatchChangeState
                $viewerCanAdminister: Boolean
            ) {
                repository(name: $name) {
                    batchChanges(
                        first: $first
                        after: $after
                        state: $state
                        viewerCanAdminister: $viewerCanAdminister
                    ) {
                        nodes {
                            ...RepoBatchChange
                        }
                        pageInfo {
                            endCursor
                            hasNextPage
                        }
                        totalCount
                    }
                }
            }

            ${repoBatchChangeFragment}
        `,
        { name, repoID, first, after, state, viewerCanAdminister }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository) {
                throw new Error(`Repository "${name}" not found`)
            }
            return data.repository
        })
    )
