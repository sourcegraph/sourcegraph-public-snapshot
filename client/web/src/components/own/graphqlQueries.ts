import { gql } from '@sourcegraph/http-client'

export const ASSIGN_OWNER = gql`
    mutation AssignOwner($input: AssignOwnerOrTeamInput!) {
        assignOwner(input: $input) {
            alwaysNil
        }
    }
`

export const ASSIGN_TEAM = gql`
    mutation AssignTeam($input: AssignOwnerOrTeamInput!) {
        assignTeam(input: $input) {
            alwaysNil
        }
    }
`
