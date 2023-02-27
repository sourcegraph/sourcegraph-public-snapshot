import React, { useMemo } from 'react'

import { groupBy } from 'lodash'
import { mdiMapSearch } from '@mdi/js'

import { Icon, Text, Checkbox, Grid, Form } from '@sourcegraph/wildcard'

import { RoleFields, PermissionNamespace } from '../../../graphql-operations'
import { PermissionsMap } from '../backend'

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
    if (rolePermissions.length === 0) {
        return <EmptyPermissionList />
    }

    const allNamespaces = Object.values(PermissionNamespace)
    const rolePermissionsMap = useMemo(() => groupBy(rolePermissions, 'id'), [rolePermissions])
    return (
        <>
            {allNamespaces.map(namespace => {
                const namespacePermissions = allPermissions[namespace as PermissionNamespace]
                return (
                    <Form key={namespace}>
                        <Text className="font-weight-bold">{namespace}</Text>
                        <Grid columnCount={4}>
                            {namespacePermissions.map(ap => {
                                const isChecked = Boolean(rolePermissionsMap[ap.id])
                                return (
                                    <Checkbox
                                        key={ap.id}
                                        label={ap.action}
                                        id={ap.displayName}
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
