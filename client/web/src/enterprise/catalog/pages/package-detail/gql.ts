import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { CATALOG_ENTITY_OWNER_FRAGMENT } from '../../components/entity-owner/gql'

const PACKAGE_DETAIL_FRAGMENT = gql`
    fragment PackageDetailFields on Package {
        __typename
        id
        type
        name
        description
        url
        ...CatalogEntityOwnerFields
    }
    ${CATALOG_ENTITY_OWNER_FRAGMENT}
`

export const PACKAGE_BY_NAME = gql`
    query CatalogPackageByName($name: String!) {
        catalogEntity(type: PACKAGE, name: $name) {
            ... on Package {
                ...PackageDetailFields
            }
        }
    }
    ${PACKAGE_DETAIL_FRAGMENT}
`
