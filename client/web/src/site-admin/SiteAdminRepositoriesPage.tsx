import React, { useEffect, useCallback, useMemo } from 'react'

import { mdiCloudDownload, mdiCog } from '@mdi/js'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'

import { useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, Alert, Icon, H2, Text, Tooltip, Container, LoadingSpinner } from '@sourcegraph/wildcard'

import { TerminalLine } from '../auth/Terminal'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import {
    RepositoriesResult,
    RepositoryStatsResult,
    RepositoryStatsVariables,
    SiteAdminRepositoryFields,
} from '../graphql-operations'
import { refreshSiteFlags } from '../site/backend'

import { ValueLegendList, ValueLegendListProps } from './analytics/components/ValueLegendList'
import { fetchAllRepositoriesAndPollIfEmptyOrAnyCloning, REPOSITORY_STATS, REPO_PAGE_POLL_INTERVAL } from './backend'
import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo as RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './SiteAdminRepositoriesPage.module.scss'

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => (
    <li
        className="repository-node list-group-item py-2"
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
                    <TerminalLine className={styles.alertContent}>
                        {node.mirrorInfo.lastError.replaceAll('\r', '\n')}
                    </TerminalLine>
                </Alert>
            </div>
        )}
    </li>
)

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

const FILTERS: FilteredConnectionFilter[] = [
    {
        id: 'status',
        label: 'Status',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all repositories',
                args: {},
            },
            {
                label: 'Cloned',
                value: 'cloned',
                tooltip: 'Show cloned repositories only',
                args: { cloneStatus: 'CLONED' },
            },
            {
                label: 'Cloning',
                value: 'cloning',
                tooltip: 'Show repositories currently being cloned only',
                args: { cloneStatus: 'CLONING' },
            },
            {
                label: 'Not cloned',
                value: 'not-cloned',
                tooltip: 'Show only repositories that have not been cloned yet',
                args: { cloneStatus: 'NOT_CLONED' },
            },
            {
                label: 'Needs index',
                value: 'needs-index',
                tooltip: 'Show only repositories that need to be indexed',
                args: { indexed: false },
            },
            {
                label: 'Failed fetch/clone',
                value: 'failed-fetch',
                tooltip: 'Show only repositories that have failed to fetch or clone',
                args: { failedFetch: true },
            },
        ],
    },
]

/**
 * A page displaying the repositories on this site.
 */
export const SiteAdminRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    history,
    location,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminRepos')
    }, [telemetryService])

    // Refresh global alert about enabling repositories when the user visits & navigates away from this page.
    useEffect(() => {
        refreshSiteFlags()
            .toPromise()
            .then(null, error => console.error(error))
        return () => {
            refreshSiteFlags()
                .toPromise()
                .then(null, error => console.error(error))
        }
    }, [])

    const { data, loading, error, startPolling, stopPolling } = useQuery<
        RepositoryStatsResult,
        RepositoryStatsVariables
    >(REPOSITORY_STATS, {})

    useEffect(() => {
        if (data?.repositoryStats?.total === 0 || data?.repositoryStats?.cloning !== 0) {
            startPolling(REPO_PAGE_POLL_INTERVAL)
        } else {
            stopPolling()
        }
    }, [data, startPolling, stopPolling])

    const legends = useMemo((): ValueLegendListProps['items'] | undefined => {
        if (!data) {
            return undefined
        }
        return [
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
                tooltip: 'The number of repositories that haven not been cloned yet.',
            },
            {
                value: data.repositoryStats.cloning,
                description: 'Cloning',
                color: data.repositoryStats.cloning > 0 ? 'var(--success)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that are currently being cloned.',
            },
            {
                value: data.repositoryStats.cloned,
                description: 'Cloned',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories that have been cloned.',
            },
            {
                value: data.repositoryStats.failedFetch,
                description: 'Failed',
                color: data.repositoryStats.failedFetch > 0 ? 'var(--warning)' : 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of repositories where the last syncing attempt produced an error.',
            },
        ]
    }, [data])

    const queryRepositories = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<RepositoriesResult['repositories']> =>
            fetchAllRepositoriesAndPollIfEmptyOrAnyCloning(args),
        []
    )
    const showRepositoriesAddedBanner = new URLSearchParams(location.search).has('repositoriesUpdated')

    return (
        <div className="site-admin-repositories-page">
            <PageTitle title="Repositories - Admin" />
            {showRepositoriesAddedBanner && (
                <Alert variant="success" as="p">
                    Syncing repositories. It may take a few moments to clone and index each repository. Repository
                    statuses are displayed below.
                </Alert>
            )}
            <H2>Repositories</H2>
            <Text>
                Repositories are synced from connected{' '}
                <Link to="/site-admin/external-services" data-testid="test-repositories-code-host-connections-link">
                    code hosts
                </Link>
                .
            </Text>
            <Container className="mb-3">
                {error && !loading && (
                    <Alert variant="warning" as="p">
                        {error.message}
                    </Alert>
                )}
                {loading && !error && <LoadingSpinner />}
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                <FilteredConnection<SiteAdminRepositoryFields, Omit<RepositoryNodeProps, 'node'>>
                    className="mb-0"
                    listClassName="list-group list-group-flush mt-3"
                    noun="repository"
                    pluralNoun="repositories"
                    queryConnection={queryRepositories}
                    nodeComponent={RepositoryNode}
                    inputClassName="flex-1"
                    filters={FILTERS}
                    history={history}
                    location={location}
                />
            </Container>
        </div>
    )
}
