import React, { ChangeEvent, FC, useCallback, useEffect, useState } from 'react'

import { ApolloError } from '@apollo/client/errors'
import { mdiCancel, mdiClose, mdiMapSearch, mdiReload } from '@mdi/js'
import { noop } from 'lodash'

import { useMutation } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Alert,
    Button,
    Container,
    Icon,
    Input,
    Link,
    PageHeader,
    PageSwitcher,
    Select,
    useDebounce,
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionError, ConnectionLoading } from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import {
    CancelPermissionsSyncJobResult,
    CancelPermissionsSyncJobVariables,
    PermissionsSyncJob,
    PermissionsSyncJobReasonGroup,
    PermissionsSyncJobsResult,
    PermissionsSyncJobsSearchType,
    PermissionsSyncJobState,
    PermissionsSyncJobsVariables,
    ScheduleRepoPermissionsSyncResult,
    ScheduleRepoPermissionsSyncVariables,
    ScheduleUserPermissionsSyncResult,
    ScheduleUserPermissionsSyncVariables,
} from '../../graphql-operations'
import { useURLSyncedState } from '../../hooks'
import { IColumn, Table } from '../UserManagement/components/Table'

import {
    CANCEL_PERMISSIONS_SYNC_JOB,
    PERMISSIONS_SYNC_JOBS_QUERY,
    TRIGGER_REPO_SYNC,
    TRIGGER_USER_SYNC,
} from './backend'
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

interface Notification {
    text: React.ReactNode
    isError?: boolean
}

const DEFAULT_FILTERS = {
    reason: '',
    state: '',
    searchType: '',
    query: '',
}
const PERMISSIONS_SYNC_JOBS_POLL_INTERVAL = 5000

interface Props extends TelemetryProps {}

export const PermissionsSyncJobsTable: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('PermissionsSyncJobsTable')
    }, [telemetryService])

    const [filters, setFilters] = useURLSyncedState(DEFAULT_FILTERS)
    const debouncedQuery = useDebounce(filters.query, 300)

    const { connection, loading, startPolling, stopPolling, error, variables, ...paginationProps } =
        usePageSwitcherPagination<PermissionsSyncJobsResult, PermissionsSyncJobsVariables, PermissionsSyncJob>({
            query: PERMISSIONS_SYNC_JOBS_QUERY,
            variables: {
                first: 20,
                reasonGroup: stringToReason(filters.reason),
                state: stringToState(filters.state),
                searchType: stringToSearchType(filters.searchType),
                query: debouncedQuery,
            } as PermissionsSyncJobsVariables,
            getConnection: ({ data }) => data?.permissionsSyncJobs || undefined,
            options: { pollInterval: PERMISSIONS_SYNC_JOBS_POLL_INTERVAL },
        })

    const [polling, setPolling] = useState(true)
    const togglePolling = useCallback(() => {
        if (polling) {
            stopPolling()
        } else {
            startPolling(PERMISSIONS_SYNC_JOBS_POLL_INTERVAL)
        }
        setPolling(!polling)
    }, [polling, startPolling, stopPolling])

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

    const [notification, setNotification] = useState<Notification | undefined>(undefined)
    const dismissNotification = useCallback(() => setNotification(undefined), [])

    const [triggerUserSync] = useMutation<ScheduleUserPermissionsSyncResult, ScheduleUserPermissionsSyncVariables>(
        TRIGGER_USER_SYNC
    )
    const [triggerRepoSync] = useMutation<ScheduleRepoPermissionsSyncResult, ScheduleRepoPermissionsSyncVariables>(
        TRIGGER_REPO_SYNC
    )
    const [cancelSyncJob] = useMutation<CancelPermissionsSyncJobResult, CancelPermissionsSyncJobVariables>(
        CANCEL_PERMISSIONS_SYNC_JOB
    )

    const onError = (error: ApolloError): void => setNotification({ text: error.message, isError: true })

    const handleTriggerPermsSync = useCallback(
        ([job]: PermissionsSyncJob[]) => {
            if (job.subject.__typename === 'Repository') {
                triggerRepoSync({
                    variables: { repo: job.subject.id },
                    onCompleted: () => setNotification({ text: 'Repository permissions sync successfully scheduled' }),
                    onError,
                }).catch(
                    // noop here is used because an error is handled in `onError` option of `useMutation` above.
                    noop
                )
            } else {
                triggerUserSync({
                    variables: { user: job.subject.id },
                    onCompleted: () => setNotification({ text: 'User permissions sync successfully scheduled' }),
                    onError,
                }).catch(
                    // noop here is used because an error is handled in `onError` option of `useMutation` above.
                    noop
                )
            }
        },
        [triggerUserSync, triggerRepoSync]
    )

    const handleCancelSyncJob = useCallback(
        ([syncJob]: PermissionsSyncJob[]) => {
            cancelSyncJob({
                variables: { job: syncJob.id },
                onCompleted: ({ cancelPermissionsSyncJob }) =>
                    setNotification({ text: prettyPrintCancelSyncJobMessage(cancelPermissionsSyncJob || undefined) }),
                onError,
            }).catch(
                // noop here is used because an error is handled in `onError` option of `useMutation` above.
                noop
            )
        },
        [cancelSyncJob]
    )

    return (
        <div>
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
                actions={
                    <Button variant="secondary" onClick={togglePolling}>
                        {polling ? 'Pause polling' : 'Resume polling'}
                    </Button>
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
                {notification && (
                    <Alert
                        className="mt-2 d-flex justify-content-between align-items-center"
                        variant={notification.isError ? 'danger' : 'success'}
                    >
                        {notification.text}
                        <Button variant="secondary" outline={true} onClick={dismissNotification}>
                            <Icon aria-label="Close notification" svgPath={mdiClose} />
                        </Button>
                    </Alert>
                )}
                {connection?.nodes?.length === 0 && <EmptyList />}
                {!!connection?.nodes?.length && (
                    <Table<PermissionsSyncJob>
                        columns={TableColumns}
                        getRowId={node => node.id}
                        data={connection.nodes}
                        actions={[
                            {
                                key: 'Re-trigger job',
                                label: 'Re-trigger job',
                                icon: mdiReload,
                                onClick: handleTriggerPermsSync,
                                condition: ([node]) => finalState(node.state),
                            },
                            {
                                key: 'Cancel job',
                                label: 'Cancel job',
                                icon: mdiCancel,
                                onClick: handleCancelSyncJob,
                                condition: ([node]) => node.state === PermissionsSyncJobState.QUEUED,
                            },
                        ]}
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

const finalState = (state: PermissionsSyncJobState): boolean =>
    state !== PermissionsSyncJobState.QUEUED && state !== PermissionsSyncJobState.PROCESSING

const prettyPrintCancelSyncJobMessage = (message: string = 'Permissions sync job canceled.'): string =>
    message === 'No job that can be canceled found.'
        ? 'Permissions sync job is already dequeued and cannot be canceled.'
        : message
