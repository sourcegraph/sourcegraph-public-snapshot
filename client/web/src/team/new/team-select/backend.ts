import { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { ParentTeamSelectSearchResult, ParentTeamSelectSearchVariables } from '../../../graphql-operations'

const PARENT_TEAM_SELECT_SEARCH_FIELDS = gql`
    fragment ParentTeamSelectSearchFields on Team {
        id
        name
        displayName
        avatarURL
    }
`

const PARENT_TEAM_SELECT_SEARCH = gql`
    query ParentTeamSelectSearch($search: String!, $exceptAncestorId: ID) {
        teams(search: $search, first: 15, exceptAncestorId: $exceptAncestorId, includeChildTeams: true) {
            nodes {
                ...ParentTeamSelectSearchFields
            }
        }
    }

    ${PARENT_TEAM_SELECT_SEARCH_FIELDS}
`

export function useParentTeamSelectSearch(
    exceptAncestorId: string | null,
    search: string
): QueryResult<ParentTeamSelectSearchResult, ParentTeamSelectSearchVariables> {
    return useQuery<ParentTeamSelectSearchResult, ParentTeamSelectSearchVariables>(PARENT_TEAM_SELECT_SEARCH, {
        variables: {
            search,
            exceptAncestorId,
        },
    })
}
