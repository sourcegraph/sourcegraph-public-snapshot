import { gql } from '@sourcegraph/http-client'

import { COMPONENT_STATE_FRAGMENT } from '../../pages/overview/components/catalog-explorer/gql'
import { COMPONENT_OWNER_FRAGMENT } from '../entity-owner/gql'

const CATALOG_HEALTH_FRAGMENT = gql`
    fragment ComponentHealthFields on Component {
        __typename
        id
        name
        kind
        url
        status {
            id
            contexts {
                id
                name
                state
                title
                description
                targetURL
            }
        }
        ...ComponentOwnerFields
        ...ComponentStateFields
    }
    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_STATE_FRAGMENT}
`

export const CATALOG_HEALTH = gql`
    query CatalogHealth($query: String, $first: Int, $after: String) {
        components(query: $query, first: $first, after: $after) {
            nodes {
                ...ComponentHealthFields
            }
        }
    }
    ${CATALOG_HEALTH_FRAGMENT}
`
