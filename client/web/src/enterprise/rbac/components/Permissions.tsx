import React, { ChangeEvent, FocusEventHandler } from 'react'

import { Text, Checkbox, Grid } from '@sourcegraph/wildcard'

import { PermissionsMap, allNamespaces } from '../backend'

interface PermissionListProps {
    allPermissions: PermissionsMap
    isChecked: (value: string) => boolean
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
    onBlur: FocusEventHandler<HTMLInputElement>
}

export const PermissionsList: React.FunctionComponent<React.PropsWithChildren<PermissionListProps>> = ({
    allPermissions,
    isChecked,
    onChange,
    onBlur,
}) => (
    <>
        {allNamespaces.map(namespace => {
            const namespacePermissions = allPermissions[namespace]
            return (
                <div key={namespace}>
                    <Text className="font-weight-bold">{namespace}</Text>
                    <Grid columnCount={4}>
                        {namespacePermissions.map(permission => (
                            <Checkbox
                                key={permission.id}
                                label={permission.action}
                                id={permission.displayName}
                                checked={isChecked(permission.id)}
                                value={permission.id}
                                onChange={onChange}
                                onBlur={onBlur}
                            />
                        ))}
                    </Grid>
                </div>
            )
        })}
    </>
)
