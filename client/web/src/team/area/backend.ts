import { MutationTuple, QueryResult } from '@apollo/client'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'

import {
    ChangeTeamDisplayNameResult,
    ChangeTeamDisplayNameVariables,
    ChangeTeamParentResult,
    ChangeTeamParentVariables,
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

export function useChangeTeamParent(): MutationTuple<ChangeTeamParentResult, ChangeTeamParentVariables> {
    return useMutation<ChangeTeamParentResult, ChangeTeamParentVariables>(
        gql`
            mutation ChangeTeamParent($id: ID!, $parentName: String!) {
                updateTeam(id: $id, parentTeamName: $parentName) {
                    id
                }
            }
        `
    )
}
