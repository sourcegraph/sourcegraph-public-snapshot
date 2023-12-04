import type { MutationTuple, QueryResult } from '@apollo/client'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'

import type {
    AssignParentTeamResult,
    AssignParentTeamVariables,
    ChangeTeamDisplayNameResult,
    ChangeTeamDisplayNameVariables,
    RemoveParentTeamResult,
    RemoveParentTeamVariables,
    TeamResult,
    TeamVariables,
} from '../../graphql-operations'

export function useTeam(name: string): QueryResult<TeamResult, TeamVariables> {
    return useQuery<TeamResult, TeamVariables>(
        gql`
            query Team($name: String!) {
                team(name: $name) {
                    ...TeamAreaTeamFields
                }
            }

            fragment TeamAreaTeamFields on Team {
                __typename
                id
                name
                displayName
                avatarURL
                url
                readonly
                parentTeam {
                    id
                    name
                    displayName
                    url
                }
                viewerCanAdminister
                childTeams {
                    totalCount
                }
                members {
                    totalCount
                }
                creator {
                    username
                    displayName
                    avatarURL
                    url
                }
            }
        `,
        {
            variables: { name },
        }
    )
}

export function useChangeTeamDisplayName(): MutationTuple<ChangeTeamDisplayNameResult, ChangeTeamDisplayNameVariables> {
    return useMutation<ChangeTeamDisplayNameResult, ChangeTeamDisplayNameVariables>(
        gql`
            mutation ChangeTeamDisplayName($id: ID!, $displayName: String) {
                updateTeam(id: $id, displayName: $displayName) {
                    id
                }
            }
        `
    )
}

export function useAssignParentTeam(): MutationTuple<AssignParentTeamResult, AssignParentTeamVariables> {
    return useMutation<AssignParentTeamResult, AssignParentTeamVariables>(
        gql`
            mutation AssignParentTeam($id: ID!, $parentTeamName: String!) {
                updateTeam(id: $id, parentTeamName: $parentTeamName) {
                    id
                }
            }
        `
    )
}

export function useRemoveParentTeam(): MutationTuple<RemoveParentTeamResult, RemoveParentTeamVariables> {
    return useMutation<RemoveParentTeamResult, RemoveParentTeamVariables>(
        gql`
            mutation RemoveParentTeam($id: ID!) {
                updateTeam(id: $id, makeRoot: true) {
                    id
                }
            }
        `
    )
}
