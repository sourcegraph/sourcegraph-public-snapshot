import { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiClose, mdiFolderOpenOutline, mdiFolderOutline, mdiWrench } from '@mdi/js'
import classNames from 'classnames'
import { Location, useLocation, useNavigate } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Badge,
    Checkbox,
    Container,
    ErrorAlert,
    H3,
    Icon,
    Label,
    Link,
    LoadingSpinner,
    PageHeader,
    RadioButton,
    Select,
    Text,
    Tooltip,
    Tree,
} from '@sourcegraph/wildcard'

import { CodeIntelIndexerFields, PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { DataSummary, DataSummaryItem } from '../components/DataSummary'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'

import { buildTreeData, descendentNames } from '../components/tree/tree'
import { IndexStateBadge, IndexStateBadgeIcon } from '../components/IndexStateBadge'
import { ConfigurationStateBadge, IndexerDescription } from '../components/ConfigurationStateBadge'
import { byKey, getIndexerKey, groupBy, sanitizePath, getIndexRoot } from '../components/tree/util'

import styles from './RepoDashboardPage.module.scss'
import { INDEX_COMPLETED_STATES, INDEX_FAILURE_STATES } from '../constants'

export interface RepoDashboardPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo: { id: string; name: string }
    now?: () => Date
    // TODO: Query commit graph?
}

type ShowFilter = 'all' | 'indexes' | 'suggestions'
type IndexFilter = 'all' | 'success' | 'error'
type LanguageFilter = 'all' | string
interface DefaultFilterState {
    show: Extract<ShowFilter, 'all' | 'indexes'>
    indexState: IndexFilter
    language: LanguageFilter
}
interface SuggestionFilterState {
    show: Extract<ShowFilter, 'suggestions'>
    language: LanguageFilter
}
type FilterState = SuggestionFilterState | DefaultFilterState

/**
 * Build a valid FilterState
 * Used to allow other pages to easily link to a configured dashboard
 **/
export const buildParamsFromFilterState = (filterState: FilterState): URLSearchParams => {
    const params = new URLSearchParams()

    if (filterState.show === 'suggestions') {
        params.set('show', 'suggestions')
    } else {
        params.set('show', filterState.show)
        params.set('indexState', filterState.indexState)
    }

    params.set('language', filterState.language)

    return params
}

/**
 * Parse search parameters and build a valid FilterState.
 * Used to manage the state of the dashboard
 */
const buildFilterStateFromParams = ({ search }: Location): FilterState => {
    const queryParameters = new URLSearchParams(search)

    const show = queryParameters.get('show') || 'all'
    const language = queryParameters.get('language') || 'all'

    if (show === 'suggestions') {
        // Clean up URL
        queryParameters.delete('indexState')

        return {
            show,
            language,
        }
    }

    return {
        show: show as DefaultFilterState['show'],
        language,
        indexState: (queryParameters.get('indexState') || 'all') as IndexFilter,
    }
}

export const RepoDashboardPage: React.FunctionComponent<RepoDashboardPageProps> = ({ telemetryService, repo }) => {
    useEffect(() => {
        telemetryService.logPageView('CodeIntelRepoDashboard')
    }, [telemetryService])

    const location = useLocation()
    const navigate = useNavigate()

    const { data, loading, error } = useRepoCodeIntelStatus({ variables: { repository: repo.name } })

    const [filterState, setFilterState] = useState<FilterState>(buildFilterStateFromParams(location))

    useEffect(() => {
        setFilterState(buildFilterStateFromParams(location))
    }, [location])

    const handleFilterChange = useCallback(
        (value: string, paramKey: keyof SuggestionFilterState | keyof DefaultFilterState) => {
            const queryParameters = new URLSearchParams(location.search)
            queryParameters.set(paramKey, value)
            navigate({ search: queryParameters.toString() }, { replace: true })
        },
        [location.search, navigate]
    )

    const shouldDisplayIndex = useCallback(
        (index: PreciseIndexFields): boolean =>
            // Valid show filter
            (filterState.show === 'all' || filterState.show === 'indexes') &&
            // Valid language filter
            (filterState.language === 'all' || filterState.language === getIndexerKey(index)) &&
            // Valid indexState filter
            (filterState.indexState === 'all' ||
                (filterState.indexState === 'error' && INDEX_FAILURE_STATES.has(index.state)) ||
                (filterState.indexState === 'success' && INDEX_COMPLETED_STATES.has(index.state))),
        [filterState]
    )

    const shouldDisplayIndexerSuggestion = useCallback(
        (indexer: CodeIntelIndexerFields): boolean =>
            // Valid show filter
            (filterState.show === 'all' || filterState.show === 'suggestions') &&
            // Valid language filter
            (filterState.language === 'all' || filterState.language === indexer.key),
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

        const numCompletedIndexes = indexes.filter(index => INDEX_COMPLETED_STATES.has(index.state)).length
        const numFailedIndexes = indexes.filter(index => INDEX_FAILURE_STATES.has(index.state)).length
        const numUnconfiguredProjects = suggestedIndexers.length

        return [
            {
                label: 'Successfully indexed projects',
                value: numCompletedIndexes,
                valueClassName: 'text-success',
                to: `?${buildParamsFromFilterState({
                    show: 'indexes',
                    indexState: 'success',
                    language: 'all',
                }).toString()}`,
            },
            {
                label: 'Projects with errors',
                value: numFailedIndexes,
                className: styles.summaryItemThin,
                valueClassName: 'text-danger',
                to: `?${buildParamsFromFilterState({
                    show: 'indexes',
                    indexState: 'error',
                    language: 'all',
                }).toString()}`,
            },
            {
                label: 'Configurable projects',
                value: numUnconfiguredProjects,
                valueClassName: 'text-primary',
                to: `?${buildParamsFromFilterState({
                    show: 'suggestions',
                    language: 'all',
                }).toString()}`,
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
            <Container className="mt-3">
                <div className="d-flex justify-content-end">
                    <div className={styles.summaryContainer}>
                        <div>
                            <Label className={styles.radioGroup}>
                                Show:
                                <RadioButton
                                    name="show-filter"
                                    id="show-all"
                                    value="all"
                                    checked={filterState.show === 'all'}
                                    onChange={event => handleFilterChange(event.target.value, 'show')}
                                    label="All"
                                    wrapperClassName="ml-2 mr-3"
                                />
                                <RadioButton
                                    name="show-filter"
                                    id="show-indexes"
                                    value="indexes"
                                    checked={filterState.show === 'indexes'}
                                    onChange={event => handleFilterChange(event.target.value, 'show')}
                                    label="Indexes"
                                    wrapperClassName="mr-3"
                                />
                                <RadioButton
                                    name="show-filter"
                                    id="show-suggestions"
                                    value="suggestions"
                                    checked={filterState.show === 'suggestions'}
                                    onChange={event => handleFilterChange(event.target.value, 'show')}
                                    label="Suggestions"
                                />
                            </Label>
                        </div>
                        <div className="d-flex">
                            <Select
                                id="language-filter"
                                label="Language:"
                                value={filterState.language}
                                onChange={event => handleFilterChange(event.target.value, 'language')}
                                className="d-flex align-items-center mb-0"
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
                            {'indexState' in filterState && (
                                <Select
                                    id="index-filter"
                                    label="Indexing:"
                                    value={filterState.indexState}
                                    onChange={event => handleFilterChange(event.target.value, 'indexState')}
                                    className="d-flex align-items-center mb-0 ml-3"
                                    selectClassName={styles.select}
                                    labelClassName="mb-0 mr-2"
                                    isCustomStyle={true}
                                >
                                    <option value="all">Most recent attempt</option>
                                    <option value="success">Most recent success</option>
                                    <option value="error">Most recent failure</option>
                                </Select>
                            )}
                        </div>
                    </div>
                </div>

                {filteredTreeData.length > 1 ? (
                    <Tree
                        data={filteredTreeData}
                        defaultExpandedIds={filteredTreeData.map(element => element.id)}
                        propagateCollapse={true}
                        renderNode={({ element: { id, name: treeRoot, displayName }, handleExpand, ...props }) => {
                            const descendentRoots = new Set(descendentNames(filteredTreeData, id).map(sanitizePath))
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
                                    onClick={handleExpand}
                                    displayName={displayName}
                                    indexesByIndexerNameForRoot={groupBy(filteredIndexesForRoot, getIndexerKey)}
                                    availableIndexersForRoot={filteredSuggestedIndexersForRoot}
                                    numDescendentErrors={
                                        filteredIndexesForDescendents.filter(index =>
                                            INDEX_FAILURE_STATES.has(index.state)
                                        ).length
                                    }
                                    numDescendentConfigurable={
                                        filterState.show === 'all' || filterState.show === 'suggestions'
                                            ? filteredSuggestedIndexersForDescendents.length
                                            : 0
                                    }
                                    {...props}
                                />
                            )
                        }}
                    />
                ) : (
                    <Text className="text-muted">No data to display.</Text>
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
    onClick: (event: React.MouseEvent) => {}
}

const TreeNode: React.FunctionComponent<TreeNodeProps> = ({
    displayName,
    isBranch,
    isExpanded,
    indexesByIndexerNameForRoot,
    availableIndexersForRoot,
    numDescendentErrors,
    numDescendentConfigurable,
    onClick,
}) => (
    // We already handle accessibility events for expansion in the <TreeView />
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
    <div className={styles.treeNode} onClick={onClick}>
        <div className={classNames('d-inline', !isBranch ? styles.spacer : '')}>
            <Icon
                svgPath={isBranch && isExpanded ? mdiFolderOpenOutline : mdiFolderOutline}
                className={classNames('mr-1', styles.icon)}
                aria-hidden={true}
            />
            {displayName}
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

        <div className="d-flex align-items-center">
            {[...indexesByIndexerNameForRoot?.entries()].sort(byKey).map(([indexerName, indexes]) => (
                <IndexStateBadge
                    key={indexerName}
                    indexes={indexes}
                    className={classNames('text-muted', styles.badge)}
                />
            ))}

            {availableIndexersForRoot.map(indexer => (
                <Badge
                    as={Link}
                    to="../index-configuration"
                    variant="outlineSecondary"
                    key={indexer.key}
                    className={classNames('text-muted', styles.badge)}
                >
                    <Icon svgPath={mdiWrench} aria-hidden={true} className="mr-1 text-primary" />
                    Configure {indexer.key}
                </Badge>
            ))}
        </div>
    </div>
)
