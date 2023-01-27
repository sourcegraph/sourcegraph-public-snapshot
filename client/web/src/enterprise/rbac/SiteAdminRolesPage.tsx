import React, { useEffect, FC, useCallback, useState, useMemo } from 'react'

import { RouteComponentProps } from 'react-router'
import { mdiPlus, mdiChevronUp, mdiChevronDown, mdiMapSearch } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon, Container, Text, Tooltip, Checkbox, Grid } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'

import styles from './SiteAdminRolesPage.module.scss'

export interface SiteAdminRolesPageProps extends RouteComponentProps, TelemetryProps {}

type Role = {
    id: number,
    name: string,
    system: boolean
}

type Permission = { displayName: string, action: string, namespace: string }
type PermissionMap = Record<string, Record<string, Permission>>

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])


    const allPermsMap = useMemo(() => {
        const allPermissions = getAllPermissions()
        return allPermissions.reduce<PermissionMap>((acc, curr) => {
            const { displayName, namespace } = curr
            const namespaceDetails = acc[namespace] ? { ...acc[namespace], [displayName]: curr } : { [displayName]: curr }

            return { ...acc, [namespace]: namespaceDetails }
        }, {})
    }, [])

    const sampleRoles: Role[] = [
        {
            id: 1,
            name: 'USER',
            system: true
        },
        {
            id: 2,
            name: 'SITE_ADMINISTRATOR',
            system: true
        },
        {
            id: 3,
            name: 'TEST-ROLE-3',
            system: false
        },
        {
            id: 4,
            name: 'TEST-ROLE-4',
            system: false
        },
        {
            id: 5,
            name: 'TEST-ROLE-5',
            system: false
        },
    ]

    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                path={[{ text: 'Roles' }]}
                headingElement="h2"
                description={
                    <>
                        A role is a set of permissions assigned to a user.
                    </>
                }
                className="mb-3"
                actions={
                    <Button variant="primary">
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add Role
                    </Button>
                }
            />

            <Container>
                <ul className={styles.rolesList}>
                    {sampleRoles.map(role => <RoleNode node={role} key={role.id} permissions={allPermsMap} />)}
                </ul>
            </Container>
        </div>
    )
}

const RoleNode: FC<{
    node: Role
    permissions: PermissionMap
}> = ({ node, permissions }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    return (
        <li className={styles.roleNode}>
             <Button
                variant="icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
            </Button>

            <div className="d-flex">
                <Text className="font-weight-bold m-0">
                    {node.name}
                </Text>

                {node.system && (
                    <Tooltip content="System roles are sourcegraph-seeded roles available on every instance." placement="topStart">
                        <Text className={styles.roleNodeSystemText}>System</Text>
                    </Tooltip>
                )}
            </div>

            {isExpanded ? (
                <div className={styles.roleNodePermissions}>
                    <PermissionList roleId={node.id} permissions={permissions} />
                </div>
            ) : <span />}
        </li>
    )
}

const getRolePermissions = (roleId: number): Array<{ namespace: string, displayName: string, action: string}> => {
    if (roleId === 2) {
        return []
    }

    return [
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:READ',
            action: 'READ'
        },
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:APPLY',
            action: 'APPLY'
        },
        {
            namespace: 'NOTEBOOKS',
            displayName: 'NOTEBOOKS:READ',
            action: 'READ'
        },
    ]
}

const getAllPermissions = (): Array<{ namespace: string, displayName: string, action: string}> => {
    return [
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:READ',
            action: 'READ'
        },
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:WRITE',
            action: 'WRITE'
        },
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:EXECUTE',
            action: 'EXECUTE'
        },
        {
            namespace: 'BATCHCHANGES',
            displayName: 'BATCHCHANGES:APPLY',
            action: 'APPLY'
        },
        {
            namespace: 'NOTEBOOKS',
            displayName: 'NOTEBOOKS:READ',
            action: 'READ'
        },
        {
            namespace: 'NOTEBOOKS',
            displayName: 'NOTEBOOKS:WRITE',
            action: 'WRITE'
        },
        {
            namespace: 'CODEINSIGHTS',
            displayName: 'CODEINSIGHTS:READ',
            action: 'READ'
        },
    ]
}


const EmptyPermissionList: FC<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center m-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No permissions associated with this role.</div>
    </div>
)

const PermissionList: FC<React.PropsWithChildren<{ roleId: number, permissions: PermissionMap }>> = ({ roleId, permissions }) => {
    const rolePermissions = getRolePermissions(roleId)

    if (rolePermissions.length === 0) {
        return <EmptyPermissionList />
    }

    // const allDisplayNames = rolePermissions.map(rp => rp.displayName)
    const permissionsDisplayMap: PermissionMap[keyof PermissionMap] = rolePermissions.reduce((acc, permission) => {
        const { displayName } = permission
        return { ...acc, [displayName]: true }
    }, {})

    const namespaces = Object.keys(permissions)
    console.log(permissionsDisplayMap)

    return (
        <>
            {namespaces.map((namespace) => {
                const namespacePerms = permissions[namespace]
                const allNamespacePerms = Object.values(namespacePerms)
                return (
                    <div key={namespace}>
                        <Text>{namespace}</Text>
                        <Grid columnCount={4}>
                            {allNamespacePerms.map((ap, index) => {
                                const isChecked = Boolean(permissionsDisplayMap[ap.displayName])
                                return (
                                    <Checkbox label={ap.action} id={ap.displayName} key={ap.displayName} defaultChecked={isChecked} />
                                )
                            })}

                        </Grid>
                    </div>
                )
            })}
            <Button variant="primary">Update</Button>
        </>
    )
}
