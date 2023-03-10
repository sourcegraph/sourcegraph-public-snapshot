import React, { useEffect, useState, useCallback } from 'react'

import { mdiPlus } from '@mdi/js'
import { groupBy } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon, ProductStatusBadge } from '@sourcegraph/wildcard'

import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'

import { useRolesConnection, usePermissions, PermissionsMap, DEFAULT_PAGE_LIMIT } from './backend'
import { CreateRoleModal } from './components/CreateRoleModal'
import { RoleNode } from './components/RoleNode'

import styles from './SiteAdminRolesPage.module.scss'

export interface SiteAdminRolesPageProps extends TelemetryProps {}

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])

    const [permissions, setPermissions] = useState<PermissionsMap>({} as PermissionsMap)

    // Fetch paginated roles.
    const {
        connection,
        error: rolesError,
        loading: rolesLoading,
        fetchMore,
        hasNextPage,
        refetchAll,
    } = useRolesConnection()
    // We need to query all permissions from the database, so site admins can update easily if they want to.
    const { error: permissionsError, loading: permissionsLoading } = usePermissions(result => {
        const permissions = groupBy(result.permissions.nodes, 'namespace')
        setPermissions(permissions as PermissionsMap)
    })

    const [showCreateRoleModal, setShowCreateRoleModal] = useState<boolean>(false)
    const openModal = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setShowCreateRoleModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setShowCreateRoleModal(false)
    }, [])

    const afterCreate = useCallback(() => {
        closeModal()
        refetchAll()
    }, [closeModal, refetchAll])

    const loading = rolesLoading || permissionsLoading
    const error = rolesError || permissionsError

    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                className={styles.rolesPageHeader}
                description="Roles represent a set of permissions that are granted to a user. Roles are currently only available for Batch Changes functionality."
                actions={
                    <Button variant="primary" onClick={openModal}>
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Create Role
                    </Button>
                }
            >
                <PageHeader.Heading as="h2">
                    <PageHeader.Breadcrumb>
                        Roles <ProductStatusBadge status="experimental" />
                    </PageHeader.Breadcrumb>
                </PageHeader.Heading>
            </PageHeader>

            {showCreateRoleModal && !loading && (
                <CreateRoleModal onCancel={closeModal} afterCreate={afterCreate} allPermissions={permissions} />
            )}

            <ConnectionContainer className="mb-3">
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                {connection && (
                    <>
                        <ConnectionList as="ul" className="list-group" aria-label="Roles">
                            {connection.nodes?.map(node => (
                                <RoleNode
                                    key={node.id}
                                    node={node}
                                    refetchAll={refetchAll}
                                    allPermissions={permissions}
                                />
                            ))}
                        </ConnectionList>
                        <SummaryContainer className="mt-2">
                            <ConnectionSummary
                                noSummaryIfAllNodesVisible={true}
                                first={DEFAULT_PAGE_LIMIT}
                                centered={true}
                                connection={connection}
                                noun="role"
                                pluralNoun="roles"
                                hasNextPage={hasNextPage}
                            />
                            {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                        </SummaryContainer>
                    </>
                )}
            </ConnectionContainer>
        </div>
    )
}
