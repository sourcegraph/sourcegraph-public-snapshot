import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { COMPONENT_FOR_EXPLORER_FRAGMENT } from './gql'

export const COMPONENT_RELATIONS_FOR_EXPLORER = gql`
    query ComponentRelationsForExplorer($entity: ID!, $query: String, $first: Int, $after: String) {
        node(id: $entity) {
            __typename
            ... on Component {
                relatedEntities(query: $query, first: $first, after: $after) {
                    edges {
                        ...ComponentRelationFields
                    }
                }
            }
        }
    }

    fragment ComponentRelationFields on ComponentRelatedEntityEdge {
        type
        node {
            ...ComponentForExplorerFields
        }
    }

    ${COMPONENT_FOR_EXPLORER_FRAGMENT}
`
