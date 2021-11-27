import { gql } from '@apollo/client'

export const LIST_COMPONENTS_GQL = gql`
    query ListComponents {
        catalog {
            components {
                nodes {
                    name
                }
            }
        }
    }
`
