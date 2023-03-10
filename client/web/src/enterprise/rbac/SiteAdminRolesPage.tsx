import React, { useEffect, useState, useCallback } from 'react'

import { mdiPlus } from '@mdi/js'
import { groupBy } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon, ProductStatusBadge, ErrorAlert, LoadingSpinner } from '@sourcegraph/wildcard'

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
        data,
        error: rolesError,
        loading: rolesLoading,
        refetch
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
        refetch()
    }, [closeModal, refetch])

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

            {error && <ErrorAlert error={error} />}
            {loading && <LoadingSpinner />}
            {!loading && data && (
                <ul className="list-unstyled">
                    {data.roles.nodes.map(node => (
                    <RoleNode
                        key={node.id}
                        node={node}
                        refetch={refetch}
                        allPermissions={permissions}
                    />
                ))}
                </ul>
            )}
        </div>
    )
}
