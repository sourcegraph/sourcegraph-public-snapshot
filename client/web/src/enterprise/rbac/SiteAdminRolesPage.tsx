import React, { useEffect, FC, useCallback, useState } from 'react'

import { RouteComponentProps } from 'react-router'
import { mdiPlus, mdiChevronUp, mdiChevronDown, mdiMapSearch, mdiDelete } from '@mdi/js'

import { logger } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon, Text, Tooltip, LoadingSpinner } from '@sourcegraph/wildcard'
import { RoleFields } from '../../graphql-operations'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'

import { useRolesConnection, useDeleteRole } from './backend'
import { PageTitle } from '../../components/PageTitle'

import styles from './SiteAdminRolesPage.module.scss'

export interface SiteAdminRolesPageProps extends RouteComponentProps, TelemetryProps {}

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])

    const { connection, error, loading, fetchMore, hasNextPage, refetchAll } = useRolesConnection()

    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                path={[{ text: 'Roles' }]}
                headingElement="h2"
                description={
                    <>
                        Roles represent a characteristic of a group of users, it can be used to define a function or
                        attribute a group of users possess.
                    </>
                }
                className="mb-3"
                actions={
                    <Button variant="primary">
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add Role
                    </Button>
                }
            />

            <ConnectionContainer className="mb-3">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                <ConnectionList as="ul" className="list-group" aria-label="Roles">
                    {connection?.nodes?.map(node => (
                        <RoleNode key={node.id} node={node} afterDelete={refetchAll} />
                    ))}
                </ConnectionList>
                {connection && (
                    <SummaryContainer className="mt-2">
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={15}
                            centered={true}
                            connection={connection}
                            noun="role"
                            pluralNoun="roles"
                            hasNextPage={hasNextPage}
                        />
                        {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </div>
    )
}

const RoleNode: FC<{
    node: RoleFields
    afterDelete: () => void
}> = ({ node, afterDelete }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )
    const [deleteRole, { loading, error }] = useDeleteRole()
    const onDelete = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await deleteRole({ variables: { role: node.id } })
                afterDelete()
            } catch (error) {
                logger.error(error)
            }
        },
        [deleteRole, name, afterDelete]
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

            <div className="d-flex align-items-center">
                <Text className="font-weight-bold m-0">{node.name}</Text>

                {node.system && (
                    <Tooltip
                        content="System roles are sourcegraph-seeded roles available on every instance."
                        placement="topStart"
                    >
                        <Text className={styles.roleNodeSystemText}>System</Text>
                    </Tooltip>
                )}
            </div>

            {loading ? (
                <LoadingSpinner />
            ) : (
                <Tooltip content={node.system ? 'System roles cannot be deleted.' : 'Delete this role.'}>
                    <Button
                        aria-label="Delete"
                        onClick={onDelete}
                        disabled={node.system || loading}
                        variant="danger"
                        size="sm"
                    >
                        <Icon aria-hidden={true} svgPath={mdiDelete} />
                    </Button>
                </Tooltip>
            )}

            {isExpanded ? (
                <div className={styles.roleNodePermissions}>
                    <EmptyPermissionList />
                    {/* <PermissionList roleId={node.id} permissions={permissions} /> */}
                </div>
            ) : (
                <span />
            )}
        </li>
    )
}

const EmptyPermissionList: FC<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center m-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No permissions associated with this role.</div>
    </div>
)

// const PermissionList: FC<React.PropsWithChildren<{ roleId: number; permissions: PermissionMap }>> = ({
//     roleId,
//     permissions,
// }) => {
//     const rolePermissions = getRolePermissions(roleId)

//     if (rolePermissions.length === 0) {
//         return <EmptyPermissionList />
//     }

//     // const allDisplayNames = rolePermissions.map(rp => rp.displayName)
//     const permissionsDisplayMap: PermissionMap[keyof PermissionMap] = rolePermissions.reduce((acc, permission) => {
//         const { displayName } = permission
//         return { ...acc, [displayName]: true }
//     }, {})

//     const namespaces = Object.keys(permissions)
//     console.log(permissionsDisplayMap)

//     return (
//         <>
//             {namespaces.map(namespace => {
//                 const namespacePerms = permissions[namespace]
//                 const allNamespacePerms = Object.values(namespacePerms)
//                 return (
//                     <div key={namespace}>
//                         <Text>{namespace}</Text>
//                         <Grid columnCount={4}>
//                             {allNamespacePerms.map((ap, index) => {
//                                 const isChecked = Boolean(permissionsDisplayMap[ap.displayName])
//                                 return (
//                                     <Checkbox
//                                         label={ap.action}
//                                         id={ap.displayName}
//                                         key={ap.displayName}
//                                         defaultChecked={isChecked}
//                                     />
//                                 )
//                             })}
//                         </Grid>
//                     </div>
//                 )
//             })}
//             <Button variant="primary">Update</Button>
//         </>
//     )
// }
