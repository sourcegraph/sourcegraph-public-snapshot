import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, AlertLink, Code, Container, H4, Link, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

import { PermissionsSyncJobsTable } from './PermissionsSyncJobsTable'

interface Props extends TelemetryProps {}

const FEATURE_FLAG_NAME = 'database-permission-sync-worker'

export const PermissionsSyncJobsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    const [enabled] = useFeatureFlag(FEATURE_FLAG_NAME, true)

    return enabled ? (
        <PermissionsSyncJobsTable telemetryService={telemetryService} />
    ) : (
        <>
            <PageTitle title="Permissions - Admin" />
            <PageHeader
                path={[{ text: 'Permissions' }]}
                headingElement="h2"
                description={
                    <>
                        List of permissions sync jobs. A permission sync job fetches the newest permissions for a given
                        repository or user from the respective code host. Learn more about{' '}
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
