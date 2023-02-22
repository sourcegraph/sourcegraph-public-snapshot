import { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Badge,
    Checkbox,
    Container,
    ErrorAlert,
    H3,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    Select,
    Tree,
} from '@sourcegraph/wildcard'

import { CodeIntelIndexerFields, PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { DataSummary, DataSummaryItem } from '../components/DataSummary'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'

import { buildTreeData, descendentNames } from '../components/tree/tree'
import { IndexStateBadge } from '../components/IndexStateBadge'
import { ConfigurationStateBadge, IndexerDescription } from '../components/ConfigurationStateBadge'
import { mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
import { byKey, groupBy, sanitizePath } from '../components/tree/util'

import styles from './RepoDashboardPage.module.scss'

export interface RepoDashboardPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo: { id: string; name: string }
    now?: () => Date
    // queryCommitGraph?: typeof defaultQueryCommitGraph
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

type InfoFilter = 'all' | 'errors' | 'suggestions'
type LanguageFilter = 'all' | string
interface FilterState {
    information: InfoFilter
    language: LanguageFilter
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
        information: 'all',
        language: 'all',
    })
    const applyFilter = (patch: Partial<FilterState>): void => setFilterState({ ...filterState, ...patch })

    const shouldDisplayIndex = useCallback(
        (index: PreciseIndexFields): boolean =>
            // Valid language filter
            (filterState.language === 'all' || filterState.language === getIndexerKey(index)) &&
            // Valid information filter
            (filterState.information === 'all' || filterState.information === 'errors') &&
            // Valid failure
            failureStates.has(index.state),
        [filterState]
    )

    const shouldDisplayIndexerSuggestion = useCallback(
        (indexer: CodeIntelIndexerFields): boolean =>
            // Valid language filter
            (filterState.language === 'all' || filterState.language === indexer.key) &&
            // Valid information filter
            (filterState.information === 'all' || filterState.information === 'suggestions'),
        [filterState]
    )

    const indexes = useMemo(() => {
        if (!data) {
            return []
        }
        return data.recentActivity
    }, [data])

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

    const filteredIndexes = indexes.filter(shouldDisplayIndex)
    const filteredSuggestedIndexers = suggestedIndexers.filter(shouldDisplayIndexerSuggestion)
    const languageKeys = new Set([...indexes.map(getIndexerKey), ...suggestedIndexers.map(indexer => indexer.key)])
    const filteredRoots = new Set([
        ...filteredIndexes.map(getIndexRoot),
        ...filteredSuggestedIndexers.map(indexer => indexer.root),
    ])

    const filteredTreeData = buildTreeData(filteredRoots)

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

            <ul className="nav nav-tabs mt-2 w-100">
                {/* TODO Add tab roles etc */}
                <li className="nav-item w-50">
                    <Link to="?view=tree" className={classNames('nav-link w-100', activeTab === 'tree' && 'active')}>
                        Explore
                    </Link>
                </li>
                <li className="nav-item w-50">
                    <Link to="?view=list" className={classNames('nav-link w-100', activeTab === 'list' && 'active')}>
                        Suggestions
                    </Link>
                </li>
            </ul>

            <Container>
                {activeTab === 'tree' && (
                    <>
                        <H3>Explore</H3>
                        <div className="d-flex justify-content-end">
                            <Select
                                id="info-filter"
                                label="Show:"
                                value={filterState.information}
                                onChange={event => applyFilter({ information: event.target.value as InfoFilter })}
                                className="d-flex align-items-center mr-3"
                                selectClassName={styles.select}
                                labelClassName="mb-0 mr-2"
                                isCustomStyle={true}
                            >
                                <option value="all">All</option>
                                <option value="errors">Errors</option>
                                <option value="suggestions">Suggestions</option>
                            </Select>

                            <Select
                                id="language-filter"
                                label="Language:"
                                value={filterState.information}
                                onChange={event => applyFilter({ language: event.target.value })}
                                className="d-flex align-items-center"
                                selectClassName={styles.select}
                                labelClassName="mb-0 mr-2"
                                isCustomStyle={true}
                            >
                                <option value="all">All</option>
                                {[...languageKeys].sort().map(key => (
                                    <option key={key} value={key}>
                                        {key}
                                    </option>
                                ))}
                            </Select>
                        </div>

                        {filteredTreeData.length > 0 ? (
                            <Tree
                                data={filteredTreeData}
                                defaultExpandedIds={filteredTreeData.map(element => element.id)}
                                propagateCollapse={true}
                                renderNode={({ element: { id, name: treeRoot, displayName }, ...props }) => {
                                    const descendentRoots = new Set(
                                        descendentNames(filteredTreeData, id).map(sanitizePath)
                                    )
                                    const filteredIndexesForRoot = filteredIndexes.filter(
                                        index => getIndexRoot(index) === treeRoot
                                    )
                                    const filteredIndexesForDescendents = filteredIndexes.filter(index =>
                                        descendentRoots.has(getIndexRoot(index))
                                    )
                                    const filteredSuggestedIndexersForRoot = filteredSuggestedIndexers.filter(
                                        ({ root }) => sanitizePath(root) === treeRoot
                                    )
                                    const filteredSuggestedIndexersForDescendents = filteredSuggestedIndexers.filter(
                                        ({ root }) => descendentRoots.has(sanitizePath(root))
                                    )

                                    return (
                                        <TreeNode
                                            displayName={displayName}
                                            indexesByIndexerNameForRoot={groupBy(filteredIndexesForRoot, getIndexerKey)}
                                            availableIndexersForRoot={filteredSuggestedIndexersForRoot}
                                            numDescendentErrors={
                                                filteredIndexesForDescendents.filter(index =>
                                                    failureStates.has(index.state)
                                                ).length
                                            }
                                            numDescendentConfigurable={
                                                filterState.information === 'all' ||
                                                filterState.information === 'suggestions'
                                                    ? filteredSuggestedIndexersForDescendents.length
                                                    : 0
                                            }
                                            {...props}
                                        />
                                    )
                                }}
                            />
                        ) : (
                            <>No code intel to display.</>
                        )}
                    </>
                )}
                {activeTab === 'list' && (
                    <>
                        <H3>Suggestions</H3>
                        <pre>{JSON.stringify(data.availableIndexers, null, 2)}</pre>
                    </>
                )}
            </Container>
        </>
    )
}

interface TreeNodeProps {
    displayName: string
    isBranch: boolean
    isExpanded: boolean
    indexesByIndexerNameForRoot: Map<string, PreciseIndexFields[]>
    availableIndexersForRoot: IndexerDescription[]
    numDescendentErrors: number
    numDescendentConfigurable: number
}

const TreeNode: React.FunctionComponent<TreeNodeProps> = ({
    displayName,
    isBranch,
    isExpanded,
    indexesByIndexerNameForRoot,
    availableIndexersForRoot,
    numDescendentErrors,
    numDescendentConfigurable,
}) => (
    <div className="w-100">
        <div className={classNames('d-inline', !isBranch ? styles.spacer : '')}>
            <Icon
                svgPath={isBranch && isExpanded ? mdiFolderOpenOutline : mdiFolderOutline}
                className={classNames('mr-1', styles.icon)}
                aria-hidden={true}
            />
            {displayName}
        </div>

        {[...indexesByIndexerNameForRoot?.entries()].sort(byKey).map(([indexerName, indexes]) => (
            <IndexStateBadge key={indexerName} indexes={indexes} />
        ))}

        {availableIndexersForRoot.map(indexer => (
            <ConfigurationStateBadge indexer={indexer} className="float-right ml-2" key={indexer.key} />
        ))}

        {isBranch && !isExpanded && (numDescendentConfigurable > 0 || numDescendentErrors > 0) && (
            <>
                {numDescendentConfigurable > 0 && (
                    <Badge variant="primary" className="ml-2" pill={true} small={true}>
                        {numDescendentConfigurable}
                    </Badge>
                )}
                {numDescendentErrors > 0 && (
                    <Badge variant="danger" className="ml-2" pill={true} small={true}>
                        {numDescendentErrors}
                    </Badge>
                )}
            </>
        )}
    </div>
)
