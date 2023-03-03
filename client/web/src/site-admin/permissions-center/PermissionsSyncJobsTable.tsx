import React, { ChangeEvent, FC, useCallback, useEffect } from 'react'

import { mdiMapSearch } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, H5, Icon, PageSwitcher, Select } from '@sourcegraph/wildcard'

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
    const { connection, loading, error, variables, ...paginationProps } = usePageSwitcherPagination<
        PermissionsSyncJobsResult,
        PermissionsSyncJobsVariables,
        PermissionsSyncJob
    >({
        query: PERMISSIONS_SYNC_JOBS_QUERY,
        variables: {
            first: 20,
            reasonGroup: stringToReason(filters.reason),
            state: stringToState(filters.state),
        } as PermissionsSyncJobsVariables,
        getConnection: ({ data }) => data?.permissionsSyncJobs || undefined,
        options: { pollInterval: 5000 },
    })

    const setReason = useCallback(
        (reasonGroup: PermissionsSyncJobReasonGroup | null) => setFilters({ reason: reasonGroup?.toString() || '' }),
        [setFilters]
    )
    const setState = useCallback(
        (state: PermissionsSyncJobState | null) => setFilters({ state: state?.toString() || '' }),
        [setFilters]
    )

    return (
        <div>
            <Container>
                <ConnectionContainer>
                    {error && <ConnectionError errors={[error.message]} />}
                    {!connection && <ConnectionLoading />}
                    {connection?.nodes && (
                        <div className={styles.filtersGrid}>
                            <PermissionsSyncJobReasonGroupPicker value={filters.reason} onChange={setReason} />
                            <PermissionsSyncJobStatePicker value={filters.state} onChange={setState} />
                        </div>
                    )}
                    {connection?.nodes?.length === 0 && <EmptyList />}
                    {!!connection?.nodes?.length && (
                        <ConnectionList className={styles.jobsGrid} aria-label="Permissions sync jobs">
                            {connection?.nodes && <Header />}
                            {connection?.nodes?.map(node => (
                                <PermissionsSyncJobNode key={node.id} node={node} />
                            ))}
                        </ConnectionList>
                    )}
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

const stringToReason = (reason: string): PermissionsSyncJobReasonGroup | null =>
    reason === '' ? null : PermissionsSyncJobReasonGroup[reason as keyof typeof PermissionsSyncJobReasonGroup]

const stringToState = (state: string): PermissionsSyncJobState | null =>
    state === '' ? null : PermissionsSyncJobState[state as keyof typeof PermissionsSyncJobState]

interface PermissionsSyncJobReasonGroupPickerProps {
    value: string
    onChange: (reasonGroup: PermissionsSyncJobReasonGroup | null) => void
}

export const PermissionsSyncJobReasonGroupPicker: FC<PermissionsSyncJobReasonGroupPickerProps> = props => {
    const { onChange, value } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value === '' ? null : (event.target.value as PermissionsSyncJobReasonGroup)
        onChange(nextValue)
    }

    return (
        <Select id="reasonSelector" value={stringToReason(value) || ''} label="Reason" onChange={handleSelect}>
            <option value="">Any</option>
            <option value={PermissionsSyncJobReasonGroup.MANUAL}>Manual</option>
            <option value={PermissionsSyncJobReasonGroup.SCHEDULE}>Schedule</option>
            <option value={PermissionsSyncJobReasonGroup.SOURCEGRAPH}>Sourcegraph</option>
            <option value={PermissionsSyncJobReasonGroup.WEBHOOK}>Webhook</option>
        </Select>
    )
}

interface PermissionsSyncJobStatePickerProps {
    value: string
    onChange: (state: PermissionsSyncJobState | null) => void
}

export const PermissionsSyncJobStatePicker: FC<PermissionsSyncJobStatePickerProps> = props => {
    const { onChange, value } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value === '' ? null : (event.target.value as PermissionsSyncJobState)
        onChange(nextValue)
    }

    return (
        <Select id="stateSelector" value={stringToState(value) || ''} label="State" onChange={handleSelect}>
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

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No permissions sync jobs have been found.</div>
    </div>
)
