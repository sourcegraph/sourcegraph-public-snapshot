import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import { RepoBatchChangesResult, RepoBatchChangesVariables } from '../../../graphql-operations'
import { changesetFieldsFragment } from '../detail/backend'

const changesetsStatsFragment = gql`
    fragment ChangesetsStatsFields on ChangesetsStats {
        total
        closed
        deleted
        draft
        merged
        open
        unpublished
        archived
    }
`

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
    RepoBatchChangesResult['repository']
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
            console.log(data)
            return data.repository
        })
    )
