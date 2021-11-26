import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_GRAPH_FRAGMENT } from '../../../../components/entity-graph/gql'

export const CATALOG_GRAPH = gql`
    query CatalogGraph($query: String!) {
        graph(query: $query) {
            ...CatalogGraphFields
        }
    }

    ${CATALOG_GRAPH_FRAGMENT}
`
