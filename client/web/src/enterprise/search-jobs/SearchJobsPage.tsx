import { type FC, useEffect, useMemo, useState } from 'react'

import { mdiDelete, mdiDownload, mdiRefresh, mdiStop } from '@mdi/js'
import classNames from 'classnames'
import { timeFormat } from 'd3-time-format'
import { upperFirst } from 'lodash'
import LayersSearchOutlineIcon from 'mdi-react/LayersSearchOutlineIcon'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { SearchJobsOrderBy, SearchJobState } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    Button,
    Container,
    ErrorAlert,
    FeedbackBadge,
    H2,
    Icon,
    Input,
    Link,
    LoadingSpinner,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxPopover,
    PageHeader,
    PageSwitcher,
    Select,
    Text,
    Tooltip,
    useDebounce,
} from '@sourcegraph/wildcard'

import { DownloadFileButton } from '../../components/DownloadFileButton'
import { usePageSwitcherPagination } from '../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { ListPageZeroState } from '../../components/ZeroStates/ListPageZeroState'
import type { SearchJobNode, SearchJobsResult, SearchJobsVariables } from '../../graphql-operations'

import { SearchJobBadge } from './SearchJobBadge/SearchJobBadge'
import { CancelSearchJobModal, RerunSearchJobModal, SearchJobDeleteModal } from './SearchJobModal/SearchJobModal'
import { type User, UsersPicker } from './UsersPicker'

import styles from './SearchJobsPage.module.scss'

const SEARCH_JOB_STATES = [
    SearchJobState.COMPLETED,
    SearchJobState.ERRORED,
    SearchJobState.FAILED,
    SearchJobState.QUEUED,
    SearchJobState.PROCESSING,
    SearchJobState.CANCELED,
]

/**
 * Main query to fetch list of search job nodes, exported only for Storybook story
 * apollo mocks, not designed to be reused in other places.
 */
export const SEARCH_JOBS_QUERY = gql`
    fragment SearchJobNode on SearchJob {
        id
        query
        state
        URL
        logURL
        startedAt
        finishedAt
        repoStats {
            total
            completed
            failed
            inProgress
        }
        creator {
            id
            displayName
            username
            avatarURL
        }
    }

    query SearchJobs(
        $first: Int
        $after: String
        $last: Int
        $before: String
        $query: String!
        $userIDs: [ID!]
        $states: [SearchJobState!]
        $orderBy: SearchJobsOrderBy
    ) {
        searchJobs(
            first: $first
            after: $after
            last: $last
            before: $before
            query: $query
            userIDs: $userIDs
            states: $states
            orderBy: $orderBy
            descending: true
        ) {
            nodes {
                ...SearchJobNode
            }
            totalCount
            pageInfo {
                startCursor
                endCursor
                hasNextPage
                hasPreviousPage
            }
        }
    }
`

interface SearchJobsPageProps extends TelemetryProps {
    isAdmin: boolean
}

export const SearchJobsPage: FC<SearchJobsPageProps> = props => {
    const { isAdmin, telemetryService } = props

    const [searchTerm, setSearchTerm] = useState<string>('')
    const [searchStateTerm, setSearchStateTerm] = useState('')
    const [selectedUsers, setUsers] = useState<User[]>([])
    const [selectedStates, setStates] = useState<SearchJobState[]>([])
    const [sortBy, setSortBy] = useState<SearchJobsOrderBy>(SearchJobsOrderBy.CREATED_AT)

    const [jobToDelete, setJobToDelete] = useState<SearchJobNode | null>(null)
    const [jobToCancel, setJobToCancel] = useState<SearchJobNode | null>(null)
    const [jobToRestart, setJobToRestart] = useState<SearchJobNode | null>(null)

    const debouncedSearchTerm = useDebounce(searchTerm, 500)

    const { connection, error, loading, refetch, ...paginationProps } = usePageSwitcherPagination<
        SearchJobsResult,
        SearchJobsVariables,
        SearchJobNode
    >({
        query: SEARCH_JOBS_QUERY,
        variables: {
            query: debouncedSearchTerm,
            userIDs: selectedUsers.map(user => user.id),
            states: selectedStates,
            orderBy: sortBy,
        },
        options: {
            pollInterval: 5000,
            fetchPolicy: 'cache-and-network',
            pageSize: 15,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            return data?.searchJobs
        },
    })

    useEffect(() => {
        telemetryService.logViewEvent('SearchJobsListPage')
    }, [telemetryService])

    const handleSearchJobCreate = (): void => {
        setJobToRestart(null)
        refetch()
    }

    // Render only non-selected filters and filters that match with search term value
    const suggestions = SEARCH_JOB_STATES.filter(
        filter => !selectedStates.includes(filter) && filter.toLowerCase().includes(searchStateTerm.toLowerCase())
    )

    return (
        <Page>
            <PageTitle title="Search jobs" />
            <PageHeader
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                path={[{ icon: LayersSearchOutlineIcon, text: 'Search Jobs' }]}
                description={
                    <>
                        Manage Sourcegraph queries that have been run exhaustively to return all results.{' '}
                        <Link to="/help/code_search/how-to/search-jobs" target="_blank" rel="noopener noreferrer">
                            Learn more
                        </Link>{' '}
                        about search jobs.
                    </>
                }
            />

            <Container className="mt-4">
                <header className={styles.header}>
                    <Input
                        value={searchTerm}
                        placeholder="Search jobs by query..."
                        className={styles.search}
                        inputClassName={styles.searchInput}
                        onChange={event => setSearchTerm(event.target.value)}
                    />

                    <MultiCombobox
                        selectedItems={selectedStates}
                        getItemKey={formatJobState}
                        getItemName={formatJobState}
                        onSelectedItemsChange={setStates}
                        className={styles.filters}
                    >
                        <MultiComboboxInput
                            placeholder="Filter by search status..."
                            value={searchStateTerm}
                            autoCorrect="false"
                            autoComplete="off"
                            onChange={event => setSearchStateTerm(event.target.value)}
                        />

                        <MultiComboboxPopover>
                            <MultiComboboxList items={suggestions}>
                                {items =>
                                    items.map((item, index) => (
                                        <MultiComboboxOption
                                            key={formatJobState(item)}
                                            value={formatJobState(item)}
                                            index={index}
                                        />
                                    ))
                                }
                            </MultiComboboxList>
                        </MultiComboboxPopover>
                    </MultiCombobox>

                    {isAdmin && <UsersPicker value={selectedUsers} onChange={setUsers} />}

                    <Select
                        aria-label="Filter by search job status"
                        value={sortBy}
                        onChange={event => setSortBy(event.target.value as SearchJobsOrderBy)}
                        isCustomStyle={true}
                        className={styles.sort}
                        selectClassName={styles.sortSelect}
                    >
                        <option value={SearchJobsOrderBy.CREATED_AT}>Sort by Created date</option>
                        <option value={SearchJobsOrderBy.QUERY}>Sort by Query</option>
                        <option value={SearchJobsOrderBy.STATE}>Sort by Status</option>
                    </Select>
                </header>

                {error && !loading && <ErrorAlert error={error} className="mt-4 mb-0" />}

                {!error && loading && !connection && (
                    <div>
                        <LoadingSpinner /> Fetching search jobs list
                    </div>
                )}

                {!error && connection && (
                    <ul className={styles.jobs}>
                        {connection.nodes.length === 0 && (
                            <SearchJobsZeroState
                                searchTerm={searchTerm}
                                selectedUsers={selectedUsers}
                                selectedStates={selectedStates}
                            />
                        )}

                        {connection.nodes.map(searchJob => (
                            <SearchJob
                                key={searchJob.id}
                                job={searchJob}
                                withCreatorColumn={isAdmin}
                                telemetryService={telemetryService}
                                onRerun={setJobToRestart}
                                onCancel={setJobToCancel}
                                onDelete={setJobToDelete}
                            />
                        ))}
                    </ul>
                )}

                {!error && connection && connection.nodes.length > 0 && (
                    <footer className={styles.footer}>
                        <PageSwitcher
                            {...paginationProps}
                            className="mt-3"
                            totalCount={connection?.totalCount ?? null}
                            totalLabel="search jobs"
                        />
                    </footer>
                )}
            </Container>

            {jobToDelete && <SearchJobDeleteModal searchJob={jobToDelete} onDismiss={() => setJobToDelete(null)} />}
            {jobToRestart && <RerunSearchJobModal searchJob={jobToRestart} onDismiss={handleSearchJobCreate} />}
            {jobToCancel && <CancelSearchJobModal searchJob={jobToCancel} onDismiss={() => setJobToCancel(null)} />}
        </Page>
    )
}

const formatDate = timeFormat('%Y-%m-%d %H:%M:%S')
const formatDateSlim = timeFormat('%Y-%m-%d')

interface SearchJobProps extends TelemetryProps {
    job: SearchJobNode
    withCreatorColumn: boolean
    onRerun: (job: SearchJobNode) => void
    onCancel: (job: SearchJobNode) => void
    onDelete: (job: SearchJobNode) => void
}

const SearchJob: FC<SearchJobProps> = props => {
    const { job, withCreatorColumn, telemetryService, onRerun, onCancel, onDelete } = props
    const { repoStats } = job

    const startDate = useMemo(() => (job.startedAt ? formatDateSlim(new Date(job.startedAt)) : ''), [job.startedAt])
    const fullStartDate = useMemo(() => (job.startedAt ? formatDate(new Date(job.startedAt)) : ''), [job.startedAt])

    return (
        <li className={styles.job}>
            <span className={styles.jobStatus}>
                <Tooltip content={`Started at ${fullStartDate}`} placement="top">
                    <span>{startDate}</span>
                </Tooltip>
                <SearchJobBadge job={job} withProgress={true} />
            </span>

            <span className={styles.jobQuery}>
                {job.state !== SearchJobState.COMPLETED && (
                    <Text className="m-0 text-muted">
                        {repoStats.completed} out of {repoStats.total} tasks
                    </Text>
                )}

                <SyntaxHighlightedSearchQuery query={job.query} />
            </span>

            {withCreatorColumn && (
                <span className={styles.jobCreator}>
                    <UserAvatar user={job.creator!} className={styles.jobAvatar} />
                    {job.creator?.displayName ?? job.creator?.username ?? 'UNKNOWN'}
                </span>
            )}

            <Tooltip content={!job.logURL ? 'There are no logs yet' : ''}>
                <DownloadFileButton
                    variant="link"
                    disabled={!job.logURL}
                    fileUrl={job.logURL ?? ''}
                    debounceTime={1000}
                    className={styles.jobViewLogs}
                    onClick={() => {
                        telemetryService.log('SearchJobsResultViewLogsClick', {}, {})
                    }}
                >
                    View logs
                </DownloadFileButton>
            </Tooltip>

            <span className={styles.jobActions}>
                <Tooltip content="Rerun search job">
                    <Button
                        variant="secondary"
                        outline={true}
                        className={styles.jobSlimAction}
                        onClick={() => onRerun(job)}
                    >
                        <Icon svgPath={mdiRefresh} aria-hidden={true} />
                    </Button>
                </Tooltip>

                {job.state !== SearchJobState.FAILED &&
                    job.state !== SearchJobState.CANCELED &&
                    job.state !== SearchJobState.COMPLETED && (
                        <Tooltip content="Cancel search job">
                            <Button
                                variant="secondary"
                                outline={true}
                                className={styles.jobSlimAction}
                                onClick={() => onCancel(job)}
                            >
                                <Icon svgPath={mdiStop} aria-hidden={true} />
                            </Button>
                        </Tooltip>
                    )}

                <Tooltip content="Delete search job">
                    <Button
                        variant="danger"
                        outline={true}
                        className={styles.jobSlimAction}
                        onClick={() => onDelete(job)}
                    >
                        <Icon svgPath={mdiDelete} aria-hidden={true} />
                    </Button>
                </Tooltip>
            </span>

            <Tooltip content={!job.URL ? 'Results are not available yet' : ''}>
                <DownloadFileButton
                    fileUrl={job.URL ?? ''}
                    variant="secondary"
                    debounceTime={1000}
                    disabled={job.URL === null}
                    className={styles.jobDownload}
                    onClick={() => {
                        telemetryService.log('SearchJobsResultDownloadClick', {}, {})
                    }}
                >
                    <Icon svgPath={mdiDownload} aria-hidden={true} />
                    Download
                </DownloadFileButton>
            </Tooltip>
        </li>
    )
}

interface SearchJobsZeroStateProps {
    searchTerm: string
    selectedUsers: User[]
    selectedStates: SearchJobState[]
}

const SearchJobsZeroState: FC<SearchJobsZeroStateProps> = props => {
    const { searchTerm, selectedUsers, selectedStates } = props

    return hasFiltersValues(selectedStates, selectedUsers, searchTerm) ? (
        <SearchJobsWithFiltersZeroState />
    ) : (
        <SearchJobsInitialZeroState />
    )
}

const SearchJobsWithFiltersZeroState: FC = () => (
    <ListPageZeroState
        title="No search jobs found"
        subTitle="Reset filters to see all search jobs."
        withIllustration={false}
        className={styles.zeroStateWithFilters}
    />
)

interface SearchJobsInitialZeroStateProps {
    className?: string
}

const SearchJobsInitialZeroState: FC<SearchJobsInitialZeroStateProps> = props => {
    const isLightTheme = useIsLightTheme()
    const assetsRoot = window.context?.assetsRoot || ''

    return (
        <div className={classNames(props.className, styles.initialZeroState)}>
            <img
                alt="Search jobs creation button UI"
                width={384}
                height={267}
                src={`${assetsRoot}/img/no-jobs-state-${isLightTheme ? 'light' : 'dark'}.png`}
                className={styles.initialZeroStateImage}
            />
            <div className={styles.initialZeroStateText}>
                <H2 className={styles.initialZeroStateHeading}>No search jobs found</H2>

                <Text>
                    Search jobs are long running searches that will exhaustively return all results for widely scoped
                    queries.
                </Text>

                <Text>
                    You can trigger a search job from the results information panel when a normal search hits a result
                    limit.
                </Text>

                <Text>
                    Learn more in the search jobs{' '}
                    <Link to="/help/code_search/how-to/search-jobs" target="_blank" rel="noopener noreferrer">
                        documentation page
                    </Link>
                </Text>
            </div>
        </div>
    )
}

const formatJobState = (state: SearchJobState): string => upperFirst(state.toLowerCase())
const hasFiltersValues = (states: SearchJobState[], users: User[], searchTerm: string): boolean =>
    states.length > 0 || users.length > 0 || searchTerm.trim().length > 0
