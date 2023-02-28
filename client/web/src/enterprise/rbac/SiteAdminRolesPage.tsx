import React, { useEffect, useMemo } from 'react'

import { RouteComponentProps } from 'react-router'
import { mdiPlus } from '@mdi/js'
import { groupBy } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon } from '@sourcegraph/wildcard'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'

import { RoleNode } from './components/RoleNode'
import { useRolesConnection, usePermissions, PermissionsMap } from './backend'
import { PageTitle } from '../../components/PageTitle'

export interface SiteAdminRolesPageProps extends RouteComponentProps, TelemetryProps {}

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])

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
    const { data, error: permissionsError, loading: permissionsLoading } = usePermissions()

    const loading = rolesLoading || permissionsLoading
    const error = rolesError || permissionsError

    // We group permissions by namespace because that's how they're displayed in the UI. This allows
    // for quick lookup to know when a role is assigned a permission.
    const permissions = useMemo(() => {
        let result = {} as PermissionsMap
        if (permissionsLoading || permissionsError) {
            return result
        }

        const nodes = data?.permissions.nodes
        if (nodes && nodes.length > 0) {
            result = groupBy(nodes, 'namespace') as PermissionsMap
        }

        return result
    }, [data, permissionsLoading])

    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                path={[{ text: 'Roles' }]}
                headingElement="h2"
                description={
                    <>
                        Roles represent a set of permissions that are granted to a user.
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
                        <RoleNode key={node.id} node={node} afterDelete={refetchAll} allPermissions={permissions} />
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
