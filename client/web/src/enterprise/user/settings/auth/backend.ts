import { Observable } from 'rxjs'
import { mapTo, map, tap } from 'rxjs/operators'

import { resetAllMemoizationCaches } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    ScheduleUserPermissionsSyncResult,
    ScheduleUserPermissionsSyncVariables,
    UserPermissionsInfoFields,
    UserPermissionsInfoResult,
    UserPermissionsInfoVariables,
} from '../../../../graphql-operations'

export function scheduleUserPermissionsSync(args: { user: Scalars['ID'] }): Observable<void> {
    return requestGraphQL<ScheduleUserPermissionsSyncResult, ScheduleUserPermissionsSyncVariables>(
        gql`
            mutation ScheduleUserPermissionsSync($user: ID!) {
                scheduleUserPermissionsSync(user: $user) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        mapTo(undefined)
    )
}

export function userPermissionsInfo(user: Scalars['ID']): Observable<UserPermissionsInfoFields | null> {
    return requestGraphQL<UserPermissionsInfoResult, UserPermissionsInfoVariables>(
        gql`
            query UserPermissionsInfo($user: ID!) {
                node(id: $user) {
                    __typename
                    ... on User {
                        permissionsInfo {
                            ...UserPermissionsInfoFields
                        }
                    }
                }
            }

            fragment UserPermissionsInfoFields on PermissionsInfo {
                syncedAt
                updatedAt
            }
        `,
        { user }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.node) {
                throw new Error('User not found')
            }
            if (data.node.__typename !== 'User') {
                throw new Error(`Node is a ${data.node.__typename}, not a User`)
            }
            return data.node.permissionsInfo
        })
    )
}
