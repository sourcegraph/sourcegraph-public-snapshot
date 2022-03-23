import { gql } from '@apollo/client'

export const GET_EXAMPLE_TODO_REPOSITORY_GQL = gql`
    query ExampleTodoRepository {
        search(patternType: literal, version: V2, query: "select:repo TODO count:1") {
            results {
                repositories {
                    name
                }
            }
        }
    }
`

export const GET_EXAMPLE_FIRST_REPOSITORY_GQL = gql`
    query ExampleFirstRepository {
        search(patternType: literal, version: V2, query: "select:repo count:1") {
            results {
                repositories {
                    name
                }
            }
        }
    }
`
