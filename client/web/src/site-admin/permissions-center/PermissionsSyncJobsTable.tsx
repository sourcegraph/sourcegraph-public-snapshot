import React, { ChangeEvent, FC, useCallback, useEffect } from 'react'

import { mdiMapSearch } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, Icon, Input, Link, PageHeader, PageSwitcher, Select, useDebounce } from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionError, ConnectionLoading } from '../../components/FilteredConnection/ui'
import {
    PermissionsSyncJob,
    PermissionsSyncJobReasonGroup,
    PermissionsSyncJobsResult,
    PermissionsSyncJobsSearchType,
    PermissionsSyncJobState,
    PermissionsSyncJobsVariables,
} from '../../graphql-operations'
import { useURLSyncedState } from '../../hooks'
import { IColumn, Table } from '../UserManagement/components/Table'

import { PERMISSIONS_SYNC_JOBS_QUERY } from './backend'
import {
    PermissionsSyncJobNumbers,
    PermissionsSyncJobReasonByline,
    PermissionsSyncJobStatusBadge,
    PermissionsSyncJobSubject,
} from './PermissionsSyncJobNode'

import styles from './PermissionsSyncJobsTable.module.scss'

interface Filters {
    reason: string
    state: string
    searchType: string
    query: string
}

const DEFAULT_FILTERS = {
    reason: '',
    state: '',
    searchType: '',
    query: '',
}

interface Props extends TelemetryProps {}

export const PermissionsSyncJobsTable: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('PermissionsSyncJobsTable')
    }, [telemetryService])

    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)
    const debouncedQuery = useDebounce(filters.query, 300)

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
            searchType: stringToSearchType(filters.searchType),
            query: debouncedQuery,
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
    const setSearchType = useCallback(
        (searchType: PermissionsSyncJobsSearchType | null) => setFilters({ searchType: searchType?.toString() || '' }),
        [setFilters]
    )

    return (
        <div>
            <PageHeader
                path={[{ text: 'Permissions Sync Dashboard' }]}
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
                {error && <ConnectionError errors={[error.message]} />}
                {!connection && <ConnectionLoading />}
                {connection?.nodes && (
                    <div className={styles.filtersGrid}>
                        <PermissionsSyncJobReasonGroupPicker value={filters.reason} onChange={setReason} />
                        <PermissionsSyncJobStatePicker value={filters.state} onChange={setState} />
                        <PermissionsSyncJobSearchTypePicker value={filters.searchType} onChange={setSearchType} />
                        <PermissionsSyncJobSearchPane filters={filters} setFilters={setFilters} />
                    </div>
                )}
                {connection?.nodes?.length === 0 && <EmptyList />}
                {!!connection?.nodes?.length && (
                    <Table<PermissionsSyncJob>
                        columns={TableColumns}
                        getRowId={node => node.id}
                        data={connection.nodes}
                    />
                )}
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

const TableColumns: IColumn<PermissionsSyncJob>[] = [
    {
        key: 'Status',
        header: 'Status',
        render: ({ state }: PermissionsSyncJob) => <PermissionsSyncJobStatusBadge state={state} />,
    },
    {
        key: 'Name',
        header: 'Name',
        render: (node: PermissionsSyncJob) => <PermissionsSyncJobSubject job={node} />,
    },
    {
        key: 'Reason',
        header: 'Reason',
        render: (node: PermissionsSyncJob) => <PermissionsSyncJobReasonByline job={node} />,
    },
    {
        key: 'Added',
        header: 'Added',
        render: (node: PermissionsSyncJob) => <PermissionsSyncJobNumbers job={node} added={true} />,
    },
    {
        key: 'Removed',
        header: 'Removed',
        render: (node: PermissionsSyncJob) => <PermissionsSyncJobNumbers job={node} added={false} />,
    },
    {
        key: 'Total',
        header: 'Total',
        render: ({ permissionsFound }: PermissionsSyncJob) => (
            <div className="text-secondary">
                <b>{permissionsFound}</b>
            </div>
        ),
    },
]

const stringToReason = (reason: string): PermissionsSyncJobReasonGroup | null =>
    reason === '' ? null : PermissionsSyncJobReasonGroup[reason as keyof typeof PermissionsSyncJobReasonGroup]

const stringToState = (state: string): PermissionsSyncJobState | null =>
    state === '' ? null : PermissionsSyncJobState[state as keyof typeof PermissionsSyncJobState]

const stringToSearchType = (searchType: string): PermissionsSyncJobsSearchType | null =>
    searchType === '' ? null : PermissionsSyncJobsSearchType[searchType as keyof typeof PermissionsSyncJobsSearchType]

interface PermissionsSyncJobReasonGroupPickerProps {
    value: string
    onChange: (reasonGroup: PermissionsSyncJobReasonGroup | null) => void
}

const PermissionsSyncJobReasonGroupPicker: FC<PermissionsSyncJobReasonGroupPickerProps> = props => {
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

const PermissionsSyncJobStatePicker: FC<PermissionsSyncJobStatePickerProps> = props => {
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

interface PermissionsSyncJobSearchTypePickerProps {
    value: string
    onChange: (searchType: PermissionsSyncJobsSearchType | null) => void
}

const PermissionsSyncJobSearchTypePicker: FC<PermissionsSyncJobSearchTypePickerProps> = props => {
    const { onChange, value } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value === '' ? null : (event.target.value as PermissionsSyncJobsSearchType)
        onChange(nextValue)
    }

    return (
        <Select id="searchTypeSelector" value={stringToSearchType(value) || ''} label="Search" onChange={handleSelect}>
            <option value="">Choose User/Repository</option>
            <option value={PermissionsSyncJobsSearchType.USER}>User</option>
            <option value={PermissionsSyncJobsSearchType.REPOSITORY}>Repository</option>
        </Select>
    )
}

interface PermissionsSyncJobSearchPaneProps {
    filters: Filters
    setFilters: (data: Partial<Filters>) => void
}

const PermissionsSyncJobSearchPane: FC<PermissionsSyncJobSearchPaneProps> = props => {
    const { filters, setFilters } = props

    return (
        <Input
            type="search"
            placeholder={
                filters.searchType === ''
                    ? 'Select a search context'
                    : filters.searchType === PermissionsSyncJobsSearchType.USER
                    ? 'Search users...'
                    : 'Search repositories...'
            }
            name="query"
            value={filters.query}
            onChange={event => setFilters({ ...filters, query: event.currentTarget.value })}
            autoComplete="off"
            autoCorrect="off"
            autoCapitalize="off"
            spellCheck={false}
            aria-label="Search sync jobs..."
            variant="regular"
            disabled={filters.searchType === ''}
            className={styles.searchInput}
        />
    )
}

const EmptyList: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className="text-muted text-center mb-3 w-100">
        <Icon className="icon" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <div className="pt-2">No permissions sync jobs have been found.</div>
    </div>
)
