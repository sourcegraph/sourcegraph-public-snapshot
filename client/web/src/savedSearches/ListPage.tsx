import { useEffect, useMemo, type FunctionComponent } from 'react'

import { mdiLink, mdiMagnify } from '@mdi/js'
import classNames from 'classnames'

import type { SearchPatternTypeProps } from '@sourcegraph/shared/src/search'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import {
    Badge,
    Button,
    Container,
    ErrorAlert,
    Icon,
    Link,
    LoadingSpinner,
    PageSwitcher,
    Text,
    useDebounce,
} from '@sourcegraph/wildcard'

import { buildFilterArgs, type Filter } from '../components/FilteredConnection'
import { useUrlSearchParamsForConnectionState } from '../components/FilteredConnection/hooks/connectionState'
import {
    usePageSwitcherPagination,
    type PaginationKeys,
} from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionForm } from '../components/FilteredConnection/ui'
import {
    SavedSearchesOrderBy,
    type SavedSearchFields,
    type SavedSearchesResult,
    type SavedSearchesVariables,
} from '../graphql-operations'
import { LibraryItemStatusBadge, LibraryItemVisibilityBadge } from '../library/itemBadges'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'
import { useNavbarQueryState } from '../stores'

import { savedSearchesQuery } from './graphql'
import { telemetryRecordSavedSearchViewSearchResults } from './telemetry'
import { urlToEditSavedSearch } from './util'

import styles from './ListPage.module.scss'

const SavedSearchNode: FunctionComponent<
    SearchPatternTypeProps &
        TelemetryV2Props & {
            savedSearch: SavedSearchFields
        }
> = ({ savedSearch, patternType, telemetryRecorder }) => (
    <div className={classNames(styles.row, 'list-group-item align-items-center flex-gap-2')}>
        <Button
            as={Link}
            to={`/search?${buildSearchURLQuery(savedSearch.query, patternType, false)}`}
            variant="link"
            size="lg"
            className={classNames(
                'd-flex flex-gap-2 align-items-center flex-grow-1 text-left text-decoration-none pl-0',
                styles.searchLink
            )}
            onClick={() => telemetryRecordSavedSearchViewSearchResults(telemetryRecorder, savedSearch, 'List')}
        >
            <Badge
                variant="primary"
                className="py-1 d-flex flex-gap-1 align-items-center mr-1 bg-transparent border border-primary text-primary"
            >
                <Icon aria-hidden={true} svgPath={mdiMagnify} className="flex-shrink-0" size="sm" />
                Run search
            </Badge>
            <span className={styles.searchLinkDescription}>{savedSearch.description}</span>
            <LibraryItemVisibilityBadge item={savedSearch} />
            <LibraryItemStatusBadge item={savedSearch} />
        </Button>
        <div className="flex-1" />
        <Badge variant="outlineSecondary" tooltip="Owner">
            {savedSearch.owner.namespaceName}
        </Badge>
        <Button to={savedSearch.url} variant="secondary" as={Link}>
            <Icon aria-label="Permalink" svgPath={mdiLink} />
        </Button>
        {savedSearch.viewerCanAdminister && (
            <Button to={urlToEditSavedSearch(savedSearch)} variant="secondary" as={Link}>
                Edit
            </Button>
        )}
    </div>
)

export function urlToSavedSearchesList(owner: SavedSearchFields['owner']['id']): string {
    return `${PageRoutes.SavedSearches}?owner=${encodeURIComponent(owner)}`
}

/**
 * List of saved searches.
 */
export const ListPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('savedSearches.list', 'view')
    }, [telemetryRecorder])

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const filters = useMemo<
        Filter<'drafts' | 'owner' | 'order', Partial<Omit<SavedSearchesVariables, PaginationKeys | 'query'>>>[]
    >(
        () => [
            {
                label: 'Show drafts',
                type: 'select',
                id: 'drafts',
                tooltip: 'Include draft saved searches',
                options: [
                    {
                        value: 'true',
                        label: 'Yes',
                        args: {
                            includeDrafts: true,
                        },
                    },
                    {
                        value: 'false',
                        label: 'No',
                        args: {
                            includeDrafts: false,
                        },
                    },
                ],
            },
            {
                label: 'Owner',
                type: 'select',
                id: 'owner',
                tooltip: 'User or organization that owns the saved search',
                options: [
                    {
                        value: 'all',
                        label: 'All',
                        args: {},
                    },
                    ...(namespaces?.map(namespace => ({
                        value: namespace.id,
                        label: namespace.namespaceName,
                        args: {
                            owner: namespace.id,
                        },
                    })) ?? []),
                ],
            },
            {
                label: 'Sort',
                type: 'select',
                id: 'order',
                options: [
                    {
                        value: 'updated-at-desc',
                        label: 'Recently updated',
                        args: {
                            orderBy: SavedSearchesOrderBy.SAVED_SEARCH_UPDATED_AT,
                        },
                    },
                    {
                        value: 'description-asc',
                        label: 'By description',
                        args: {
                            orderBy: SavedSearchesOrderBy.SAVED_SEARCH_DESCRIPTION,
                        },
                    },
                ],
            },
        ],
        [namespaces]
    )

    const [connectionState, setConnectionState] = useUrlSearchParamsForConnectionState(filters)
    const debouncedQuery = useDebounce(connectionState.query, 300)
    const {
        connection,
        loading: listLoading,
        error: listError,
        ...paginationProps
    } = usePageSwitcherPagination<
        SavedSearchesResult,
        Partial<SavedSearchesVariables>,
        SavedSearchFields,
        typeof connectionState
    >({
        query: savedSearchesQuery,
        variables: { ...buildFilterArgs(filters, connectionState), viewerIsAffiliated: true, query: debouncedQuery },
        getConnection: ({ data }) => data?.savedSearches || undefined,
        state: [connectionState, setConnectionState],
    })

    const searchPatternType = useNavbarQueryState(state => state.searchPatternType)

    const error = namespacesError || listError
    const loading = namespacesLoading || listLoading

    return (
        <>
            <Container data-testid="saved-searches-list-page">
                <ConnectionForm
                    hideSearch={false}
                    showSearchFirst={true}
                    inputClassName="mw-30"
                    inputPlaceholder="Find a saved search..."
                    inputAriaLabel=""
                    inputValue={connectionState.query}
                    onInputChange={event => {
                        setConnectionState(prev => ({ ...prev, query: event.target.value }))
                    }}
                    autoFocus={false}
                    filters={filters}
                    onFilterSelect={(filter, value) => setConnectionState(prev => ({ ...prev, [filter.id]: value }))}
                    filterValues={connectionState}
                    compact={false}
                    formClassName="flex-gap-4 mb-4"
                />
                {loading ? (
                    <LoadingSpinner />
                ) : error ? (
                    <ErrorAlert error={error} className="mb-3" />
                ) : !connection?.nodes || connection.nodes.length === 0 ? (
                    <Text className="text-center text-muted mb-0">No saved searches found.</Text>
                ) : (
                    <div className="list-group list-group-flush">
                        {connection.nodes.map(savedSearch => (
                            <SavedSearchNode
                                key={savedSearch.id}
                                patternType={searchPatternType}
                                savedSearch={savedSearch}
                                telemetryRecorder={telemetryRecorder}
                            />
                        ))}
                    </div>
                )}
            </Container>
            <PageSwitcher {...paginationProps} className="mt-4" totalCount={connection?.totalCount ?? null} />
        </>
    )
}
