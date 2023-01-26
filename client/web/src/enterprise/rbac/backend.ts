import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'

const PERMISSIONS_FRAGMENT = gql`
    fragment Permission on Permission {
        id
        namespace
        action
    }
`

const ROLES_WITH_PERMISSIONS_FRAGMENT = gql`
    fragment RolesWithPermissions on Role {
        id
        name
        system
    }
`

export const ROLES_QUERY = gql`
    query RolsList($first: Int, $after: String) {
        roles(first: $first, after: $after) {
            totalCount
            pageInfo {
                hasNextPage
                endCursor
            }
            nodes {
                ...RolesWithPermissions
            }
        }
    }

    ${ROLES_WITH_PERMISSIONS_FRAGMENT}
`

export const useRolesConnection = () => useShowMorePagination({
    query: ROLES_QUERY,
    variables: {
        first: 20,
        after: null,
    },
    getConnection: result => {
        const { roles } = dataOrThrowErrors(result)
        return roles
    }
})
