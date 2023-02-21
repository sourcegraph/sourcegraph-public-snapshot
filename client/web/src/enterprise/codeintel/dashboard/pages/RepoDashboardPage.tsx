import { useEffect, useMemo } from 'react'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { DataSummary, DataSummaryItem } from '../components/DataSummary'

import styles from './RepoDashboardPage.module.scss'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'
import { PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'

export interface RepoDashboardPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo: { id: string; name: string }
    now?: () => Date
    // queryCommitGraph?: typeof defaultQueryCommitGraph
}

const loading = false
const error = false

// Strip leading/trailing slashes and add a single leading slash
// TODO Understand more, possibly move out?
export function sanitizePath(root: string): string {
    return `/${root.replaceAll(/(^\/+)|(\/+$)/g, '')}`
}

// TODO: Understand more
function getIndexRoot(index: PreciseIndexFields): string {
    return sanitizePath(index.projectRoot?.path || index.inputRoot)
}

// TODO: Understand more
function getIndexerKey(index: PreciseIndexFields): string {
    return index.indexer?.key || index.inputIndexer
}

// TODO: Understand more
const completedStates = new Set<PreciseIndexState>([PreciseIndexState.COMPLETED])
const failureStates = new Set<PreciseIndexState>([
    PreciseIndexState.INDEXING_ERRORED,
    PreciseIndexState.PROCESSING_ERRORED,
])

export const RepoDashboardPage: React.FunctionComponent<RepoDashboardPageProps> = ({ telemetryService, repo }) => {
    useEffect(() => {
        telemetryService.logPageView('CodeIntelRepoDashboard')
    }, [telemetryService])

    const { data, loading, error } = useRepoCodeIntelStatus({ variables: { repository: repo.name } })

    const summaryItems = useMemo((): DataSummaryItem[] => {
        if (!data) {
            return []
        }

        const indexes = data.recentActivity

        const suggestedIndexers = data.availableIndexers
            .flatMap(({ roots, indexer }) => roots.map(root => ({ root, ...indexer })))
            .filter(
                ({ root, key }) =>
                    !indexes.some(index => getIndexRoot(index) === sanitizePath(root) && getIndexerKey(index) === key)
            )

        const numCompletedIndexes = indexes.filter(index => completedStates.has(index.state)).length
        const numFailedIndexes = indexes.filter(index => failureStates.has(index.state)).length
        const numUnconfiguredProjects = suggestedIndexers.length

        return [
            {
                label: 'Successfully indexed projects',
                value: numCompletedIndexes,
                valueClassName: 'text-success',
            },
            {
                label: 'Failing projects',
                value: numFailedIndexes,
                className: styles.summaryItemThin,
                valueClassName: 'text-danger',
            },
            {
                label: 'Unconfigured projects',
                value: numUnconfiguredProjects,
                valueClassName: 'text-merged',
            },
        ]
    }, [data])

    if (loading || !data) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert prefix="Failed to load code intelligence summary for repository" error={error} />
    }

    return (
        <>
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code intelligence summary for {repo.name}</>,
                    },
                ]}
                className="mb-3"
            />
            <Container>
                <DataSummary items={summaryItems} className="pb-3" />
            </Container>
        </>
    )
}
