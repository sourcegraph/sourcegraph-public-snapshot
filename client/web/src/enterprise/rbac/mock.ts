import { groupBy } from 'lodash'

import { AllPermissionsResult, PermissionNamespace, AllRolesResult } from '../../graphql-operations'

import { PermissionsMap } from './backend'

export const mockPermissions: AllPermissionsResult = {
    permissions: {
        nodes: [
            {
                __typename: 'Permission',
                id: 'perm-1',
                namespace: PermissionNamespace.BATCH_CHANGES,
                action: 'WRITE',
                displayName: 'BATCH_CHANGES#WRITE',
            },
            {
                __typename: 'Permission',
                id: 'perm-2',
                namespace: PermissionNamespace.BATCH_CHANGES,
                action: 'READ',
                displayName: 'BATCH_CHANGES#READ',
            },
            {
                __typename: 'Permission',
                id: 'perm-3',
                namespace: 'CODE_INSIGHTS' as PermissionNamespace,
                action: 'READ',
                displayName: 'CODE_INSIGHTS#READ',
            },
            {
                __typename: 'Permission',
                id: 'perm-4',
                namespace: 'CODE_INSIGHTS' as PermissionNamespace,
                action: 'WRITE',
                displayName: 'CODE_INSIGHTS#WRITE',
            },
            {
                __typename: 'Permission',
                id: 'perm-5',
                namespace: 'REPO_MANAGEMENT' as PermissionNamespace,
                action: 'ADD',
                displayName: 'REPO_MANAGEMENT#ADD',
            },
        ],
    },
}

export const mockRoles: AllRolesResult = {
    roles: {
        __typename: 'RoleConnection',
        totalCount: 5,
        pageInfo: {
            hasNextPage: true,
            endCursor: 'role-3',
        },
        nodes: [
            {
                __typename: 'Role',
                id: 'role-1',
                name: 'Site Administrator',
                system: true,
                permissions: {
                    nodes: mockPermissions.permissions.nodes,
                },
            },
            {
                __typename: 'Role',
                id: 'role-2',
                name: 'Batch Changes Admin',
                system: false,
                permissions: {
                    nodes: [
                        {
                            __typename: 'Permission',
                            id: 'perm-1',
                            namespace: PermissionNamespace.BATCH_CHANGES,
                            action: 'READ',
                            displayName: 'BATCH_CHANGES#WRITE',
                        },
                    ],
                },
            },
            {
                __typename: 'Role',
                id: 'role-3',
                name: 'Operator',
                system: false,
                permissions: {
                    nodes: [],
                },
            },
        ],
    },
}

export const mockPermissionsMap = groupBy(mockPermissions.permissions.nodes, 'namespace') as PermissionsMap
