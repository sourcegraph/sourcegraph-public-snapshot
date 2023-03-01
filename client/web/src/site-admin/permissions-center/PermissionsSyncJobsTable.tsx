import React, { ChangeEvent, FC, useCallback, useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, H5, PageSwitcher, Select } from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
} from '../../components/FilteredConnection/ui'
import {
    PermissionsSyncJob,
    PermissionsSyncJobReasonGroup,
    PermissionsSyncJobsResult,
    PermissionsSyncJobState,
    PermissionsSyncJobsVariables,
} from '../../graphql-operations'
import { useURLSyncedState } from '../../hooks'

import { PERMISSIONS_SYNC_JOBS_QUERY } from './backend'
import { PermissionsSyncJobNode } from './PermissionsSyncJobNode'

import styles from './PermissionsSyncJobsTable.module.scss'

const DEFAULT_FILTERS = {
    reason: '',
    state: '',
}

interface Props extends TelemetryProps {}

export const PermissionsSyncJobsTable: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('PermissionsSyncJobsTable')
    }, [telemetryService])

    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)
    const { connection, loading, error, refetch, ...paginationProps } = usePageSwitcherPagination<
        PermissionsSyncJobsResult,
        PermissionsSyncJobsVariables,
        PermissionsSyncJob
    >({
        query: PERMISSIONS_SYNC_JOBS_QUERY,
        variables: {
            first: 20,
            reasonGroup:
                filters.reason === ''
                    ? null
                    : PermissionsSyncJobReasonGroup[filters.reason as keyof typeof PermissionsSyncJobReasonGroup],
            state:
                filters.state === ''
                    ? null
                    : PermissionsSyncJobState[filters.state as keyof typeof PermissionsSyncJobState],
        } as PermissionsSyncJobsVariables,
        getConnection: ({ data }) => data?.permissionsSyncJobs || undefined,
        options: { pollInterval: 5000 },
    })

    const setReason = useCallback(
        (reasonGroup: PermissionsSyncJobReasonGroup | null) => setFilters({ reason: reasonGroup?.toString() ?? '' }),
        [setFilters]
    )
    const setState = useCallback(
        (state: PermissionsSyncJobState | null) => setFilters({ state: state?.toString() ?? '' }),
        [setFilters]
    )

    return (
        <div>
            <Container>
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {loading && !connection && <ConnectionLoading />}
                    {connection && connection.nodes?.length > 0 && (
                        <div className={styles.filtersGrid}>
                            <PermissionsSyncJobReasonGroupPicker onChange={setReason} />
                            <PermissionsSyncJobStatePicker onChange={setState} />
                        </div>
                    )}
                    <ConnectionList className={styles.jobsGrid} aria-label="Permissions sync jobs">
                        {connection && connection.nodes?.length > 0 && <Header />}
                        {connection?.nodes?.map(node => (
                            <PermissionsSyncJobNode key={node.id} node={node} />
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

interface PermissionsSyncJobReasonGroupPickerProps {
    onChange: (reasonGroup: PermissionsSyncJobReasonGroup | null) => void
}

export const PermissionsSyncJobReasonGroupPicker: FC<PermissionsSyncJobReasonGroupPickerProps> = props => {
    const { onChange } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value === '' ? null : (event.target.value as PermissionsSyncJobReasonGroup)
        onChange(nextValue)
    }

    return (
        <Select id="reasonSelector" label="Reason" onChange={handleSelect}>
            <option value="">Any</option>
            <option value={PermissionsSyncJobReasonGroup.MANUAL}>Manual</option>
            <option value={PermissionsSyncJobReasonGroup.SCHEDULE}>Schedule</option>
            <option value={PermissionsSyncJobReasonGroup.SOURCEGRAPH}>Sourcegraph</option>
            <option value={PermissionsSyncJobReasonGroup.WEBHOOK}>Webhook</option>
        </Select>
    )
}

interface PermissionsSyncJobStatePickerProps {
    onChange: (state: PermissionsSyncJobState | null) => void
}

export const PermissionsSyncJobStatePicker: FC<PermissionsSyncJobStatePickerProps> = props => {
    const { onChange } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value === '' ? null : (event.target.value as PermissionsSyncJobState)
        onChange(nextValue)
    }

    return (
        <Select id="stateSelector" label="State" onChange={handleSelect}>
            <option value="">Any</option>
            <option value={PermissionsSyncJobState.CANCELED}>Canceled</option>
            <option value={PermissionsSyncJobState.COMPLETED}>Completed</option>
            <option value={PermissionsSyncJobState.ERRORED}>Errored</option>
            <option value={PermissionsSyncJobState.FAILED}>Failed</option>
            <option value={PermissionsSyncJobState.PROCESSING}>Processing</option>
            <option value={PermissionsSyncJobState.QUEUED}>Queued</option>
        </Select>
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
