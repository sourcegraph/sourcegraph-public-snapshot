import { gql } from '@apollo/client'

export const LIST_COMPONENTS_GQL = gql`
    query ListComponents($query: String) {
        catalog {
            components(query: $query) {
                nodes {
                    id
                    name
                }
            }
        }
    }
`
