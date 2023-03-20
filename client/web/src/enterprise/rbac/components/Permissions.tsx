import React, { ChangeEvent, FocusEventHandler } from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { Text, Checkbox, Grid, Tooltip, Icon } from '@sourcegraph/wildcard'

import { BatchChangesReadPermission } from '../../../rbac/constants'
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
                        {namespacePermissions.map(permission => {
                            // This is a hack to disable the BatchChangesReadPermission
                            // from the UI for now until it's fully implemented.
                            if (permission.displayName === BatchChangesReadPermission) {
                                return (
                                    <Checkbox
                                        key={permission.id}
                                        label={
                                            <>
                                                {prettifyAction(permission.action)}
                                                <Tooltip content="Coming soon">
                                                    <Icon
                                                        aria-label="Batch changes read access restrictions coming soon"
                                                        className="ml-2"
                                                        svgPath={mdiInformationOutline}
                                                    />
                                                </Tooltip>
                                            </>
                                        }
                                        id={permission.displayName}
                                        checked={isChecked(permission.id)}
                                        value={permission.id}
                                        disabled={true}
                                    />
                                )
                            }
                            return (
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
                            )
                        })}
                    </Grid>
                </div>
            )
        })}
    </>
)
