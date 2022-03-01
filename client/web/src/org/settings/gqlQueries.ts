import { gql } from '@apollo/client'

export const REMOVE_ORG_MUTATION = gql`
    mutation RemoveOrganization($organization: ID!) {
        removeOrganization(organization: $organization) {
            alwaysNil
        }
    }
`
