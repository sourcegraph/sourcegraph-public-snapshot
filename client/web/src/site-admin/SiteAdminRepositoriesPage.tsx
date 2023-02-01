import React, { useState, useEffect, useMemo } from 'react'

import { mdiCloudDownload, mdiCog, mdiBrain } from '@mdi/js'
import { isEqual } from 'lodash'

import { logger } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Alert,
    Button,
    Code,
    Container,
    H4,
    Icon,
    Input,
    Link,
    LoadingSpinner,
    PageHeader,
    Text,
    Tooltip,
    ErrorAlert,
    LinkOrSpan,
    PageSwitcher,
} from '@sourcegraph/wildcard'

import { EXTERNAL_SERVICE_IDS_AND_NAMES } from '../components/externalServices/backend'
import {
    buildFilterArgs,
    FilterControl,
    FilteredConnectionFilterValue,
    FilteredConnectionFilter,
} from '../components/FilteredConnection'
import { usePageSwitcherPagination } from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { getFilterFromURL, getUrlQuery } from '../components/FilteredConnection/utils'
import { PageTitle } from '../components/PageTitle'
import {
    RepositoriesResult,
    RepositoriesVariables,
    RepositoryOrderBy,
    RepositoryStatsResult,
    ExternalServiceIDsAndNamesVariables,
    ExternalServiceIDsAndNamesResult,
    RepositoryStatsVariables,
    SiteAdminRepositoryFields,
} from '../graphql-operations'
import { refreshSiteFlags } from '../site/backend'

import { ValueLegendList, ValueLegendListProps } from './analytics/components/ValueLegendList'
import { REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL, REPOSITORIES_QUERY } from './backend'
import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './SiteAdminRepositoriesPage.module.scss'
import { useLocation, useNavigate } from 'react-router-dom-v5-compat'

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => (
    <li
        className="repository-node list-group-item px-0 py-2"
        data-test-repository={node.name}
        data-test-cloned={node.mirrorInfo.cloned}
    >
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <ExternalRepositoryIcon externalRepo={node.externalRepository} />
                <RepoLink repoName={node.name} to={node.url} />
                <RepoMirrorInfo mirrorInfo={node.mirrorInfo} />
            </div>

            <div className="repository-node__actions">
                {!node.mirrorInfo.cloneInProgress && !node.mirrorInfo.cloned && (
                    <Button to={node.url} variant="secondary" size="sm" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiCloudDownload} /> Clone now
                    </Button>
                )}{' '}
                <Tooltip content="Repository code graph data">
                    <Button to={`/${node.name}/-/code-graph`} variant="secondary" size="sm" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiBrain} /> Code graph data
                    </Button>
                </Tooltip>{' '}
                <Tooltip content="Repository settings">
                    <Button to={`/${node.name}/-/settings`} variant="secondary" size="sm" as={Link}>
                        <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                    </Button>
                </Tooltip>
            </div>
        </div>

        {node.mirrorInfo.lastError && (
            <div className={styles.alertWrapper}>
                <Alert variant="warning">
                    <Text className="font-weight-bold">Error syncing repository:</Text>
                    <Code className={styles.alertContent}>{node.mirrorInfo.lastError.replaceAll('\r', '\n')}</Code>
                </Alert>
            </div>
        )}
        {node.mirrorInfo.isCorrupted && (
            <div className={styles.alertWrapper}>
                <Alert variant="danger">
                    Repository is corrupt. <LinkOrSpan to={`/${node.name}/-/settings/mirror`}>More details</LinkOrSpan>
                </Alert>
            </div>
        )}
    </li>
)

interface Props extends TelemetryProps {}

const STATUS_FILTERS: { [label: string]: FilteredConnectionFilterValue } = {
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

const FILTERS: FilteredConnectionFilter[] = [
    {
        id: 'order',
        label: 'Order',
        type: 'select',
        values: [
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
        values: Object.values(STATUS_FILTERS),
    },
]

/**
 * A page displaying the repositories on this site.
 */
export const SiteAdminRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
}) => {
    const location = useLocation()
    const navigate = useNavigate()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminRepos')
    }, [telemetryService])

    // Refresh global alert about enabling repositories when the user visits & navigates away from this page.
    useEffect(() => {
        refreshSiteFlags()
            .toPromise()
            .then(null, error => logger.error(error))
        return () => {
            refreshSiteFlags()
                .toPromise()
                .then(null, error => logger.error(error))
        }
    }, [])

    const {
        data,
        loading: repoStatsLoading,
        error: repoStatsError,
        startPolling,
        stopPolling,
    } = useQuery<RepositoryStatsResult, RepositoryStatsVariables>(REPOSITORY_STATS, {})

    useEffect(() => {
        if (data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [data, startPolling, stopPolling])

    const {
        loading: extSvcLoading,
        data: extSvcs,
        error: extSvcError,
    } = useQuery<ExternalServiceIDsAndNamesResult, ExternalServiceIDsAndNamesVariables>(
        EXTERNAL_SERVICE_IDS_AND_NAMES,
        {}
    )

    const filters = useMemo(() => {
        if (!extSvcs) {
            return FILTERS
        }

        const values = [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all repositories',
                args: {},
            },
        ]

        for (const extSvc of extSvcs.externalServices.nodes) {
            values.push({
                label: extSvc.displayName,
                value: extSvc.id,
                tooltip: `Show all repositories discovered on ${extSvc.displayName}`,
                args: { externalService: extSvc.id },
            })
        }

        const filtersWithExternalServices = FILTERS.slice() // use slice to copy array
        filtersWithExternalServices.push({
            id: 'codeHost',
            label: 'Code Host',
            type: 'select',
            values,
        })
        return filtersWithExternalServices
    }, [extSvcs])

    const [filterValues, setFilterValues] = useState<Map<string, FilteredConnectionFilterValue>>(() =>
        getFilterFromURL(new URLSearchParams(location.search), filters)
    )

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
                description: 'Not cloned',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that have not been cloned yet.',
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.NotCloned)
                        return newValues
                    }),
            },
            {
                value: data.repositoryStats.cloning,
                description: 'Cloning',
                color: data.repositoryStats.cloning > 0 ? 'var(--success)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that are currently being cloned.',
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.Cloning)
                        return newValues
                    }),
            },
            {
                value: data.repositoryStats.cloned,
                description: 'Cloned',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that have been cloned.',
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.Cloned)
                        return newValues
                    }),
            },
            {
                value: data.repositoryStats.indexed,
                description: 'Indexed',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that have been indexed for search.',
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.Indexed)
                        return newValues
                    }),
            },
            {
                value: data.repositoryStats.failedFetch,
                description: 'Failed',
                color: data.repositoryStats.failedFetch > 0 ? 'var(--warning)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories where the last syncing attempt produced an error.',
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.FailedFetchOrClone)
                        return newValues
                    }),
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
                onClick: () =>
                    setFilterValues(values => {
                        const newValues = new Map(values)
                        newValues.set('status', STATUS_FILTERS.Corrupted)
                        return newValues
                    }),
            })
        }
        return items
    }, [data, setFilterValues])

    const [searchQuery, setSearchQuery] = useState<string>(
        () => new URLSearchParams(location.search).get('query') || ''
    )

    useEffect(() => {
        const searchFragment = getUrlQuery({
            query: searchQuery,
            filters,
            filterValues,
            search: location.search,
        })
        const searchFragmentParams = new URLSearchParams(searchFragment)
        searchFragmentParams.sort()

        const oldParams = new URLSearchParams(location.search)
        oldParams.sort()

        if (!isEqual(Array.from(searchFragmentParams), Array.from(oldParams))) {
            navigate(
                {
                    search: searchFragment,
                    hash: location.hash,
                },
                {
                    replace: true,
                    // Do not throw away flash messages
                    state: location.state,
                }
            )
        }
    }, [filters, filterValues, searchQuery, location, navigate])

    const variables = useMemo<RepositoriesVariables>(() => {
        const args = buildFilterArgs(filterValues)

        return {
            ...args,
            query: searchQuery,
            indexed: args.indexed ?? true,
            notIndexed: args.notIndexed ?? true,
            failedFetch: args.failedFetch ?? false,
            corrupted: args.corrupted ?? false,
            cloneStatus: args.cloneStatus ?? null,
            externalService: args.externalService ?? null,
        } as RepositoriesVariables
    }, [searchQuery, filterValues])

    const {
        connection,
        loading: reposLoading,
        error: reposError,
        refetch,
        ...paginationProps
    } = usePageSwitcherPagination<RepositoriesResult, RepositoriesVariables, SiteAdminRepositoryFields>({
        query: REPOSITORIES_QUERY,
        variables,
        getConnection: ({ data }) => data?.repositories || undefined,
        options: { pollInterval: 5000 },
    })

    useEffect(() => {
        refetch(variables)
    }, [refetch, variables])

    const showRepositoriesAddedBanner = new URLSearchParams(location.search).has('repositoriesUpdated')

    const licenseInfo = window.context.licenseInfo

    const error = repoStatsError || extSvcError || reposError
    const loading = repoStatsLoading || extSvcLoading || reposLoading

    return (
        <div className="site-admin-repositories-page">
            <PageTitle title="Repositories - Admin" />
            {showRepositoriesAddedBanner && (
                <Alert variant="success" as="p">
                    Syncing repositories. It may take a few moments to clone and index each repository. Repository
                    statuses are displayed below.
                </Alert>
            )}
            <PageHeader
                path={[{ text: 'Repositories' }]}
                headingElement="h2"
                description={
                    <>
                        Repositories are synced from connected{' '}
                        <Link
                            to="/site-admin/external-services"
                            data-testid="test-repositories-code-host-connections-link"
                        >
                            code hosts
                        </Link>
                        .
                    </>
                }
                className="mb-3"
            />
            {licenseInfo && (licenseInfo.codeScaleCloseToLimit || licenseInfo.codeScaleExceededLimit) && (
                <Alert variant={licenseInfo.codeScaleExceededLimit ? 'danger' : 'warning'}>
                    <H4>
                        {licenseInfo.codeScaleExceededLimit ? (
                            <>You've used all 100GiB of storage</>
                        ) : (
                            <>Your Sourcegraph is almost full</>
                        )}
                    </H4>
                    {licenseInfo.codeScaleExceededLimit ? <>You're about to reach the 100GiB storage limit. </> : <></>}
                    Upgrade to <Link to="https://about.sourcegraph.com/pricing">Sourcegraph Enterprise</Link> for
                    unlimited storage for your code.
                </Alert>
            )}

            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {extSvcs && (
                    <>
                        <div className="d-flex justify-content-center">
                            <FilterControl
                                filters={filters}
                                values={filterValues}
                                onValueSelect={(
                                    filter: FilteredConnectionFilter,
                                    value: FilteredConnectionFilterValue
                                ) =>
                                    setFilterValues(values => {
                                        const newValues = new Map(values)
                                        newValues.set(filter.id, value)
                                        return newValues
                                    })
                                }
                            />
                            <Input
                                type="search"
                                className="flex-1"
                                placeholder="Search repositories..."
                                name="query"
                                value={searchQuery}
                                onChange={event => setSearchQuery(event.currentTarget.value)}
                                autoComplete="off"
                                autoCorrect="off"
                                autoCapitalize="off"
                                spellCheck={false}
                                aria-label="Search repositories..."
                                variant="regular"
                            />
                        </div>
                        <ul className="list-group list-group-flush mt-4">
                            {(connection?.nodes || []).map(node => (
                                <RepositoryNode key={node.id} node={node} />
                            ))}
                        </ul>
                        <PageSwitcher
                            {...paginationProps}
                            className="mt-4"
                            totalCount={connection?.totalCount ?? null}
                            totalLabel="repositories"
                        />
                    </>
                )}
            </Container>
        </div>
    )
}
