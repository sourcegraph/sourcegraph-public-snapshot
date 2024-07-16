import type { MutationTuple } from '@apollo/client'

import { dataOrThrowErrors, gql, useMutation } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    DeleteTeamResult,
    DeleteTeamVariables,
    ListTeamFields,
    ListTeamsOfParentResult,
    ListTeamsOfParentVariables,
    ListTeamsResult,
    ListTeamsVariables,
} from '../../graphql-operations'

const LIST_TEAM_FIELDS = gql`
    fragment ListTeamFields on Team {
        id
        name
        displayName
        url
        readonly
        avatarURL
        members {
            totalCount
        }
        parentTeam {
            id
            name
            displayName
            url
        }
        childTeams {
            totalCount
        }
        viewerCanAdminister
    }
`

export const LIST_TEAMS = gql`
    query ListTeams($first: Int, $after: String, $search: String) {
        teams(first: $first, after: $after, search: $search) {
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
            nodes {
                ...ListTeamFields
            }
        }
    }

    ${LIST_TEAM_FIELDS}
`

export const LIST_TEAMS_OF_PARENT = gql`
    query ListTeamsOfParent($first: Int, $after: String, $search: String, $teamName: String!) {
        team(name: $teamName) {
            childTeams(first: $first, after: $after, search: $search) {
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
                nodes {
                    ...ListTeamFields
                }
            }
        }
    }

    ${LIST_TEAM_FIELDS}
`

export function useTeams(search: string | null): UseShowMorePaginationResult<ListTeamsResult, ListTeamFields> {
    return useShowMorePagination<ListTeamsResult, ListTeamsVariables, ListTeamFields>({
        query: LIST_TEAMS,
        variables: {
            search,
        },
        options: {
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => dataOrThrowErrors(result).teams,
    })
}

export function useChildTeams(
    parentTeam: string,
    search: string | null
): UseShowMorePaginationResult<ListTeamsOfParentResult, ListTeamFields> {
    return useShowMorePagination<ListTeamsOfParentResult, ListTeamsOfParentVariables, ListTeamFields>({
        query: LIST_TEAMS_OF_PARENT,
        variables: {
            teamName: parentTeam,
            search,
        },
        options: {
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const { team } = dataOrThrowErrors(result)

            if (!team) {
                throw new Error(`Parent team ${parentTeam} not found`)
            }

            return team.childTeams
        },
    })
}

const DELETE_TEAM = gql`
    mutation DeleteTeam($id: ID!) {
        deleteTeam(id: $id) {
            alwaysNil
        }
    }
`

export function useDeleteTeam(): MutationTuple<DeleteTeamResult, DeleteTeamVariables> {
    return useMutation<DeleteTeamResult, DeleteTeamVariables>(DELETE_TEAM)
}
