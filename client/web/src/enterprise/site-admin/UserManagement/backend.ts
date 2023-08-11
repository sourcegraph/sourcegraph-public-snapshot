import type { MutationTuple, QueryResult } from '@apollo/client'

import { gql, useMutation, useQuery } from '@sourcegraph/http-client'

import type {
    GetAllRolesAndUserRolesResult,
    GetAllRolesAndUserRolesVariables,
    Scalars,
    SetRolesForUserResult,
    SetRolesForUserVariables,
} from '../../../graphql-operations'
import { roleFragment } from '../../rbac/backend'

export const GET_ALL_ROLES_AND_USER_ROLES = gql`
    query GetAllRolesAndUserRoles($user: ID!) {
        node(id: $user) {
            ... on User {
                roles {
                    nodes {
                        ...RoleFields
                    }
                }
            }
        }

        roles {
            nodes {
                ...RoleFields
            }
        }
    }

    ${roleFragment}
`

export const SET_ROLES_FOR_USER = gql`
    mutation SetRolesForUser($user: ID!, $roles: [ID!]!) {
        setRoles(user: $user, roles: $roles) {
            alwaysNil
        }
    }
`

export const useSetRoles = (onCompleted: () => void): MutationTuple<SetRolesForUserResult, SetRolesForUserVariables> =>
    useMutation<SetRolesForUserResult, SetRolesForUserVariables>(SET_ROLES_FOR_USER, { onCompleted })

export const useGetUserRolesAndAllRoles = (
    user: Scalars['ID'],
    onCompleted: (data: GetAllRolesAndUserRolesResult) => void
): QueryResult<GetAllRolesAndUserRolesResult, GetAllRolesAndUserRolesVariables> =>
    useQuery<GetAllRolesAndUserRolesResult, GetAllRolesAndUserRolesVariables>(GET_ALL_ROLES_AND_USER_ROLES, {
        fetchPolicy: 'cache-and-network',
        variables: { user },
        onCompleted,
    })
