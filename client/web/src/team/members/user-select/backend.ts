import { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { UserSelectSearchResult, UserSelectSearchVariables } from '../../../graphql-operations'

const USER_SELECT_SEARCH_FIELDS = gql`
    fragment UserSelectSearchFields on User {
        id
        username
        displayName
        avatarURL
    }
`

const USER_SELECT_SEARCH = gql`
    query UserSelectSearch($search: String!) {
        users(query: $search, first: 15) {
            nodes {
                ...UserSelectSearchFields
            }
        }
    }

    ${USER_SELECT_SEARCH_FIELDS}
`

export function useUserSelectSearch(
    searchTerm: string
): QueryResult<UserSelectSearchResult, UserSelectSearchVariables> {
    return useQuery<UserSelectSearchResult, UserSelectSearchVariables>(USER_SELECT_SEARCH, {
        variables: {
            search: searchTerm,
        },
    })
}
