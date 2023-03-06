import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, AlertLink, Code, Container, H4, Link, PageHeader } from '@sourcegraph/wildcard'

import { ConnectionError, ConnectionLoading } from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import { DatabaseBackedPermsSyncFeatureFlagResult } from '../../graphql-operations'

import { DB_BACKED_PERMISSIONS_SYNC_FEATURE_FLAG_QUERY } from './backend'
import { PermissionsSyncJobsTable } from './PermissionsSyncJobsTable'

interface Props extends TelemetryProps {}

const FEATURE_FLAG_NAME = 'database-permission-sync-worker'

export const PermissionsSyncJobsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    const { data, loading, error } = useQuery<DatabaseBackedPermsSyncFeatureFlagResult>(
        DB_BACKED_PERMISSIONS_SYNC_FEATURE_FLAG_QUERY,
        {
            variables: {
                name: FEATURE_FLAG_NAME,
            },
        }
    )

    if (error) {
        return <ConnectionError errors={[error.message]} />
    }
    if (loading) {
        return <ConnectionLoading />
    }
    const enabled = (data?.featureFlag?.__typename === 'FeatureFlagBoolean' && data.featureFlag.value) || undefined

    return enabled ? (
        <PermissionsSyncJobsTable telemetryService={telemetryService} />
    ) : (
        <>
            <PageTitle title="Permissions Sync - Admin" />
            <PageHeader
                path={[{ text: 'Permissions Sync' }]}
                headingElement="h2"
                description={
                    <>
                        List of permissions sync jobs. Learn more about{' '}
                        <Link to="/help/admin/permissions/syncing">permissions syncing</Link>.
                    </>
                }
                className="mb-3"
            />
            <Container>
                <Alert variant="info" className="d-flex align-items-center">
                    <div className="flex-grow-1">
                        <H4>Database-backed permissions sync is disabled.</H4>
                        Please create and enable a <Code>{FEATURE_FLAG_NAME}</Code> feature flag to use this dashboard.
                    </div>
                    <AlertLink className="mr-2" to="/site-admin/feature-flags/configuration/new">
                        Create feature flag
                    </AlertLink>
                </Alert>
            </Container>
        </>
    )
}
