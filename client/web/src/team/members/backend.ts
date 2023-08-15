import type { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    AddTeamMembersResult,
    AddTeamMembersVariables,
    ListTeamMemberFields,
    ListTeamMembersResult,
    ListTeamMembersVariables,
    RemoveTeamMembersResult,
    RemoveTeamMembersVariables,
} from '../../graphql-operations'

const LIST_TEAM_MEMBER_FIELDS = gql`
    fragment ListTeamMemberFields on TeamMember {
        __typename
        ... on User {
            __typename
            id
            url
            username
            displayName
            avatarURL
        }
    }
`

export const LIST_TEAM_MEMBERS = gql`
    query ListTeamMembers($first: Int, $after: String, $search: String, $teamName: String!) {
        team(name: $teamName) {
            id
            members(first: $first, after: $after, search: $search) {
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
                nodes {
                    ...ListTeamMemberFields
                }
            }
        }
    }

    ${LIST_TEAM_MEMBER_FIELDS}
`

export function useTeamMembers(
    teamName: string,
    search: string | null
): UseShowMorePaginationResult<ListTeamMembersResult, ListTeamMemberFields> {
    return useShowMorePagination<ListTeamMembersResult, ListTeamMembersVariables, ListTeamMemberFields>({
        query: LIST_TEAM_MEMBERS,
        variables: {
            after: null,
            first: 15,
            search,
            teamName,
        },
        options: {
            fetchPolicy: 'network-only',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.team) {
                throw new Error(`Team ${teamName} not found`)
            }

            return data.team.members
        },
    })
}

const REMOVE_TEAM_MEMBERS = gql`
    mutation RemoveTeamMembers($team: ID!, $members: [TeamMemberInput!]!) {
        removeTeamMembers(team: $team, members: $members) {
            id
        }
    }
`

export function useRemoveTeamMembers(): MutationTuple<RemoveTeamMembersResult, RemoveTeamMembersVariables> {
    return useMutation<RemoveTeamMembersResult, RemoveTeamMembersVariables>(REMOVE_TEAM_MEMBERS)
}

const ADD_TEAM_MEMBERS = gql`
    mutation AddTeamMembers($team: ID!, $members: [TeamMemberInput!]!) {
        addTeamMembers(team: $team, members: $members) {
            id
        }
    }
`

export function useAddTeamMembers(): MutationTuple<AddTeamMembersResult, AddTeamMembersVariables> {
    return useMutation<AddTeamMembersResult, AddTeamMembersVariables>(ADD_TEAM_MEMBERS)
}
