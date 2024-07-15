import React, { useEffect, useMemo } from 'react'

import { useLocation } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import { Container, ErrorAlert, Input, LoadingSpinner, PageSwitcher, useDebounce } from '@sourcegraph/wildcard'

import { EXTERNAL_SERVICE_IDS_AND_NAMES } from '../components/externalServices/backend'
import { buildFilterArgs, FilterControl, type Filter, type FilterOption } from '../components/FilteredConnection'
import { useUrlSearchParamsForConnectionState } from '../components/FilteredConnection/hooks/connectionState'
import { usePageSwitcherPagination } from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import {
    RepositoryOrderBy,
    type ExternalServiceIDsAndNamesResult,
    type ExternalServiceIDsAndNamesVariables,
    type RepositoriesResult,
    type RepositoriesVariables,
    type SiteAdminRepositoryFields,
    type StatusAndRepoStatsResult,
} from '../graphql-operations'
import { PageRoutes } from '../routes.constants'

import { ValueLegendList, type ValueLegendListProps } from './analytics/components/ValueLegendList'
import { REPO_PAGE_POLL_INTERVAL, REPOSITORIES_QUERY, STATUS_AND_REPO_STATS } from './backend'
import { RepositoryNode } from './RepositoryNode'

import styles from './SiteAdminRepositoriesContainer.module.scss'

const STATUS_FILTERS: { [label: string]: FilterOption } = {
    All: {
        label: 'All',
        value: 'all',
        tooltip: 'Show all repositories',
        args: {},
    },
    Cloned: {
        label: 'Cloned',
        value: 'cloned',
        tooltip: 'Show cloned repositories only',
        args: { cloneStatus: 'CLONED' },
    },
    Cloning: {
        label: 'Cloning',
        value: 'cloning',
        tooltip: 'Show repositories currently being cloned only',
        args: { cloneStatus: 'CLONING' },
    },
    NotCloned: {
        label: 'Not cloned',
        value: 'not-cloned',
        tooltip: 'Show only repositories that have not been cloned yet',
        args: { cloneStatus: 'NOT_CLONED' },
    },
    Indexed: {
        label: 'Indexed',
        value: 'indexed',
        tooltip: 'Show only repositories that have already been indexed',
        args: { notIndexed: false },
    },
    NeedsIndex: {
        label: 'Needs index',
        value: 'needs-index',
        tooltip: 'Show only repositories that need to be indexed',
        args: { indexed: false },
    },
    FailedFetchOrClone: {
        label: 'Failed fetch/clone',
        value: 'failed-fetch',
        tooltip: 'Show only repositories that have failed to fetch or clone',
        args: { failedFetch: true },
    },
    Corrupted: {
        label: 'Corrupted',
        value: 'corrupted',
        tooltip: 'Show only repositories which are corrupt',
        args: { corrupted: true },
    },
}

const FILTERS: Filter<'orderBy' | 'status' | 'codeHost'>[] = [
    {
        id: 'orderBy',
        label: 'Order',
        type: 'select',
        options: [
            {
                label: 'Name (A-Z)',
                value: 'name-asc',
                tooltip: 'Order repositories by name in ascending order',
                args: {
                    orderBy: RepositoryOrderBy.REPOSITORY_NAME,
                    descending: false,
                },
            },
            {
                label: 'Name (Z-A)',
                value: 'name-desc',
                tooltip: 'Order repositories by name in descending order',
                args: {
                    orderBy: RepositoryOrderBy.REPOSITORY_NAME,
                    descending: true,
                },
            },
            {
                label: 'Size (largest first)',
                value: 'size-desc',
                tooltip: 'Order repositories by size in descending order',
                args: {
                    orderBy: RepositoryOrderBy.SIZE,
                    descending: true,
                },
            },
            {
                label: 'Size (smallest first)',
                value: 'size-asc',
                tooltip: 'Order repositories by size in ascending order',
                args: {
                    orderBy: RepositoryOrderBy.SIZE,
                    descending: false,
                },
            },
        ],
    },
    {
        id: 'status',
        label: 'Status',
        type: 'select',
        options: Object.values(STATUS_FILTERS),
    },
]

export const SiteAdminRepositoriesContainer: React.FunctionComponent<{ alwaysPoll?: boolean }> = ({
    alwaysPoll = false,
}) => {
    const {
        data,
        loading: repoStatsLoading,
        error: repoStatsError,
        startPolling,
        stopPolling,
    } = useQuery<StatusAndRepoStatsResult>(STATUS_AND_REPO_STATS, {})
    const location = useLocation()

    useEffect(() => {
        if (alwaysPoll || data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [alwaysPoll, data, startPolling, stopPolling])

    const {
        loading: extSvcLoading,
        data: extSvcs,
        error: extSvcError,
    } = useQuery<ExternalServiceIDsAndNamesResult, ExternalServiceIDsAndNamesVariables>(
        EXTERNAL_SERVICE_IDS_AND_NAMES,
        {}
    )

    const filters = useMemo<Filter<'orderBy' | 'status' | 'codeHost'>[]>(
        () => [
            ...FILTERS,
            ...(extSvcs && location.pathname !== PageRoutes.SetupWizard
                ? [
                      {
                          id: 'codeHost' as const,
                          label: 'Code Host',
                          type: 'select' as const,
                          options: [
                              {
                                  label: 'All',
                                  value: 'all',
                                  tooltip: 'Show all repositories',
                                  args: {},
                              },
                              ...extSvcs.externalServices.nodes.map(extSvc => ({
                                  label: extSvc.displayName,
                                  value: extSvc.id,
                                  tooltip: `Show all repositories discovered on ${extSvc.displayName}`,
                                  args: { externalService: extSvc.id },
                              })),
                          ],
                      },
                  ]
                : []),
        ],
        [extSvcs, location.pathname]
    )

    const [connectionState, setConnectionState] = useUrlSearchParamsForConnectionState(filters)
    const debouncedQuery = useDebounce(connectionState.query, 300)
    const {
        connection,
        loading: reposLoading,
        error: reposError,
        refetch,
        ...paginationProps
    } = usePageSwitcherPagination<
        RepositoriesResult,
        RepositoriesVariables,
        SiteAdminRepositoryFields,
        typeof connectionState
    >({
        query: REPOSITORIES_QUERY,
        variables: {
            ...buildFilterArgs(filters, connectionState),
            query: debouncedQuery,
        },
        getConnection: ({ data }) => data?.repositories || undefined,
        options: { pollInterval: 5000 },
        state: [connectionState, setConnectionState],
    })

    const error = repoStatsError || extSvcError || reposError
    const loading = repoStatsLoading || extSvcLoading || reposLoading
    const debouncedLoading = useDebounce(loading, 300)

    const legends = useMemo((): ValueLegendListProps['items'] | undefined => {
        if (!data) {
            return undefined
        }
        const items: ValueLegendListProps['items'] = [
            {
                value: data.repositoryStats.total,
                description: 'Repositories',
                color: 'var(--purple)',
                tooltip:
                    'Total number of repositories in the Sourcegraph instance. This number might be higher than the total number of repositories in the list below in case repository permissions do not allow you to view some repositories.',
            },
            {
                value: data.repositoryStats.notCloned,
                description: 'Queued',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that are queued to be cloned.',
                onClick: () => setConnectionState(prev => ({ ...prev, status: STATUS_FILTERS.NotCloned.value })),
            },
            {
                value: data.repositoryStats.cloning,
                description: 'Cloning',
                color: data.repositoryStats.cloning > 0 ? 'var(--primary)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that are currently being cloned.',
                onClick: () => setConnectionState(prev => ({ ...prev, status: STATUS_FILTERS.Cloning.value })),
            },
            {
                value: data.repositoryStats.cloned,
                description: 'Cloned',
                color: 'var(--success)',
                position: 'right',
                tooltip: 'The number of repositories that have been cloned.',
                onClick: () => setConnectionState(prev => ({ ...prev, status: STATUS_FILTERS.Cloned.value })),
            },
            {
                value: data.repositoryStats.indexed,
                description: 'Indexed',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that have been indexed for search.',
                onClick: () => setConnectionState(prev => ({ ...prev, status: STATUS_FILTERS.Indexed.value })),
            },
            {
                value: data.repositoryStats.failedFetch,
                description: 'Failed',
                color: data.repositoryStats.failedFetch > 0 ? 'var(--warning)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories where the last syncing attempt produced an error.',
                onClick: () =>
                    setConnectionState(prev => ({
                        ...prev,
                        status: STATUS_FILTERS.FailedFetchOrClone.value,
                    })),
            },
        ]

        if (data.repositoryStats.corrupted > 0) {
            items.push({
                value: data.repositoryStats.corrupted,
                description: 'Corrupted',
                color: 'var(--danger)',
                position: 'right',
                tooltip:
                    'The number of repositories where corruption has been detected. Reclone these repositories to get rid of corruption.',
                onClick: () => setConnectionState(prev => ({ ...prev, status: STATUS_FILTERS.Corrupted.value })),
            })
        }
        return items
    }, [setConnectionState, data])

    return (
        <>
            <Container className="py-3 mb-1">
                {error && !loading && <ErrorAlert error={error} />}
                {legends && <ValueLegendList items={legends} className={styles.legend} />}
            </Container>
            {extSvcs && (
                <Container>
                    <div className="d-flex flex-sm-row flex-column-reverse justify-content-center">
                        <FilterControl
                            filters={filters}
                            values={connectionState}
                            onValueSelect={(filter, value) =>
                                setConnectionState(prev => ({ ...prev, [filter.id]: value }))
                            }
                        />
                        <Input
                            type="search"
                            className="flex-1 md-ml-5 mb-1"
                            placeholder="Search repositories..."
                            name="query"
                            value={connectionState.query}
                            onChange={event =>
                                setConnectionState(prev => ({
                                    ...prev,
                                    query: event.currentTarget.value,
                                }))
                            }
                            autoComplete="off"
                            autoCorrect="off"
                            autoCapitalize="off"
                            spellCheck={false}
                            aria-label="Search repositories..."
                            variant="regular"
                        />
                    </div>
                    {debouncedLoading && !error && (
                        <div className="d-flex justify-content-center align-items-center ">
                            <LoadingSpinner />
                        </div>
                    )}
                    <ul className="list-group list-group-flush mt-4">
                        {(connection?.nodes || []).map(node => (
                            <RepositoryNode key={node.id} node={node} refetchAllRepos={refetch} />
                        ))}
                    </ul>
                    <PageSwitcher
                        {...paginationProps}
                        className="mt-4"
                        totalCount={connection?.totalCount ?? null}
                        totalLabel="repositories"
                    />
                </Container>
            )}
        </>
    )
}
