import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, H5, PageSwitcher } from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
} from '../../components/FilteredConnection/ui'
import { PermissionsSyncJob, PermissionsSyncJobsResult, PermissionsSyncJobsVariables } from '../../graphql-operations'

import { PERMISSIONS_SYNC_JOBS_QUERY } from './backend'
import { ChangesetCloseNode } from './PermissionsSyncJobsTableItem'

import styles from './PermissionsSyncJobsTable.module.scss'

interface Props extends TelemetryProps {}

export const PermissionsSyncJobsTable: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('PermissionsSyncJobsTable')
    }, [telemetryService])

    const { connection, loading, error, refetch, ...paginationProps } = usePageSwitcherPagination<
        PermissionsSyncJobsResult,
        PermissionsSyncJobsVariables,
        PermissionsSyncJob
    >({
        query: PERMISSIONS_SYNC_JOBS_QUERY,
        variables: {
            first: 20,
        },
        getConnection: ({ data }) => data?.permissionsSyncJobs || undefined,
        options: { pollInterval: 5000 },
    })

    return (
        <div>
            <Container>
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    <ConnectionList className={styles.jobsGrid} aria-label="Permissions sync jobs">
                        {connection && connection.nodes?.length > 0 && <Header />}
                        {connection?.nodes?.map(node => (
                            <ChangesetCloseNode key={node.id} node={node} />
                        ))}
                    </ConnectionList>
                </ConnectionContainer>
                <PageSwitcher
                    {...paginationProps}
                    className="mt-4"
                    totalCount={connection?.totalCount ?? null}
                    totalLabel="permissions sync jobs"
                />
            </Container>
        </div>
    )
}

const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <H5 className="text-uppercase">Status</H5>
        <H5 className="text-uppercase">Name</H5>
        <H5 className="text-uppercase">Reason</H5>
        <H5 className="text-uppercase">Added</H5>
        <H5 className="text-uppercase">Removed</H5>
        <H5 className="text-uppercase">Total</H5>
    </>
)
