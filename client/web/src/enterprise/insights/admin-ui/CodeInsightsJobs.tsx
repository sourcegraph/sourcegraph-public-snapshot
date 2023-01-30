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
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { PageTitle } from '../../../components/PageTitle'
import { GetCodeInsightsJobsResult, GetCodeInsightsJobsVariables, InsightJob } from '../../../graphql-operations'

import { GET_CODE_INSIGHTS_JOBS } from './query'

import styles from './CodeInsightsJobs.module.scss'

const JOBS: InsightJob[] = [
    {
        id: '001',
        insightViewTitle: 'Global css to Css modules migration',
        seriesLabel: 'Global css',
        seriesSearchQuery: 'file:*.scss -file:*.module.scss',
        backfillQueueStatus: {
            state: InsightQueueItemState.COMPLETED,
            queuePosition: 1,
            cost: 20,
            errors: [],
            percentComplete: 100,
            createdAt: '2023-01-10',
            startedAt: '2023-01-10',
            completedAt: '2023-01-10',
            runtime: '',
        },
    },
    {
        id: '002',
        insightViewTitle: 'Global css to Css modules migration',
        seriesLabel: 'Global css',
        seriesSearchQuery: 'file:*.scss -file:*.module.scss',
        backfillQueueStatus: {
            state: InsightQueueItemState.PROCESSING,
            queuePosition: 1,
            cost: 20,
            errors: [],
            percentComplete: 100,
            createdAt: '2023-01-10',
            startedAt: '2023-01-10',
            completedAt: '2023-01-10',
            runtime: '',
        },
    },
    {
        id: '003',
        insightViewTitle: 'Global css to Css modules migration',
        seriesLabel: 'Global css',
        seriesSearchQuery: 'file:*.scss -file:*.module.scss',
        backfillQueueStatus: {
            state: InsightQueueItemState.FAILED,
            queuePosition: 1,
            cost: 20,
            errors: ['Hello this is test message', 'Hello again test message 2'],
            percentComplete: 100,
            createdAt: '2023-01-10',
            startedAt: '2023-01-10',
            completedAt: '2023-01-10',
            runtime: '',
        },
    },
    {
        id: '004',
        insightViewTitle: 'Global css to Css modules migration',
        seriesLabel: 'Global css',
        seriesSearchQuery: 'file:*.scss -file:*.module.scss',
        backfillQueueStatus: {
            state: InsightQueueItemState.QUEUED,
            queuePosition: 1,
            cost: 20,
            errors: [],
            percentComplete: 100,
            createdAt: '2023-01-10',
            startedAt: '2023-01-10',
            completedAt: '2023-01-10',
            runtime: '',
        },
    },
]

export const CodeInsightsJobs: FC = props => {
    const [orderBy, setOrderBy] = useState<BackfillQueueOrderBy>(BackfillQueueOrderBy.STATE)
    const [selectedFilters, setFilters] = useState<InsightQueueItemState[]>([])
    const [search, setSearch] = useState<string>('')

    const { connection, loading, error, ...paginationProps } = usePageSwitcherPagination<
        GetCodeInsightsJobsResult,
        GetCodeInsightsJobsVariables,
        InsightJob
    >({
        query: GET_CODE_INSIGHTS_JOBS,
        variables: { orderBy, states: selectedFilters, search },
        getConnection: ({ data }) => data?.insightAdminBackfillQueue,
        options: { pollInterval: 5000, fetchPolicy: 'cache-and-network' },
    })

    return (
        <div>
            <PageTitle title="Code Insights jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code Insights jobs' }]}
                description="List of actionable Code Insights queued jobs"
                className="mb-3"
            />

            <header className={styles.header}>
                <CodeInsightsJobActions />

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
                    onChange={event => setSearch(event.target.value)}
                />
            </header>

            {loading && !connection && (
                <small className={styles.insightJobsMessage}>
                    <LoadingSpinner /> Loading code insights job
                </small>
            )}

            {/* { error && <ErrorAlert error={error}/> }*/}

            {connection && connection.nodes.length === 0 && (
                <span className={styles.insightJobsMessage}>
                    <Icon svgPath={mdiMapSearch} inline={false} aria-hidden={true} /> No code insight jobs yet. Enable
                    code insights and <Link to="/insights/create">create</Link> at least one insight to see its jobs
                    here.
                </span>
            )}

            {JOBS && (
                <ul className={styles.insightJobs}>
                    {JOBS.map(job => (
                        <CodeInsightJobCard key={job.id} job={job} />
                    ))}
                </ul>
            )}

            <PageSwitcher {...paginationProps} hasPreviousPage={true} className="mt-3" />
        </div>
    )
}

interface CodeInsightsJobActionsProps {
    className?: string
}

function CodeInsightsJobActions(props: CodeInsightsJobActionsProps): ReactElement {
    const { className } = props

    return (
        <div className={classNames(className, styles.actions)}>
            <Button variant="primary">Retry</Button>
            <Button variant="primary">Front of queue</Button>
            <Button variant="primary">Back of queue</Button>
        </div>
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
}

function CodeInsightJobCard(props: CodeInsightJobCardProps): ReactElement {
    const {
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
    } = props

    return (
        <li className={styles.insightJob}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input type="checkbox" className={styles.insightJobCheckbox} />
            <div className={styles.insightJobContent}>
                <header className={styles.insightJobHeader}>
                    <H3 className={styles.insightJobTitle}>{seriesLabel}</H3>
                    <Pill className={styles.insightJobSubtitle}>From {insightViewTitle} insight</Pill>
                </header>

                <span className="mt-1">
                    Series query: <Pill>{seriesSearchQuery}</Pill>
                    {percentComplete && (
                        <>
                            {', '} Сompleted by: {percentComplete}%
                        </>
                    )}
                    {errors && errors.length > 0 && (
                        <>
                            {', '} <InsightJobErrors errors={errors} />
                        </>
                    )}
                </span>

                <small className="mt-1 text-muted">
                    Queue position: {queuePosition}, Cost: {cost}, Created at: {createdAt}, Started at: {startedAt},
                    {completedAt && (
                        <>
                            {', '} Completed at: {completedAt}
                        </>
                    )}
                </small>
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
                    <ErrorAlert key={error} error={error} className="m-0" />
                ))}
            </PopoverContent>
            <PopoverTail size="sm" />
        </Popover>
    )
}
