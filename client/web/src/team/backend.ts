import type { MutationTuple } from '@apollo/client'

import { gql, useMutation } from '@sourcegraph/http-client'

import type { CreateTeamResult, CreateTeamVariables } from '../graphql-operations'

export function useCreateTeam(): MutationTuple<CreateTeamResult, CreateTeamVariables> {
    return useMutation<CreateTeamResult, CreateTeamVariables>(
        gql`
            mutation CreateTeam($name: String!, $displayName: String, $parentTeam: String) {
                createTeam(name: $name, displayName: $displayName, parentTeamName: $parentTeam) {
                    id
                    name
                    url
                }
            }
        `
    )
}
