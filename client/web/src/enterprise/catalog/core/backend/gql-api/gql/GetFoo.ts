import { gql } from '@apollo/client'

export const GET_FOO_GQL = gql`
    query GetFoo {
        catalog {
            foo
        }
    }
`
