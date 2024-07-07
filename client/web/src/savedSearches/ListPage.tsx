import { useEffect, useMemo, type FunctionComponent, type MutableRefObject } from 'react'

import { mdiLink, mdiMagnify } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'
import { useCallbackRef } from 'use-callback-ref'

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

import { buildFilterArgs, type Filter, type FilterValues } from '../components/FilteredConnection'
import {
    usePageSwitcherPagination,
    type PaginatedConnectionQueryArguments,
    type PaginationKeys,
} from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { useUrlSearchParamsForConnectionState } from '../components/FilteredConnection/hooks/useUrlSearchParamsForConnectionState'
import { ConnectionContainer, ConnectionForm } from '../components/FilteredConnection/ui'
import {
    SavedSearchesOrderBy,
    type SavedSearchFields,
    type SavedSearchesResult,
    type SavedSearchesVariables,
} from '../graphql-operations'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'
import { useNavbarQueryState } from '../stores'

import { savedSearchesQuery } from './graphql'
import { telemetryRecordSavedSearchViewSearchResults } from './telemetry'

import styles from './ListPage.module.scss'

const SavedSearchNode: FunctionComponent<
    SearchPatternTypeProps &
        TelemetryV2Props & {
            savedSearch: SavedSearchFields
            linkRef: MutableRefObject<HTMLAnchorElement | null> | null
        }
> = ({ savedSearch, patternType, linkRef, telemetryRecorder }) => (
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
            ref={linkRef}
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
        </Button>
        <div className="flex-1" />
        <Badge variant="outlineSecondary" tooltip="Owner">
            {('displayName' in savedSearch.owner ? savedSearch.owner.displayName : null) ??
                savedSearch.owner.namespaceName}
        </Badge>
        <Button to={savedSearch.url} variant="secondary" as={Link}>
            <Icon aria-label="Permalink" svgPath={mdiLink} />
        </Button>
        {savedSearch.viewerCanAdminister && (
            <Button to={`${savedSearch.url}/edit`} variant="secondary" as={Link}>
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

    const location = useLocation()

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const filters = useMemo<
        Filter<
            Exclude<keyof SavedSearchesVariables, PaginationKeys | 'query'>,
            Partial<Omit<SavedSearchesVariables, PaginationKeys | 'query'>>
        >[]
    >(
        () => [
            {
                label: 'Sort',
                type: 'select',
                id: 'orderBy',
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
                        label: (namespace.__typename === 'Org' && namespace.displayName) || namespace.namespaceName,
                        args: {
                            owner: namespace.id,
                        },
                    })) ?? []),
                ],
            },
        ],
        [namespaces]
    )

    type ConnectionStateParams = PaginatedConnectionQueryArguments & {
        orderBy: string
        owner: string
        query?: string
    }
    const connectionState = useUrlSearchParamsForConnectionState<
        ConnectionStateParams,
        Exclude<keyof SavedSearchesVariables, PaginationKeys | 'query'>
    >(filters)

    const debouncedQuery = useDebounce(connectionState.value.query, 300)
    const args = buildFilterArgs(filters, connectionState.value)
    const {
        connection,
        loading: listLoading,
        error: listError,
        ...paginationProps
    } = usePageSwitcherPagination<
        SavedSearchesResult,
        Partial<SavedSearchesVariables>,
        SavedSearchFields,
        ConnectionStateParams
    >({
        query: savedSearchesQuery,
        variables: { ...args, query: debouncedQuery },
        getConnection: ({ data }) => data?.savedSearches || undefined,
        state: connectionState,
    })

    const searchPatternType = useNavbarQueryState(state => state.searchPatternType)
    const callbackReference = useCallbackRef<HTMLAnchorElement>(null, ref => ref?.focus())

    const error = namespacesError || listError
    const loading = namespacesLoading || listLoading

    return (
        <>
            <Container>
                <ConnectionContainer>
                    <ConnectionForm
                        hideSearch={false}
                        showSearchFirst={true}
                        inputClassName="mw-30"
                        inputPlaceholder="Find a saved search..."
                        inputAriaLabel=""
                        inputValue={connectionState.value.query ?? ''}
                        onInputChange={event => {
                            connectionState.setValue(prev => ({ ...prev, query: event.target.value }))
                        }}
                        autoFocus={false}
                        filters={filters}
                        onFilterSelect={(filter, value) =>
                            connectionState.setValue(prev => ({ ...prev, [filter.id]: value }))
                        }
                        filterValues={
                            connectionState.value as FilterValues<
                                Exclude<keyof SavedSearchesVariables, PaginationKeys | 'query'>
                            >
                        }
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
                                    linkRef={
                                        location.state?.description === savedSearch.description
                                            ? callbackReference
                                            : null
                                    }
                                    patternType={searchPatternType}
                                    savedSearch={savedSearch}
                                    telemetryRecorder={telemetryRecorder}
                                />
                            ))}
                        </div>
                    )}
                </ConnectionContainer>
            </Container>
            <PageSwitcher
                {...paginationProps}
                className="mt-4"
                totalCount={connection?.totalCount ?? null}
                totalLabel="saved searches"
            />
        </>
    )
}
