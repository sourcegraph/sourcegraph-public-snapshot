import React, { ChangeEvent, FocusEventHandler } from 'react'

import { Text, Checkbox, Grid } from '@sourcegraph/wildcard'

import { prettifyAction, prettifyNamespace } from '../../../util/settings'
import { PermissionsMap, allNamespaces } from '../backend'

interface PermissionListProps {
    allPermissions: PermissionsMap
    isChecked: (value: string) => boolean
    onChange?: (event: ChangeEvent<HTMLInputElement>) => void
    onBlur?: FocusEventHandler<HTMLInputElement>
    disabled?: boolean
}

export const PermissionsList: React.FunctionComponent<React.PropsWithChildren<PermissionListProps>> = ({
    allPermissions,
    isChecked,
    onChange,
    onBlur,
    disabled,
}) => (
    <>
        {allNamespaces.map(namespace => {
            const namespacePermissions = allPermissions[namespace]
            return (
                <div key={namespace}>
                    <Text className="font-weight-bold">{prettifyNamespace(namespace)}</Text>
                    <Grid columnCount={4}>
                        {namespacePermissions.map(permission => (
                            <Checkbox
                                key={permission.id}
                                label={prettifyAction(permission.action)}
                                id={permission.displayName}
                                checked={isChecked(permission.id)}
                                value={permission.id}
                                onChange={onChange}
                                onBlur={onBlur}
                                disabled={disabled}
                            />
                        ))}
                    </Grid>
                </div>
            )
        })}
    </>
)
