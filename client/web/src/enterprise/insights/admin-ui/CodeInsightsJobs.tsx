import { ChangeEvent, FC, PropsWithChildren, ReactElement, useId, useState } from 'react'

import { mdiAlertCircle, mdiCheckCircle, mdiHelp, mdiMapSearch, mdiMoonNew, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'
import { identity, upperFirst } from 'lodash'

import { BackfillQueueOrderBy, InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations'
import {
    H3,
    H4,
    Label,
    Input,
    Select,
    Button,
    MultiCombobox,
    MultiComboboxEmptyList,
    MultiComboboxInput,
    MultiComboboxList,
    MultiComboboxOption,
    MultiComboboxPopover,
    PageHeader,
    LoadingSpinner,
    ErrorAlert,
    Icon,
    Link,
    Badge,
    Popover,
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    PageSwitcher,
    Container,
    ButtonProps,
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { PageTitle } from '../../../components/PageTitle'
import { GetCodeInsightsJobsResult, GetCodeInsightsJobsVariables, InsightJob } from '../../../graphql-operations'

import { GET_CODE_INSIGHTS_JOBS } from './query'

import styles from './CodeInsightsJobs.module.scss'

export const CodeInsightsJobs: FC = () => {
    const [orderBy, setOrderBy] = useState<BackfillQueueOrderBy>(BackfillQueueOrderBy.STATE)
    const [selectedFilters, setFilters] = useState<InsightQueueItemState[]>([])
    const [search, setSearch] = useState<string>('')
    const [selectedJobs, setSelectedJobs] = useState<string[]>([])

    const { connection, loading, error, ...paginationProps } = usePageSwitcherPagination<
        GetCodeInsightsJobsResult,
        GetCodeInsightsJobsVariables,
        InsightJob
    >({
        query: GET_CODE_INSIGHTS_JOBS,
        variables: { orderBy, states: selectedFilters, search },
        getConnection: ({ data }) => data?.insightAdminBackfillQueue,
        options: { pollInterval: 15000 },
    })

    const handleJobSelect = (event: ChangeEvent, jobId): void => {
        if (event.target.checked) {
            setSelectedJobs(selectedJobs => [...selectedJobs, jobId])
        } else {
            setSelectedJobs(selectedJobs => selectedJobs.filter(id => id !== jobId))
        }
    }

    return (
        <div>
            <PageTitle title="Code Insights jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code Insights jobs' }]}
                description="List of actionable Code Insights queued jobs"
                className="mb-3"
            />

            <Container className={styles.root}>
                <header className={styles.header}>
                    <CodeInsightsJobActions
                        selectedJobIds={selectedJobs}
                        onSelectionClear={() => setSelectedJobs([])}
                    />

                    <CodeInsightsFiltersPicker
                        selectedFilters={selectedFilters}
                        className={styles.statusFilter}
                        onFiltersChange={setFilters}
                    />

                    <CodeInsightsOrderPicker order={orderBy} className={styles.orderBy} onOrderChange={setOrderBy} />

                    <Input
                        placeholder="Search jobs by title or series label"
                        value={search}
                        className={styles.search}
                        status={loading ? 'loading' : 'initial'}
                        onChange={event => setSearch(event.target.value)}
                    />
                </header>

                {loading && !connection && (
                    <small className={styles.insightJobsMessage}>
                        <LoadingSpinner /> Loading code insights job
                    </small>
                )}

                {error && <ErrorAlert error={error} />}

                {connection && connection.nodes.length === 0 && !error && (
                    <span className={styles.insightJobsMessage}>
                        <Icon svgPath={mdiMapSearch} inline={false} aria-hidden={true} /> No code insight jobs yet.
                        Enable code insights and <Link to="/insights/create">create</Link> at least one insight to see
                        its jobs here.
                    </span>
                )}

                {connection && connection.nodes.length > 0 && (
                    <ul className={styles.insightJobs}>
                        {connection.nodes.map(job => (
                            <CodeInsightJobCard
                                key={job.id}
                                job={job}
                                selected={selectedJobs.includes(job.id)}
                                onSelectChange={event => handleJobSelect(event, job.id)}
                            />
                        ))}
                    </ul>
                )}

                <PageSwitcher
                    totalLabel="jobs"
                    totalCount={connection?.totalCount ?? null}
                    {...paginationProps}
                    className="mt-5"
                />
            </Container>
        </div>
    )
}

interface CodeInsightsJobActionsProps {
    selectedJobIds: string[]
    className?: string
    onSelectionClear: () => void
}

function CodeInsightsJobActions(props: CodeInsightsJobActionsProps): ReactElement {
    const { selectedJobIds, className, onSelectionClear } = props

    return (
        <div className={classNames(className, styles.actions)}>
            <JobActionButton actionCount={selectedJobIds.length}>Retry</JobActionButton>
            <JobActionButton actionCount={selectedJobIds.length}>Front of queue</JobActionButton>
            <JobActionButton actionCount={selectedJobIds.length}>Back of queue</JobActionButton>
            {selectedJobIds.length > 0 && (
                <Button variant="secondary" outline={true} onClick={onSelectionClear}>
                    Clear selection
                </Button>
            )}
        </div>
    )
}

interface JobActionButton extends ButtonProps {
    actionCount: number
}

function JobActionButton(props: JobActionButton): ReactElement {
    const { actionCount, children } = props

    return (
        <Button {...props} variant="primary" disabled={actionCount === 0} className={styles.actionButton}>
            {actionCount > 0 && <span className={styles.actionCount}>{actionCount}</span>}
            {children}
        </Button>
    )
}

interface CodeInsightsOrderPickerProps {
    order: BackfillQueueOrderBy
    className?: string
    onOrderChange: (order: BackfillQueueOrderBy) => void
}

function CodeInsightsOrderPicker(props: CodeInsightsOrderPickerProps): ReactElement {
    const { order, className, onOrderChange } = props

    const handleSelect = (event: ChangeEvent<HTMLSelectElement>): void => {
        const nextValue = event.target.value as BackfillQueueOrderBy

        onOrderChange(nextValue)
    }

    return (
        <Select
            id={useId()}
            label="Order jobs by:"
            value={order}
            isCustomStyle={true}
            className={classNames(className, 'm-0')}
            labelClassName="flex-shrink-0"
            onChange={handleSelect}
        >
            <option value={BackfillQueueOrderBy.STATE}>State</option>
            <option value={BackfillQueueOrderBy.QUEUE_POSITION}>Queue position</option>
        </Select>
    )
}

const CODE_INSIGHTS_JOBS_FILTERS = [
    InsightQueueItemState.COMPLETED,
    InsightQueueItemState.FAILED,
    InsightQueueItemState.NEW,
    InsightQueueItemState.PROCESSING,
    InsightQueueItemState.QUEUED,
    InsightQueueItemState.UNKNOWN,
]

const formatFilter = (filter: InsightQueueItemState): string => upperFirst(filter.toLowerCase())

interface CodeInsightsFiltersPickerProps {
    selectedFilters: InsightQueueItemState[]
    className?: string
    onFiltersChange: (filters: InsightQueueItemState[]) => void
}

function CodeInsightsFiltersPicker(props: CodeInsightsFiltersPickerProps): ReactElement {
    const { selectedFilters, className, onFiltersChange } = props

    const filterInputId = useId()
    const [filterInput, setFilterInput] = useState<string>('')

    // Render only non-selected filters and filters that match with search term value
    const suggestions = CODE_INSIGHTS_JOBS_FILTERS.filter(
        filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(filterInput.toLowerCase())
    )

    return (
        <Label htmlFor={filterInputId} className={className}>
            <H4 as={H3} className="mb-0">
                Status:
            </H4>

            <MultiCombobox
                selectedItems={selectedFilters}
                getItemName={formatFilter}
                getItemKey={identity}
                onSelectedItemsChange={onFiltersChange}
                className={styles.statusFilterField}
            >
                <MultiComboboxInput
                    id={filterInputId}
                    autoCorrect="false"
                    autoComplete="off"
                    placeholder="Select filter..."
                    value={filterInput}
                    onChange={event => setFilterInput(event.target.value)}
                />

                <MultiComboboxPopover>
                    <MultiComboboxList items={suggestions}>
                        {filters =>
                            filters.map((filter, index) => (
                                <MultiComboboxOption
                                    key={filter.toString()}
                                    value={formatFilter(filter)}
                                    index={index}
                                />
                            ))
                        }
                    </MultiComboboxList>

                    {suggestions.length === 0 && (
                        <MultiComboboxEmptyList>
                            {!filterInput ? <>All filters are selected</> : <>No options</>}
                        </MultiComboboxEmptyList>
                    )}
                </MultiComboboxPopover>
            </MultiCombobox>
        </Label>
    )
}

interface CodeInsightJobCardProps {
    job: InsightJob
    selected: boolean
    onSelectChange: (event: ChangeEvent<HTMLInputElement>) => void
}

function CodeInsightJobCard(props: CodeInsightJobCardProps): ReactElement {
    const {
        selected,
        job: {
            insightViewTitle,
            seriesLabel,
            seriesSearchQuery,
            backfillQueueStatus: {
                state,
                cost,
                errors,
                percentComplete,
                queuePosition,
                createdAt,
                startedAt,
                completedAt,
            },
        },
        onSelectChange,
    } = props

    const checkboxId = useId()

    const details = [
        queuePosition !== null && `Queue position: ${queuePosition}`,
        cost !== null && `Cost: ${cost}`,
        createdAt !== null && `Created at: ${createdAt}`,
        startedAt !== null && `Started at: ${startedAt}`,
        completedAt !== null && `Completed at: ${completedAt}`,
    ].filter(item => item)

    return (
        <li className={classNames(styles.insightJob, { [styles.insightJobActive]: selected })}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                id={checkboxId}
                type="checkbox"
                checked={selected}
                className={styles.insightJobCheckbox}
                onChange={onSelectChange}
            />
            <div className={styles.insightJobContent}>
                <header className={styles.insightJobHeader}>
                    <H3 as={Label} htmlFor={checkboxId} className={styles.insightJobTitle}>
                        {seriesLabel}
                    </H3>
                    <small className="text-muted">From</small>
                    <Pill className={styles.insightJobSubtitle}>{insightViewTitle} insight</Pill>
                </header>

                <span className={styles.insightJobMainInfo}>
                    {percentComplete !== null && <span>Сompleted by: {percentComplete}%</span>}
                    <span>
                        Series query: <Pill>{seriesSearchQuery}</Pill>
                        {errors && errors.length > 0 && (
                            <>
                                {', '} <InsightJobErrors errors={errors} />
                            </>
                        )}
                    </span>
                </span>

                {details.length > 0 && <small className="mt-1 text-muted">{details.join(', ')}</small>}
            </div>
            <div className={styles.insightJobState}>
                <InsightJobStatusIcon status={state} className={StatusClasses[state]} />
                {formatFilter(state)}
            </div>
        </li>
    )
}

const Pill: FC<PropsWithChildren<{ className?: string }>> = props => (
    <Badge {...props} as="small" variant="secondary" className={classNames(styles.pill, props.className)} />
)

const StatusIcon: Record<InsightQueueItemState, string> = {
    [InsightQueueItemState.COMPLETED]: mdiCheckCircle,
    [InsightQueueItemState.FAILED]: mdiAlertCircle,
    [InsightQueueItemState.NEW]: mdiMoonNew,
    [InsightQueueItemState.QUEUED]: mdiTimerSand,
    [InsightQueueItemState.UNKNOWN]: mdiHelp,
    [InsightQueueItemState.PROCESSING]: '',
}

const StatusClasses: Record<InsightQueueItemState, string> = {
    [InsightQueueItemState.COMPLETED]: styles.insightJobStateCompleted,
    [InsightQueueItemState.FAILED]: styles.insightJobStateErrored,
    [InsightQueueItemState.NEW]: styles.insightJobStateQueued,
    [InsightQueueItemState.QUEUED]: styles.insightJobStateQueued,
    [InsightQueueItemState.PROCESSING]: '',
    [InsightQueueItemState.UNKNOWN]: '',
}

interface InsightJobStatusProps {
    status: InsightQueueItemState
    className?: string
}

const InsightJobStatusIcon: FC<InsightJobStatusProps> = props => {
    const { status, className } = props

    if (status === InsightQueueItemState.PROCESSING) {
        return <LoadingSpinner inline={false} className={className} />
    }

    return (
        <Icon
            svgPath={StatusIcon[status]}
            width={20}
            height={20}
            inline={false}
            className={className}
            aria-hidden={true}
        />
    )
}

interface InsightJobErrorsProps {
    errors: string[]
}

const InsightJobErrors: FC<InsightJobErrorsProps> = props => {
    const { errors } = props

    return (
        <Popover>
            <PopoverTrigger as={Button} size="sm" outline={false} variant="danger" className={styles.errorsTrigger}>
                Errors
            </PopoverTrigger>
            <PopoverContent className={styles.errorsContent} focusLocked={false}>
                {errors.map(error => (
                    <ErrorAlert key={error} error={error} className={styles.error} />
                ))}
            </PopoverContent>
            <PopoverTail size="sm" />
        </Popover>
    )
}
