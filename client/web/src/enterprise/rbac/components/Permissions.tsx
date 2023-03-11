import React, { useMemo } from 'react'

import { mdiMapSearch } from '@mdi/js'
import { groupBy } from 'lodash'

import { Icon, Text, Checkbox, Grid, Form } from '@sourcegraph/wildcard'

import { RoleFields } from '../../../graphql-operations'
import { PermissionsMap, allNamespaces } from '../backend'

const EmptyPermissionList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center m-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No permissions associated with this role.</div>
    </div>
)

interface PermissionListProps {
    role: RoleFields
    allPermissions: PermissionsMap
}

export const PermissionList: React.FunctionComponent<React.PropsWithChildren<PermissionListProps>> = ({
    role,
    allPermissions,
}) => {
    const rolePermissions = role.permissions.nodes
    // We create a map for the role permissions using their ID, so we can perform an easy lookup when rendering
    // the list of all permissions.
    const rolePermissionsMap = useMemo(() => groupBy(rolePermissions, 'id'), [rolePermissions])

    // We display EmptyPermissionList when the role has no permissions assigned to it.
    if (rolePermissions.length === 0) {
        return <EmptyPermissionList />
    }

    return (
        <>
            {allNamespaces.map(namespace => {
                const namespacePermissions = allPermissions[namespace]
                return (
                    <Form key={namespace}>
                        <Text className="font-weight-bold">{namespace}</Text>
                        <Grid columnCount={4}>
                            {namespacePermissions.map(permission => {
                                const isChecked = Boolean(rolePermissionsMap[permission.id])
                                return (
                                    <Checkbox
                                        key={permission.id}
                                        label={permission.action}
                                        id={permission.displayName}
                                        defaultChecked={isChecked}
                                    />
                                )
                            })}
                        </Grid>
                    </Form>
                )
            })}
        </>
    )
}
