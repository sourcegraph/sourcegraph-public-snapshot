import { type ChangeEvent, type FC, useEffect, useState } from 'react'

import { mdiMapSearch } from '@mdi/js'

import { BackfillQueueOrderBy, InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations'
import { type TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Container,
    ErrorAlert,
    Icon,
    Input,
    Link,
    LoadingSpinner,
    PageHeader,
    PageSwitcher,
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { PageTitle } from '../../../components/PageTitle'
import type { GetCodeInsightsJobsResult, GetCodeInsightsJobsVariables, InsightJob } from '../../../graphql-operations'

import { CodeInsightsJobsActions } from './components/job-actions'
import { CodeInsightsJobCard } from './components/job-card'
import { CodeInsightsJobsOrderPicker, CodeInsightsJobStatusPicker } from './components/job-filters'
import { GET_CODE_INSIGHTS_JOBS } from './query'

import styles from './CodeInsightsJobs.module.scss'

interface Props extends TelemetryV2Props {}

export const CodeInsightsJobs: FC<Props> = ({ telemetryRecorder }) => {
    const [search, setSearch] = useState<string>('')
    const [orderBy, setOrderBy] = useState<BackfillQueueOrderBy>(BackfillQueueOrderBy.QUEUE_POSITION)
    const [selectedJobs, setSelectedJobs] = useState<string[]>([])
    const [selectedFilters, setFilters] = useState<InsightQueueItemState[]>([
        InsightQueueItemState.PROCESSING,
        InsightQueueItemState.QUEUED,
    ])

    useEffect(() => telemetryRecorder.recordEvent('admin.codeInsightsJobs', 'view'), [telemetryRecorder])

    const { connection, loading, error, ...paginationProps } = usePageSwitcherPagination<
        GetCodeInsightsJobsResult,
        GetCodeInsightsJobsVariables,
        InsightJob
    >({
        query: GET_CODE_INSIGHTS_JOBS,
        variables: { orderBy, states: selectedFilters, search },
        getConnection: ({ data }) => data?.insightAdminBackfillQueue,
        options: { pollInterval: 10000 },
    })

    const handleJobSelect = (event: ChangeEvent<HTMLInputElement>, jobId: string): void => {
        if (event.target.checked) {
            setSelectedJobs(selectedJobs => [...selectedJobs, jobId])
        } else {
            setSelectedJobs(selectedJobs => selectedJobs.filter(id => id !== jobId))
        }
    }

    const handlePaginationClick = (): void => {
        // TODO: Put AppRouterContainer element in the global state to avoid bad querySelector calls
        const root = document.querySelector('[data-layout]')
        root?.scrollTo({ top: 0 })
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
                    <CodeInsightsJobsActions
                        selectedJobIds={selectedJobs}
                        className={styles.actions}
                        onSelectionClear={() => setSelectedJobs([])}
                    />

                    <CodeInsightsJobStatusPicker
                        selectedFilters={selectedFilters}
                        className={styles.statusFilter}
                        onFiltersChange={setFilters}
                    />

                    <CodeInsightsJobsOrderPicker
                        order={orderBy}
                        className={styles.orderBy}
                        onOrderChange={setOrderBy}
                    />

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
                            <CodeInsightsJobCard
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
                    onClick={handlePaginationClick}
                />
            </Container>
        </div>
    )
}
