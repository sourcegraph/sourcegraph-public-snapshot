import { FunctionComponent, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiFolderOpenOutline, mdiFolderOutline } from '@mdi/js'
import classNames from 'classnames'
import { of } from 'rxjs'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
    Alert,
    Badge,
    Checkbox,
    Container,
    ErrorAlert,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    Tree,
    useObservable,
} from '@sourcegraph/wildcard'

import { CodeIntelIndexerFields, PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { ConfigurationStateBadge, IndexerDescription } from '../components/ConfigurationStateBadge'
import { IndexStateBadge } from '../components/IndexStateBadge'
import { queryCommitGraph as defaultQueryCommitGraph } from '../hooks/queryCommitGraph'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'
import { buildTreeData, descendentNames } from '../tree/tree'
import { byKey, groupBy, sanitizePath } from '../tree/util'

import styles from './RepoDashboardPage.module.scss'

export interface RepoDashboardPageProps {
    authenticatedUser: AuthenticatedUser | null
    repo: { id: string; name: string }
    now?: () => Date
    queryCommitGraph?: typeof defaultQueryCommitGraph
}

interface FilterState {
    failures: 'only' | 'show' | 'hide'
    suggestions: 'only' | 'show' | 'hide'
    allowedLanguageKeys: Set<string>
}

const completedStates = new Set<string>([PreciseIndexState.COMPLETED])
const failureStates = new Set<string>([PreciseIndexState.INDEXING_ERRORED, PreciseIndexState.PROCESSING_ERRORED])

export const RepoDashboardPage: FunctionComponent<RepoDashboardPageProps> = ({
    authenticatedUser,
    repo,
    now,
    queryCommitGraph = defaultQueryCommitGraph,
}) => {
    const apolloClient = useApolloClient()
    const { data, loading, error } = useRepoCodeIntelStatus({ variables: { repository: repo.name } })
    const commitGraphMetadata = useObservable(
        useMemo(
            () => (repo ? queryCommitGraph(repo.id, apolloClient) : of(undefined)),
            [repo, queryCommitGraph, apolloClient]
        )
    )

    //
    //

    const [filterState, setFilterState] = useState<FilterState>({
        failures: 'show',
        suggestions: 'show',
        allowedLanguageKeys: new Set([]),
    })
    const applyFilter = (patch: Partial<FilterState>): void => setFilterState({ ...filterState, ...patch })

    const setFailureFilter = (mode: 'only' | 'show' | 'hide'): void =>
        applyFilter({
            failures: mode,
            suggestions: mode === 'only' ? 'hide' : filterState.suggestions !== 'hide' ? 'show' : 'hide',
        })

    const setSuggestionFilter = (mode: 'only' | 'show' | 'hide'): void =>
        applyFilter({
            suggestions: mode,
            failures: mode === 'only' ? 'hide' : filterState.failures !== 'hide' ? 'show' : 'hide',
        })

    const toggleLanguage = (key: string): void =>
        applyFilter({
            allowedLanguageKeys: new Set(filterState.allowedLanguageKeys.has(key) ? [] : [key]),
        })

    const shouldDisplayIndex = (index: PreciseIndexFields): boolean =>
        // Indexer key filter
        (filterState.allowedLanguageKeys.size === 0 || filterState.allowedLanguageKeys.has(getIndexerKey(index))) &&
        // Suggestion filter
        filterState.suggestions !== 'only' &&
        // Failure filter
        (filterState.failures === 'show' || (filterState.failures === 'only') === failureStates.has(index.state))

    const shouldDisplayIndexerSuggestion = (indexer: CodeIntelIndexerFields): boolean =>
        // Indexer key filter
        (filterState.allowedLanguageKeys.size === 0 || filterState.allowedLanguageKeys.has(indexer.key)) &&
        // Suggestion filter
        filterState.suggestions !== 'hide' &&
        // Failure filter
        filterState.failures !== 'only'

    //
    //

    const indexes = data?.recentActivity || []
    const suggestedIndexers = (data?.availableIndexers || [])
        .flatMap(({ roots, indexer }) => roots.map(root => ({ root, ...indexer })))
        .filter(
            ({ root, key }) =>
                !indexes.some(index => getIndexRoot(index) === sanitizePath(root) && getIndexerKey(index) === key)
        )

    const filteredIndexes = indexes.filter(shouldDisplayIndex)
    const filteredSuggestedIndexers = suggestedIndexers.filter(shouldDisplayIndexerSuggestion)
    const languageKeys = new Set([...indexes.map(getIndexerKey), ...suggestedIndexers.map(indexer => indexer.key)])
    const filteredRoots = new Set([
        ...filteredIndexes.map(getIndexRoot),
        ...filteredSuggestedIndexers.map(indexer => indexer.root),
    ])
    const filteredTreeData = buildTreeData(filteredRoots)

    const numCompletedIndexes = indexes.filter(index => completedStates.has(index.state)).length
    const numFailedIndexes = indexes.filter(index => failureStates.has(index.state)).length
    const numUnconfiguredProjects = suggestedIndexers.length

    //
    //

    return loading ? (
        <LoadingSpinner />
    ) : error ? (
        <ErrorAlert prefix="Failed to load code intelligence summary for repository" error={error} />
    ) : data ? (
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

            {authenticatedUser?.siteAdmin && repo && (
                <Container className="mb-2">
                    View <Link to="/site-admin/code-graph/dashboard">dashboard</Link> for all repositories.
                </Container>
            )}

            <Container className="mb-2">
                <small className="d-block">
                    This repository was scanned for auto-indexing{' '}
                    {data.lastIndexScan ? <Timestamp date={data.lastIndexScan} /> : <>never</>}.
                </small>

                <small className="d-block">
                    The indexes of this repository were last considered for expiration{' '}
                    {data.lastUploadRetentionScan ? <Timestamp date={data.lastUploadRetentionScan} /> : <>never</>}.
                </small>
            </Container>

            {commitGraphMetadata && (
                <Alert variant={commitGraphMetadata.stale ? 'primary' : 'success'} aria-live="off">
                    {commitGraphMetadata.stale ? (
                        <>
                            Repository commit graph is currently stale and is queued to be refreshed. Refreshing the
                            commit graph updates which uploads are visible from which commits.
                        </>
                    ) : (
                        <>Repository commit graph is currently up to date.</>
                    )}{' '}
                    {commitGraphMetadata.updatedAt && (
                        <>
                            Last refreshed <Timestamp date={commitGraphMetadata.updatedAt} now={now} />.
                        </>
                    )}
                </Alert>
            )}

            <Container className="mb-2">
                {indexes.length + suggestedIndexers.length > 0 ? (
                    <>
                        <Container className="mb-2">
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline text-success">{numCompletedIndexes}</span>
                                <span className="text-muted ml-1">Successfully indexed projects</span>
                            </div>
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline text-danger">{numFailedIndexes}</span>
                                <span className="text-muted ml-1">Current failures</span>
                            </div>
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline">{numUnconfiguredProjects}</span>
                                <span className="text-muted ml-1">Unconfigured projects</span>
                            </div>
                        </Container>

                        <div className="mb-2">
                            <span className="float-left mr-2">
                                <Checkbox
                                    id="suggestions"
                                    label="Show suggestions"
                                    checked={filterState.suggestions === 'show'}
                                    onChange={event => setSuggestionFilter(event.target.checked ? 'show' : 'hide')}
                                />
                            </span>

                            <span className="float-left mr-2">
                                <Checkbox
                                    id="suggestions-2"
                                    label="Suggestions only"
                                    checked={filterState.suggestions === 'only'}
                                    onChange={event => setSuggestionFilter(event.target.checked ? 'only' : 'show')}
                                />
                            </span>

                            <span className="float-left mr-2">
                                <Checkbox
                                    id="failures"
                                    label="Show failures"
                                    checked={filterState.failures === 'show'}
                                    onChange={event => setFailureFilter(event.target.checked ? 'show' : 'hide')}
                                />
                            </span>

                            <span className="float-left mr-2">
                                <Checkbox
                                    id="failures-2"
                                    label="Failures only"
                                    checked={filterState.failures === 'only'}
                                    onChange={event => setFailureFilter(event.target.checked ? 'only' : 'show')}
                                />
                            </span>

                            {[...languageKeys].sort().map(key => (
                                <span className="float-right ml-2" key={key}>
                                    <div
                                        role="button"
                                        tabIndex={0}
                                        onClick={() => toggleLanguage(key)}
                                        onKeyDown={() => toggleLanguage(key)}
                                        className={filterState.allowedLanguageKeys.has(key) ? styles.selected : ''}
                                    >
                                        {key}
                                    </div>
                                </span>
                            ))}
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
                                                filterState.suggestions === 'hide'
                                                    ? 0
                                                    : filteredSuggestedIndexersForDescendents.length
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
                ) : (
                    <>No code intel available.</>
                )}
            </Container>
        </>
    ) : (
        <></>
    )
}

function getIndexRoot(index: PreciseIndexFields): string {
    return sanitizePath(index.projectRoot?.path || index.inputRoot)
}

function getIndexerKey(index: PreciseIndexFields): string {
    return index.indexer?.key || index.inputIndexer
}

//
//
//

interface TreeNodeProps {
    displayName: string
    isBranch: boolean
    isExpanded: boolean
    indexesByIndexerNameForRoot: Map<string, PreciseIndexFields[]>
    availableIndexersForRoot: IndexerDescription[]
    numDescendentErrors: number
    numDescendentConfigurable: number
}

const TreeNode: FunctionComponent<TreeNodeProps> = ({
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
