import { FC, useState } from 'react'

import { mdiRefresh, mdiDelete, mdiDownload } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'
import LayersSearchOutlineIcon from 'mdi-react/LayersSearchOutlineIcon'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { SearchJobsOrderBy, SearchJobState } from '@sourcegraph/shared/src/graphql-operations'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    Badge,
    BadgeVariantType,
    Button,
    Container,
    ErrorAlert,
    FeedbackBadge,
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
    Select,
    Text,
    H2,
    Tooltip,
} from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { Page } from '../../components/Page'
import { ListPageZeroState } from '../../components/ZeroStates/ListPageZeroState'
import { SearchJobNode, SearchJobsResult, SearchJobsVariables } from '../../graphql-operations'

import { User, UsersPicker } from './UsersPicker'

import styles from './SearchJobsPage.module.scss'

const SEARCH_JOB_STATES = [
    SearchJobState.COMPLETED,
    SearchJobState.ERRORED,
    SearchJobState.FAILED,
    SearchJobState.QUEUED,
    SearchJobState.PROCESSING,
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
        $first: Int!
        $after: String
        $query: String!
        $states: [SearchJobState!]
        $orderBy: SearchJobsOrderBy
    ) {
        searchJobs(first: $first, after: $after, query: $query, states: $states, orderBy: $orderBy) {
            nodes {
                ...SearchJobNode
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }
`

export const SearchJobsPage: FC = props => {
    const [searchTerm, setSearchTerm] = useState<string>('')
    const [searchStateTerm, setSearchStateTerm] = useState('')
    const [selectedUsers, setUsers] = useState<User[]>([])
    const [selectedStates, setStates] = useState<SearchJobState[]>([])
    const [sortBy, setSortBy] = useState<SearchJobsOrderBy>(SearchJobsOrderBy.CREATED_DATE)

    const { connection, error, loading, fetchMore, hasNextPage } = useShowMorePagination<
        SearchJobsResult,
        SearchJobsVariables,
        SearchJobNode
    >({
        query: SEARCH_JOBS_QUERY,
        variables: {
            first: 5,
            after: null,
            query: searchStateTerm,
            states: selectedStates,
            orderBy: sortBy,
        },
        options: {
            // Comment out since it causes problem in storybook stories,
            // TODO Bring back polling interval as soon as BE is ready
            // pollInterval: 5000,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            return data?.searchJobs
        },
    })

    // Render only non-selected filters and filters that match with search term value
    const suggestions = SEARCH_JOB_STATES.filter(
        filter => !selectedStates.includes(filter) && filter.toLowerCase().includes(searchStateTerm.toLowerCase())
    )

    const searchJobs = connection?.nodes ?? []

    return (
        <Page>
            <PageHeader
                annotation={<FeedbackBadge status="experimental" feedback={{ mailto: 'support@sourcegraph.com' }} />}
                path={[{ icon: LayersSearchOutlineIcon, text: 'Search Jobs' }]}
                description={
                    <>
                        Run search queries over all repositories, branches, commit and revisions.{' '}
                        <Link to="">Learn more</Link> about search jobs.
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

                    <UsersPicker value={selectedUsers} onChange={setUsers} />

                    <Select
                        aria-label="Filter by search job status"
                        value={sortBy}
                        onChange={event => setSortBy(event.target.value as SearchJobsOrderBy)}
                        isCustomStyle={true}
                        className={styles.sort}
                        selectClassName={styles.sortSelect}
                    >
                        <option value={SearchJobsOrderBy.CREATED_DATE}>Sort by Created date</option>
                        <option value={SearchJobsOrderBy.QUERY}>Sort by Query</option>
                        <option value={SearchJobsOrderBy.STATE}>Sort by Status</option>
                    </Select>
                </header>

                {error && !loading && <ErrorAlert error={error} />}

                {loading && !connection && (
                    <Text>
                        <LoadingSpinner /> Fetching search jobs list
                    </Text>
                )}

                {connection && (
                    <ul className={styles.jobs}>
                        {connection.nodes.length === 0 && (
                            <SearchJobsZeroState
                                searchTerm={searchTerm}
                                selectedUsers={selectedUsers}
                                selectedStates={selectedStates}
                            />
                        )}

                        {connection.nodes.map(searchJob => (
                            <SearchJob key={searchJob.id} job={searchJob} />
                        ))}
                    </ul>
                )}

                {connection && connection.nodes.length > 0 && (
                    <footer className={styles.footer}>
                        {hasNextPage && (
                            <Button variant="secondary" outline={true} disabled={loading} onClick={fetchMore}>
                                Show more
                            </Button>
                        )}
                        <span className={styles.paginationInfo}>
                            {connection?.totalCount ?? 0} <b>search jobs</b> total{' '}
                            {hasNextPage && <>(showing first {searchJobs.length})</>}
                        </span>
                    </footer>
                )}
            </Container>
        </Page>
    )
}

interface SearchJobProps {
    job: SearchJobNode
}

const SearchJob: FC<SearchJobProps> = props => {
    const { job } = props
    const { repoStats } = job

    return (
        <li className={styles.job}>
            <span className={styles.jobStatus}>
                <span>{job.startedAt}</span>
                <SearchJobBadge job={job} />
            </span>

            <span className={styles.jobQuery}>
                {job.state !== SearchJobState.COMPLETED && (
                    <Text className="m-0 text-muted">
                        {repoStats.completed} out of {repoStats.total} repositories
                    </Text>
                )}

                <SyntaxHighlightedSearchQuery query={job.query} />
            </span>

            <span className={styles.jobCreator}>
                <UserAvatar user={job.creator!} />
                {job.creator?.displayName}
            </span>

            <span className={styles.jobActions}>
                <Button variant="link" className={styles.jobViewLogs}>
                    View logs
                </Button>

                <Tooltip content="Rerun search job">
                    <Button variant="secondary" outline={true} className={styles.jobSlimAction}>
                        <Icon svgPath={mdiRefresh} aria-hidden={true} />
                    </Button>
                </Tooltip>

                <Tooltip content="Delete search job">
                    <Button variant="danger" outline={true} className={styles.jobSlimAction}>
                        <Icon svgPath={mdiDelete} aria-hidden={true} />
                    </Button>
                </Tooltip>
            </span>

            <Button variant="secondary" className={styles.jobDownload}>
                <Icon svgPath={mdiDownload} aria-hidden={true} />
                Download
            </Button>
        </li>
    )
}

interface SearchJobBadgeProps {
    job: SearchJobNode
}

const SearchJobBadge: FC<SearchJobBadgeProps> = props => {
    const { job } = props

    if (job.state === SearchJobState.PROCESSING) {
        const totalRepo = job.repoStats.total
        const totalProcessedRepos = job.repoStats.completed

        return (
            <div className={styles.jobProgress}>
                <div
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ width: `${100 * (totalProcessedRepos / totalRepo)}%` }}
                    className={styles.jobProgressBar}
                />
            </div>
        )
    }

    return <Badge variant={getBadgeVariant(job.state)}>{job.state.toString()}</Badge>
}

const getBadgeVariant = (jobStatus: SearchJobState): BadgeVariantType => {
    switch (jobStatus) {
        case SearchJobState.COMPLETED:
            return 'success'
        case SearchJobState.QUEUED:
            return 'secondary'
        case SearchJobState.ERRORED:
            return 'warning'
        case SearchJobState.FAILED:
            return 'danger'
        case SearchJobState.PROCESSING:
            return 'primary'
    }
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
        title="No Search jobs found"
        subTitle="Try to reset filters to see all search jobs available to you"
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

                <Text>Learn more in the search jobs documentation page.</Text>
            </div>
        </div>
    )
}

const formatJobState = (state: SearchJobState): string => upperFirst(state.toLowerCase())
const hasFiltersValues = (states: SearchJobState[], users: User[], searchTerm: string): boolean =>
    states.length > 0 || users.length > 0 || searchTerm.trim().length > 0
