import type { MutationTuple, QueryResult } from '@apollo/client'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'

import {
    type DeleteRoleVariables,
    type DeleteRoleResult,
    type AllRolesVariables,
    type AllRolesResult,
    type CreateRoleResult,
    type CreateRoleVariables,
    type AllPermissionsResult,
    type AllPermissionsVariables,
    PermissionNamespace,
    type PermissionFields,
    type SetPermissionsResult,
    type SetPermissionsVariables,
} from '../../graphql-operations'

export const DEFAULT_PAGE_LIMIT = 10

const permissionFragment = gql`
    fragment PermissionFields on Permission {
        __typename
        id
        namespace
        action
        displayName
    }
`

export const roleFragment = gql`
    fragment RoleFields on Role {
        __typename
        id
        name
        system
        permissions {
            nodes {
                ...PermissionFields
            }
        }
    }

    ${permissionFragment}
`

export const ROLES_QUERY = gql`
    query AllRoles {
        roles {
            __typename
            totalCount
            pageInfo {
                hasNextPage
                endCursor
            }
            nodes {
                ...RoleFields
            }
        }
    }

    ${roleFragment}
`

export const CREATE_ROLE = gql`
    mutation CreateRole($name: String!, $permissions: [ID!]!) {
        createRole(name: $name, permissions: $permissions) {
            ...RoleFields
        }
    }

    ${roleFragment}
`

export const DELETE_ROLE = gql`
    mutation DeleteRole($role: ID!) {
        deleteRole(role: $role) {
            alwaysNil
        }
    }
`

export const ALL_PERMISSIONS = gql`
    query AllPermissions {
        permissions {
            nodes {
                ...PermissionFields
            }
        }
    }

    ${permissionFragment}
`

export const SET_PERMISSIONS = gql`
    mutation SetPermissions($role: ID!, $permissions: [ID!]!) {
        setPermissions(role: $role, permissions: $permissions) {
            alwaysNil
        }
    }
`

export const useRolesConnection = (): QueryResult<AllRolesResult, AllRolesVariables> =>
    useQuery(ROLES_QUERY, {
        fetchPolicy: 'no-cache',
    })

export const usePermissions = (
    onCompleted: (result: AllPermissionsResult) => void
): QueryResult<AllPermissionsResult, AllPermissionsVariables> =>
    useQuery<AllPermissionsResult, AllPermissionsVariables>(ALL_PERMISSIONS, {
        fetchPolicy: 'cache-and-network',
        onCompleted,
    })

export const useCreateRole = (onCompleted: () => void): MutationTuple<CreateRoleResult, CreateRoleVariables> =>
    useMutation(CREATE_ROLE, { onCompleted })

export const useDeleteRole = (
    onCompleted: () => void,
    onError: () => void
): MutationTuple<DeleteRoleResult, DeleteRoleVariables> => useMutation(DELETE_ROLE, { onCompleted, onError })

export const useSetPermissions = (
    onCompleted: () => void
): MutationTuple<SetPermissionsResult, SetPermissionsVariables> => useMutation(SET_PERMISSIONS, { onCompleted })

export type PermissionsMap = Record<PermissionNamespace, PermissionFields[]>

// Permissions are grouped by their namespace in the UI. We do this to get all unique namespaces
// on the Sourcegraph instance.
export const allNamespaces = Object.values<PermissionNamespace>(PermissionNamespace)
