import { ChangeEvent, FC, useState } from 'react'

import { mdiMapSearch } from '@mdi/js'

import { BackfillQueueOrderBy, InsightQueueItemState } from '@sourcegraph/shared/src/graphql-operations'
import {
    Input,
    PageHeader,
    LoadingSpinner,
    ErrorAlert,
    Icon,
    Link,
    PageSwitcher,
    Container,
} from '@sourcegraph/wildcard'

import { usePageSwitcherPagination } from '../../../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { PageTitle } from '../../../components/PageTitle'
import { GetCodeInsightsJobsResult, GetCodeInsightsJobsVariables, InsightJob } from '../../../graphql-operations'

import { CodeInsightsJobsActions } from './components/job-actions'
import { CodeInsightsJobCard } from './components/job-card'
import { CodeInsightsJobStatusPicker, CodeInsightsJobsOrderPicker } from './components/job-filters'
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
        options: { pollInterval: 5000 },
    })

    const handleJobSelect = (event: ChangeEvent<HTMLInputElement>, jobId: string): void => {
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
                />
            </Container>
        </div>
    )
}
