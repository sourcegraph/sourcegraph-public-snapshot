import { useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, H3, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { CodeIntelIndexerFields, PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { DataSummary, DataSummaryItem } from '../components/DataSummary'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'

import styles from './RepoDashboardPage.module.scss'

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

type ActiveTab = 'list' | 'tree'

// TODO: Understand more
interface FilterState {
    failures: 'only' | 'show' | 'hide'
    suggestions: 'only' | 'show' | 'hide'
    allowedLanguageKeys: Set<string>
}

export const RepoDashboardPage: React.FunctionComponent<RepoDashboardPageProps> = ({ telemetryService, repo }) => {
    useEffect(() => {
        telemetryService.logPageView('CodeIntelRepoDashboard')
    }, [telemetryService])

    const location = useLocation()
    const queryParameters = new URLSearchParams(location.search)
    const activeTab: ActiveTab = queryParameters.get('view') === 'list' ? 'list' : 'tree'

    const { data, loading, error } = useRepoCodeIntelStatus({ variables: { repository: repo.name } })

    // TODO: Understand more
    const [filterState, setFilterState] = useState<FilterState>({
        failures: 'show',
        suggestions: 'show',
        allowedLanguageKeys: new Set([]),
    })

    // TODO: Understand more
    const shouldDisplayIndex = (index: PreciseIndexFields): boolean =>
        // Indexer key filter
        (filterState.allowedLanguageKeys.size === 0 || filterState.allowedLanguageKeys.has(getIndexerKey(index))) &&
        // Suggestion filter
        filterState.suggestions !== 'only' &&
        // Failure filter
        (filterState.failures === 'show' || (filterState.failures === 'only') === failureStates.has(index.state))

    // TODO: Understand more
    const shouldDisplayIndexerSuggestion = (indexer: CodeIntelIndexerFields): boolean =>
        // Indexer key filter
        (filterState.allowedLanguageKeys.size === 0 || filterState.allowedLanguageKeys.has(indexer.key)) &&
        // Suggestion filter
        filterState.suggestions !== 'hide' &&
        // Failure filter
        filterState.failures !== 'only'

    const indexes = useMemo(() => {
        if (!data) {
            return []
        }
        return data.recentActivity
    }, [data])

    console.log(indexes)

    const suggestedIndexers = useMemo(() => {
        if (!data) {
            return []
        }

        return data.availableIndexers
            .flatMap(({ roots, indexer }) => roots.map(root => ({ root, ...indexer })))
            .filter(
                ({ root, key }) =>
                    !indexes.some(index => getIndexRoot(index) === sanitizePath(root) && getIndexerKey(index) === key)
            )
    }, [data, indexes])

    const filteredRoots = useMemo(() => {
        const filteredIndexes = indexes.filter(shouldDisplayIndex)
        const filteredSuggestedIndexers = suggestedIndexers.filter(shouldDisplayIndexerSuggestion)
        const languageKeys = new Set([...indexes.map(getIndexerKey), ...suggestedIndexers.map(indexer => indexer.key)])

        return new Set([
            ...filteredIndexes.map(getIndexRoot),
            ...filteredSuggestedIndexers.map(indexer => indexer.root),
        ])
    }, [indexes, shouldDisplayIndex, shouldDisplayIndexerSuggestion, suggestedIndexers])

    const summaryItems = useMemo((): DataSummaryItem[] => {
        if (!indexes || !suggestedIndexers) {
            return []
        }

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
    }, [indexes, suggestedIndexers])

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

            <ul className="nav nav-tabs mt-2">
                {/* TODO Add tab roles etc */}
                <li className="nav-item">
                    <Link to="?view=tree" className={classNames('nav-link', activeTab === 'tree' && 'active')}>
                        Explore
                    </Link>
                </li>
                <li className="nav-item">
                    <Link to="?view=list" className={classNames('nav-link', activeTab === 'list' && 'active')}>
                        List
                    </Link>
                </li>
            </ul>

            <Container>
                {activeTab === 'list' && (
                    <>
                        <H3>List</H3>
                        <pre>{JSON.stringify(data.availableIndexers, null, 2)}</pre>
                    </>
                )}
                {activeTab === 'tree' && (
                    <>
                        <H3>Tree</H3>
                        <pre>{JSON.stringify(data.availableIndexers, null, 2)}</pre>
                    </>
                )}
            </Container>
        </>
    )
}
