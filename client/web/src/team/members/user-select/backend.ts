import type { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import type { TeamMemberUserSelectSearchResult, TeamMemberUserSelectSearchVariables } from '../../../graphql-operations'

const USER_SELECT_SEARCH_FIELDS = gql`
    fragment TeamMemberUserSelectSearchFields on User {
        id
        username
        displayName
        avatarURL
    }
`

const USER_SELECT_SEARCH = gql`
    query TeamMemberUserSelectSearch($search: String!) {
        users(query: $search, first: 15) {
            nodes {
                ...TeamMemberUserSelectSearchFields
            }
        }
    }

    ${USER_SELECT_SEARCH_FIELDS}
`

export function useUserSelectSearch(
    searchTerm: string
): QueryResult<TeamMemberUserSelectSearchResult, TeamMemberUserSelectSearchVariables> {
    return useQuery<TeamMemberUserSelectSearchResult, TeamMemberUserSelectSearchVariables>(USER_SELECT_SEARCH, {
        variables: {
            search: searchTerm,
        },
    })
}
