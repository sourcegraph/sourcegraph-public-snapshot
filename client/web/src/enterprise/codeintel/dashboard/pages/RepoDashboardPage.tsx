import { useApolloClient } from '@apollo/client'
import { mdiCheck, mdiClose, mdiFolderOpenOutline, mdiFolderOutline, mdiTimerSand } from '@mdi/js'
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
    TreeNode as WTreeNode,
    useObservable,
} from '@sourcegraph/wildcard'
import classNames from 'classnames'
import { FunctionComponent, useMemo, useState } from 'react'
import { of } from 'rxjs'
import { PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { queryCommitGraph as defaultQueryCommitGraph } from '../hooks/queryCommitGraph'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'

import styles from './RepoDashboardPage.module.scss'

export interface RepoDashboardPageProps {
    authenticatedUser: AuthenticatedUser | null
    repo: { id: string; name: string }
    now?: () => Date
    queryCommitGraph?: typeof defaultQueryCommitGraph
}

interface FilterState {
    indexers: Set<string>
    failuresOnly: boolean
    hideSuggestions: boolean
}

const failureStates = new Set<string>([PreciseIndexState.INDEXING_ERRORED, PreciseIndexState.PROCESSING_ERRORED])

const terminalStates = new Set<string>([
    PreciseIndexState.COMPLETED,
    PreciseIndexState.DELETED,
    PreciseIndexState.DELETING,
    PreciseIndexState.INDEXING_ERRORED,
    PreciseIndexState.PROCESSING_ERRORED,
])

export const RepoDashboardPage: FunctionComponent<RepoDashboardPageProps> = ({
    authenticatedUser,
    repo,
    now,
    queryCommitGraph = defaultQueryCommitGraph,
}) => {
    const { data, loading, error } = useRepoCodeIntelStatus({ variables: { repository: repo.name } })

    const indexesByIndexerNameByRoot = new Map(
        [
            // Create an outer mapping from root to indexers
            ...groupBy(data?.recentActivity || [], index =>
                sanitize(index.projectRoot?.path || index.inputRoot)
            ).entries(),
        ].map(
            // Create a sub-mapping from indexer to indexes in each bucket
            ([root, indexes]) => [root, groupBy(indexes, index => index.indexer?.name || index.inputIndexer)]
        )
    )

    const availableIndexersByRoot = groupBy(
        // Flatten nested lists of roots grouped by indexers
        (data?.availableIndexers || []).flatMap(({ roots, ...rest }) =>
            roots
                // Filter out suggestions for which there already exists a record
                .filter(root => (indexesByIndexerNameByRoot.get(sanitize(root))?.get(rest.index) || []).length === 0)
                .map(root => ({ root, ...rest }))
        ),
        // Then re-group by root
        index => sanitize(index.root)
    )

    // Aggregates
    const namesFromRecords = [...indexesByIndexerNameByRoot.values()].flatMap(im => [...im.keys()])
    const namesFromInference = [...availableIndexersByRoot.values()].flatMap(is => is.map(i => i.index))
    const indexerNames = [...new Set([...namesFromRecords, ...namesFromInference])].sort()

    // count the number of unique indexes matching the given predicate
    const count = (f: (index: PreciseIndexFields) => boolean): number =>
        [...indexesByIndexerNameByRoot.values()].flatMap(indexesByIndexerName =>
            [...indexesByIndexerName.values()].filter(indexes => indexes.some(f))
        ).length
    const successCount = count(index => index.state === PreciseIndexState.COMPLETED)
    const failureCount = count(index => failureStates.has(index.state))
    const unconfiguredCount = [...availableIndexersByRoot.values()]
        .map(availableIndexers => availableIndexers.length)
        .reduce((sum, value) => sum + value, 0)

    // Construct tree data without any filters
    const unfilteredTreeData = buildTreeData(
        new Set([...indexesByIndexerNameByRoot.keys(), ...availableIndexersByRoot.keys()])
    )

    const [filterState, setFilterState] = useState<FilterState>({
        indexers: new Set([]),
        failuresOnly: false,
        hideSuggestions: false,
    })

    // Re-construct tree data using only roots containing filtered data
    const filteredTreeData = buildTreeData(
        new Set([
            // Filter out paths from precise index records we don't want to display
            ...[...indexesByIndexerNameByRoot.entries()]
                .filter(
                    ([_, indexesByIndexerName]) =>
                        // Show only paths with a matching indexer, if set
                        (filterState.indexers.size === 0 ||
                            [...indexesByIndexerName.keys()].some(indexerName =>
                                filterState.indexers.has(indexerName)
                            )) &&
                        // Show only paths with a precise index record in a failure state, if set
                        (!filterState.failuresOnly ||
                            [...indexesByIndexerName.values()].some(indexes =>
                                indexes.some(index => failureStates.has(index.state))
                            ))
                )
                .map(([root, _]) => root),

            // Filter out paths from suggestions we don't want to display
            ...[...availableIndexersByRoot.entries()]
                .filter(
                    ([_, indexers]) =>
                        // Show nothing if suggestions are hidden
                        !filterState.hideSuggestions &&
                        // Show only paths with a matching indexer, if set
                        (filterState.indexers.size === 0 ||
                            indexers.some(indexer => filterState.indexers.has(indexer.index)))
                )
                .map(([root, _]) => root),
        ])
    )

    const apolloClient = useApolloClient()
    const commitGraphMetadata = useObservable(
        useMemo(
            () => (repo ? queryCommitGraph(repo.id, apolloClient) : of(undefined)),
            [repo, queryCommitGraph, apolloClient]
        )
    )

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
                {unfilteredTreeData.length > 0 ? (
                    <>
                        <Container className="mb-2">
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline text-success">{successCount}</span>
                                <span className="text-muted ml-1">Successfully indexed projects</span>
                            </div>
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline text-danger">{failureCount}</span>
                                <span className="text-muted ml-1">Current failures</span>
                            </div>
                            <div className="d-inline p-4 m-4 b-2">
                                <span className="d-inline">{unconfiguredCount}</span>
                                <span className="text-muted ml-1">Unconfigured projects</span>
                            </div>
                        </Container>

                        <div className="mb-2">
                            <span className="float-left mr-2">
                                <Checkbox
                                    id="suggestions"
                                    label="Show suggestions"
                                    checked={!filterState.hideSuggestions}
                                    onChange={event =>
                                        setFilterState({ ...filterState, hideSuggestions: !event.target.checked })
                                    }
                                />
                            </span>

                            <span className="float-left mr-2">
                                <Checkbox
                                    id="failures"
                                    label="Failures only"
                                    checked={filterState.failuresOnly}
                                    onChange={event =>
                                        setFilterState({ ...filterState, failuresOnly: event.target.checked })
                                    }
                                />
                            </span>

                            {indexerNames.map(indexer => (
                                <span className="float-right ml-2">
                                    <Checkbox
                                        id={`"indexer-${indexer}"`}
                                        label={`Show only ${indexer}`}
                                        checked={filterState.indexers.has(indexer)}
                                        onChange={event =>
                                            setFilterState({
                                                ...filterState,
                                                indexers: new Set(event.target.checked ? [indexer] : []),
                                            })
                                        }
                                    />
                                </span>
                            ))}
                        </div>

                        {filteredTreeData.length > 0 ? (
                            <Tree
                                data={filteredTreeData}
                                defaultExpandedIds={filteredTreeData.map(element => element.id)}
                                propagateCollapse={true}
                                renderNode={({ element: { id, name, displayName }, ...props }) => {
                                    const indexesByIndexerNameForRoot = new Map(
                                        [...(indexesByIndexerNameByRoot.get(name)?.entries() || [])].filter(
                                            ([indexerName, indexes]) =>
                                                (filterState.indexers.size === 0 ||
                                                    filterState.indexers.has(indexerName)) &&
                                                (!filterState.failuresOnly ||
                                                    indexes.some(index => failureStates.has(index.state)))
                                        )
                                    )

                                    const availableIndexersForRoot = (availableIndexersByRoot.get(name) || []).filter(
                                        indexer =>
                                            !filterState.hideSuggestions &&
                                            (filterState.indexers.size === 0 || filterState.indexers.has(indexer.index))
                                    )

                                    // Calculate the number of failed + configurable roots _under_ this key.
                                    const descendentRoots = descendentNames(filteredTreeData, id)
                                    const numDescendentErrors = descendentRoots
                                        .flatMap(root =>
                                            [...(indexesByIndexerNameByRoot.get(root)?.values() || [])].flatMap(
                                                indexes => indexes.flatMap(index => failureStates.has(index.state))
                                            )
                                        )
                                        .reduce((a, b) => (b ? a + 1 : a), 0)
                                    const numDescendentConfigurable = !filterState.hideSuggestions
                                        ? descendentRoots
                                              .map(root => (availableIndexersByRoot.get(root) || []).length)
                                              .reduce((a, b) => a + b, 0)
                                        : 0

                                    return (
                                        <TreeNode
                                            name={name}
                                            displayName={displayName}
                                            filterState={filterState}
                                            indexesByIndexerNameForRoot={indexesByIndexerNameForRoot}
                                            availableIndexersForRoot={availableIndexersForRoot}
                                            numDescendentErrors={numDescendentErrors}
                                            numDescendentConfigurable={numDescendentConfigurable}
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

function descendentNames(treeData: TreeNodeWithDisplayName[], id: number): string[] {
    return [
        // children names
        ...treeData[id].children.map(id => treeData[id].name),
        // descendent names
        ...treeData[id].children.flatMap(id => descendentNames(treeData, id)),
    ]
}

interface TreeNodeProps {
    name: string
    displayName: string
    isBranch: boolean
    isExpanded: boolean
    filterState: FilterState
    indexesByIndexerNameForRoot: Map<string, PreciseIndexFields[]>
    availableIndexersForRoot: IndexerDescription[]
    numDescendentErrors: number
    numDescendentConfigurable: number
}

interface TreeNodeWithDisplayName extends WTreeNode {
    displayName: string
}

interface IndexerDescription {
    index: string
    url: string
    root: string
}

const TreeNode: FunctionComponent<TreeNodeProps> = ({
    name,
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
            <ConfigurationStateBadge indexer={indexer} />
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

interface IndexStateBadgeProps {
    indexes: PreciseIndexFields[]
}

const IndexStateBadge: FunctionComponent<IndexStateBadgeProps> = ({ indexes }) => {
    const terminalIndexes = indexes.filter(index => terminalStates.has(index.state))
    const firstNonTerminalIndexes = indexes
        .filter(index => !terminalStates.has(index.state))
        .sort((a, b) => new Date(a.uploadedAt ?? '').getDate() - new Date(b.uploadedAt ?? '').getDate())
        .slice(0, 1)

    // Only show one relevant terminal (assumed) index and one relevant non-terminal (explicit) index
    const collapsedIndexes = [...terminalIndexes, ...firstNonTerminalIndexes]

    return collapsedIndexes.length > 0 ? (
        <Link to={`./indexes/${collapsedIndexes[0].id}`}>
            <small className={classNames('float-right', 'ml-2', styles.hint)}>
                {collapsedIndexes.map(index => (
                    <IndexStateBadgeIcon index={index} />
                ))}
                {collapsedIndexes[0].indexer ? collapsedIndexes[0].indexer.name : collapsedIndexes[0].inputIndexer}
            </small>
        </Link>
    ) : (
        <></>
    )
}

interface IndexStateBadgeIconProps {
    index: PreciseIndexFields
}

const IndexStateBadgeIcon: FunctionComponent<IndexStateBadgeIconProps> = ({ index }) =>
    index.state === PreciseIndexState.COMPLETED ? (
        <Icon aria-hidden={true} svgPath={mdiCheck} className="text-success" />
    ) : index.state === PreciseIndexState.INDEXING ||
      index.state === PreciseIndexState.PROCESSING ||
      index.state === PreciseIndexState.UPLOADING_INDEX ? (
        <LoadingSpinner />
    ) : index.state === PreciseIndexState.QUEUED_FOR_INDEXING ||
      index.state === PreciseIndexState.QUEUED_FOR_PROCESSING ? (
        <Icon aria-hidden={true} svgPath={mdiTimerSand} />
    ) : index.state === PreciseIndexState.INDEXING_ERRORED || index.state === PreciseIndexState.PROCESSING_ERRORED ? (
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-danger" />
    ) : (
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-muted" />
    )

interface ConfigurationStateBadgeProps {
    indexer: IndexerDescription
}

const ConfigurationStateBadge: FunctionComponent<ConfigurationStateBadgeProps> = ({ indexer }) => (
    <small className={classNames('float-right', 'ml-2', styles.hint)}>
        <Icon aria-hidden={true} svgPath={mdiClose} className="text-muted" />{' '}
        <strong>Configure {indexer.index}?</strong>
    </small>
)

// Constructs an outline suitable for use with the wildcard <Tree /> component. This function constructs
// a file tree outline with a dummy root node (un-rendered) so that we can display explicit data for the
// root directory. We also attempt to collapse any runs of directories that have no data of its own to
// display and only one child.
function buildTreeData(dataPaths: Set<string>): TreeNodeWithDisplayName[] {
    // Construct a list of paths reachable from the given input paths by sanitizing the input path,
    // exploding the resulting path list into directory segments, constructing all prefixes of the
    // resulting segments, and deduplicating and sorting the result. This gives all ancestor paths
    // of the original input paths in lexicographic (NOTE: topological) order.
    const allPaths = [
        ...new Set(
            [...dataPaths]
                .map(root => sanitize(root).split('/'))
                .flatMap((segments: string[]): string[] => {
                    return segments.map((_, index) => sanitize(segments.slice(0, index + 1).join('/')))
                })
        ),
    ].sort()

    // Assign a stable set of identifiers for each of these paths. We start counting at one here due
    // to needing to have our indexes align with. See inline comments below for more detail.
    const treeIdsByPath = new Map(allPaths.map((name, index) => [name, index + 1]))

    // Build functions we can use to query which paths are direct parents and children of one another
    const { parentOf, childrenOf } = buildTreeQuerier(treeIdsByPath)

    // Build our list of tree nodes
    const nodes = [
        // The first is a synthetic fake node that isn't rendered
        buildNode(0, '', null, childrenOf(undefined)),
        // The remainder of the nodes come from our treeIds (which we started counting at one)
        ...[...treeIdsByPath.entries()]
            .sort(byKey)
            .map(([root, id]) => buildNode(id, root, parentOf(id), childrenOf(id))),
    ]

    // tryUnlink will attempt to unlink the give node from the list of nodes forming a tree.
    // Returns true if a node re-link occurred.
    const tryUnlink = (nodes: TreeNodeWithDisplayName[], nodeId: number): boolean => {
        const node = nodes[nodeId]
        if (nodeId === 0 || node.parent === null || node.children.length !== 1) {
            // Not a candidate - no  unique parent/child to re-link
            return false
        }
        if (node.displayName === '/') {
            // usability :comfy:
            return false
        }
        const parentId = node.parent
        const childId = node.children[0]

        // Link parent to child
        nodes[childId].parent = parentId
        // Replace replace node by child in parent
        nodes[parentId].children = nodes[parentId].children.map(id => (id === nodeId ? childId : id))
        // Move (prepend) text from node to child
        nodes[childId].displayName = nodes[nodeId].displayName + nodes[childId].displayName

        return true
    }

    const unlinkedIds = new Set(
        nodes
            // Attempt to unlink/collapse all paths that do not have data
            .filter(node => !dataPaths.has(node.name) && tryUnlink(nodes, node.id))
            // Return node for organ harvesting :screamcat:
            .map(node => node.id)
    )

    return (
        nodes
            // Remove each of the roots we've marked for skipping in the loop above
            .filter((_, index) => !unlinkedIds.has(index))
            // Remap each of the identifiers. We just collapse numbers so the sequence remains gap-less.
            // For some reason the wildcard <Tree /> component is a big stickler for having id and indexes align.
            .map(node => rewriteNodeIds(node, id => id - [...unlinkedIds].filter(v => v < id).length))
    )
}

interface TreeQuerier {
    parentOf: (id: number) => number
    childrenOf: (id: number | undefined) => number[]
}

// Return a pair of functions that can return the immediate parents and children of paths given tree identifiers.
function buildTreeQuerier(idsByPath: Map<string, number>): TreeQuerier {
    // Construct a map from identifiers of paths to the identifier of their immediate parent path
    const parentTreeIdByTreeId = new Map(
        [...idsByPath.entries()].map(([path, id]) => [
            id,
            [...idsByPath.keys()]
                // Filter out any non-ancestor directories
                // (NOTE: paths here guaranteed to start with slash)
                .filter(child =>
                    // Trim trailing slash and split each input (covers the `/` case)
                    checkSubset(child.replace(/(\/)$/, '').split('/'), path.replace(/(\/)$/, '').split('/'))
                )
                .sort((a, b) => b.length - a.length) // Sort by reverse length (most specific proper ancestor first)
                .map(key => idsByPath.get(key))[0], // Take the first element as its associated identifier
        ])
    )

    return {
        // Return parent identifier of entry (or zero if undefined)
        parentOf: id => parentTreeIdByTreeId.get(id) || 0,
        // Return identifiers of entries that declare their own parent as the target
        childrenOf: id => keysMatchingPredicate(parentTreeIdByTreeId, parentId => parentId === id),
    }
}

// Strip leading/trailing slashes and add a single leading slash
function sanitize(root: string): string {
    return `/${root.replaceAll(/(^\/+)|(\/+$)/g, '')}`
}

// Return true if the given slices for a proper (pairwise) subset < superset relation
function checkSubset(subset: string[], superset: string[]): boolean {
    return subset.length < superset.length && subset.filter((value, index) => value !== superset[index]).length === 0
}

// Compare two flattened Map<string, T> entries by key.
function byKey<T>(a: [string, T], b: [string, T]): number {
    return a[0].localeCompare(b[0])
}

// Group values together based on the given function
function groupBy<V, K>(values: V[], f: (value: V) => K): Map<K, V[]> {
    return values.reduce((acc, val) => acc.set(f(val), (acc.get(f(val)) || []).concat([val])), new Map<K, V[]>())
}

// Return the list of keys for the associated values for which the given predicate returned true.
function keysMatchingPredicate<K, V>(m: Map<K, V>, f: (value: V) => boolean): K[] {
    return [...m.entries()].filter(([_, v]) => f(v)).map(([k, _]) => k)
}

// Create a node with a default display name based on name (a filepath in this case)
function buildNode(id: number, name: string, parent: number | null, children: number[]): TreeNodeWithDisplayName {
    return { id, name, parent, children, displayName: `${name.split('/').reverse()[0]}/` }
}

// Rewrite the identifiers in each of the given tree node's fields.
function rewriteNodeIds({ id, parent, children, ...rest }: TreeNodeWithDisplayName, rewriteId: (id: number) => number) {
    return {
        id: rewriteId(id),
        parent: parent !== null ? rewriteId(parent) : null,
        children: children.map(rewriteId).sort(),
        ...rest,
    }
}
